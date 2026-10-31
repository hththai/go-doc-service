"""DevFlow: brief -> plan -> implement -> verify -> review -> docs -> suggested commit.

    uv run dev-crew briefs/maintenance-cleanup.md
"""

from __future__ import annotations

import argparse
import json
import sys
from datetime import datetime
from pathlib import Path
from typing import Callable

from opentelemetry import trace
from crewai.flow.flow import Flow, listen, or_, router, start
from dotenv import load_dotenv
from pydantic import BaseModel, PrivateAttr

from . import routing
from .checks import run_checks
from .context import collect_context
from .crews import CrewRunner, Plan, StageRunner
from .gitops import changed_paths, create_worktree, diff_stat, full_diff, slugify
from .report import write_report
from .sandbox import Sandbox
from .settings import AUTOMATION_DIR, Settings
from .tracing import set_output, setup_tracing, shutdown_tracing, stage

RunnerFactory = Callable[[Settings, Sandbox], StageRunner]


class DevState(BaseModel):
    brief_path: str = ""
    brief: str = ""
    title: str = ""
    branch: str = ""
    worktree: str = ""
    run_dir: str = ""
    plan: dict = {}
    attempt: int = 0
    max_attempts: int = 3
    feedback: str = ""
    check_summary: str = ""
    review: dict = {}
    implement_notes: list[str] = []
    docs_files: list[str] = []
    commit_message: str = ""
    outcome: str = ""  # "ready" | "failed"


class DevFlow(Flow[DevState]):
    _settings: Settings = PrivateAttr()
    _factory: RunnerFactory = PrivateAttr()
    _runner: StageRunner | None = PrivateAttr(default=None)

    @classmethod
    def build(cls, settings: Settings, factory: RunnerFactory | None = None) -> "DevFlow":
        flow = cls()
        flow._settings = settings
        flow._factory = factory or (lambda s, sb: CrewRunner(s, sb))
        return flow

    # helpers
    @property
    def wt(self) -> Path:
        return Path(self.state.worktree)

    def _save(self, name: str, content: str) -> None:
        (Path(self.state.run_dir) / name).write_text(content, encoding="utf-8")

    def _plan(self) -> Plan:
        return Plan.model_validate(self.state.plan)

    # ---- steps ----

    @start()
    def prepare(self):
        s = self._settings
        brief_file = Path(self.state.brief_path).expanduser().resolve()
        self.state.brief = brief_file.read_text(encoding="utf-8")
        first = next((l for l in self.state.brief.splitlines() if l.strip()), brief_file.stem)
        self.state.title = first.lstrip("# ").strip()
        stamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        slug = slugify(brief_file.stem)
        self.state.branch = f"crew/{slug}-{stamp}"
        self.state.max_attempts = s.max_attempts
        self.state.worktree = str(create_worktree(s.repo_root, s.worktree_root, self.state.branch, s.base_ref))
        run_dir = s.runs_dir / f"{stamp}-{slug}"
        run_dir.mkdir(parents=True, exist_ok=True)
        self.state.run_dir = str(run_dir)
        self._runner = self._factory(s, Sandbox(self.wt))
        trace.get_current_span().set_attribute("dev_crew.branch", self.state.branch)
        print(f"[dev-crew] branch {self.state.branch}\n[dev-crew] worktree {self.wt}")

    @listen(prepare)
    def make_plan(self):
        with stage("plan", input=self.state.brief) as span:
            plan = self._runner.plan(self.state.brief, collect_context(self.wt))
            self.state.plan = plan.model_dump()
            self._save("plan.json", plan.model_dump_json(indent=2))
            set_output(span, self.state.plan)
        print(f"[dev-crew] plan: {plan.summary} ({len(plan.steps)} steps)")

    @listen(or_(make_plan, routing.RETRY))
    def implement(self):
        self.state.attempt += 1
        n = self.state.attempt
        print(f"[dev-crew] implement attempt {n}/{self.state.max_attempts}")
        with stage(f"implement #{n}", input=self.state.feedback or "first attempt",
                   attempt=n, coder=self._settings.coder) as span:
            notes = self._runner.implement(self.state.brief, self._plan(), self.state.feedback)
            set_output(span, notes)
        self.state.implement_notes.append(notes)
        self._save(f"implement-{n}.md", notes)

    @router(implement)
    def verify(self):
        n = self.state.attempt
        with stage(f"verify #{n}", attempt=n) as span:
            changed = changed_paths(self.wt)
            span.set_attribute("dev_crew.changed_files", len(changed))
            if changed:
                report = run_checks(self.wt, self._settings.checks, timeout=self._settings.check_timeout)
                passed = report.passed
                self.state.check_summary = report.summary()
                self._save(f"checks-{n}.txt", report.full_text())
                self.state.feedback = "" if passed else "The checks failed:\n\n" + report.failures()
            else:
                passed = False
                self.state.check_summary = "No files changed."
                self.state.feedback = "You did not change any files. Implement the plan using the edit/write tools."
            decision = routing.after_verify(has_changes=bool(changed), checks_passed=passed,
                                            attempt=n, max_attempts=self.state.max_attempts)
            span.set_attribute("dev_crew.route", decision)
            set_output(span, self.state.check_summary + ("\n\n" + self.state.feedback if not passed else ""), ok=passed)
        print(f"[dev-crew] checks: {'passed' if passed else 'failed'}")
        return decision

    @router(routing.PASSED)
    def review(self):
        n = self.state.attempt
        with stage(f"review #{n}", input=self.state.check_summary, attempt=n) as span:
            result = self._runner.review(self.state.brief, self._plan(), full_diff(self.wt),
                                         self.state.check_summary)
            self.state.review = result.model_dump()
            self._save(f"review-{n}.json", result.model_dump_json(indent=2))
            if not result.approved:
                self.state.feedback = "The reviewer requested changes:\n" + "\n".join(f"- {i}" for i in result.issues)
            decision = routing.after_review(approved=result.approved, attempt=n,
                                            max_attempts=self.state.max_attempts)
            span.set_attribute("dev_crew.approved", result.approved)
            span.set_attribute("dev_crew.route", decision)
            set_output(span, self.state.review, ok=result.approved)
        print(f"[dev-crew] review: {'approved' if result.approved else 'changes requested'}")
        return decision

    @listen(routing.APPROVED)
    def finalize_docs(self):
        with stage("docs") as span:
            docs = self._runner.docs(self.state.brief, self._plan(), full_diff(self.wt))
            self.state.docs_files = docs.updated_files
            self.state.commit_message = docs.commit_message.strip()
            set_output(span, docs.model_dump())
        self._save("COMMIT_MSG.txt", self.state.commit_message + "\n")
        self.state.outcome = "ready"
        self._finish()

    @listen(routing.GIVE_UP)
    def stop_failed(self):
        self.state.outcome = "failed"
        self._finish()

    def _finish(self):
        path = write_report(self.state, diff_stat(self.wt))
        print(f"[dev-crew] {self.state.outcome.upper()} - report: {path}")


def run(flow: DevFlow, brief_path: str) -> DevState:
    """Kick off the flow inside one root span so Phoenix shows a single trace per run."""
    with stage("run", input=Path(brief_path).name, brief=brief_path) as root:
        flow.kickoff(inputs={"brief_path": brief_path})
        st = flow.state
        root.set_attribute("dev_crew.branch", st.branch)
        root.set_attribute("dev_crew.outcome", st.outcome)
        root.set_attribute("dev_crew.attempts", st.attempt)
        set_output(root, {"outcome": st.outcome, "branch": st.branch, "attempts": st.attempt,
                          "report": str(Path(st.run_dir) / "report.md") if st.run_dir else ""},
                   ok=st.outcome == "ready")
    return flow.state


def cli(argv: list[str] | None = None) -> int:
    load_dotenv(AUTOMATION_DIR / ".env")
    ap = argparse.ArgumentParser(prog="dev-crew", description=__doc__.splitlines()[0])
    ap.add_argument("brief", help="Markdown task brief (see briefs/)")
    ap.add_argument("--coder", choices=["crew", "claude-code"])
    ap.add_argument("--model")
    ap.add_argument("--max-attempts", type=int)
    ap.add_argument("--base", dest="base_ref", help="Base git ref for the new branch (default HEAD)")
    ap.add_argument("--verbose", action="store_true", default=None)
    args = ap.parse_args(argv)

    settings = Settings.from_env(coder=args.coder, model=args.model,
                                 max_attempts=args.max_attempts, base_ref=args.base_ref,
                                 verbose=args.verbose)
    ui = setup_tracing()
    if ui:
        print(f"[dev-crew] tracing to Phoenix, open {ui} (project: dev-crew)")
    try:
        state = run(DevFlow.build(settings), str(Path(args.brief).resolve()))
    finally:
        shutdown_tracing()
    return 0 if state.outcome == "ready" else 1


if __name__ == "__main__":
    sys.exit(cli())
