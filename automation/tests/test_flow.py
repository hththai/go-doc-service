"""End-to-end flow test with a fake runner (no LLM, no API key, no Go toolchain)."""

from pathlib import Path

from dev_crew.crews import DocsResult, Plan, PlanStep, ReviewResult
from dev_crew.main import DevFlow, run
from dev_crew.settings import Check, Settings

TARGET = "server/internal/demo/service.go"


class FakeRunner:
    """Attempt 1 breaks the 'build', attempt 2 fixes it, review rejects once then approves."""

    def __init__(self, settings, sandbox):
        self.root = sandbox.root
        self.calls = []
        self.reviews = 0

    def plan(self, brief, repo_context):
        assert "# rules" in repo_context  # CLAUDE.md made it into the context
        self.calls.append("plan")
        return Plan(summary="flip Ok", touches_client=False,
                    steps=[PlanStep(layer="service", file=TARGET, change="return false")],
                    tests=["demo_test.go"])

    def implement(self, brief, plan, feedback):
        n = self.calls.count("implement") + 1
        self.calls.append("implement")
        body = "BROKEN" if n == 1 else f"func Ok() bool {{ return false }} // attempt {n}"
        (self.root / TARGET).write_text(f"package demo\n\n{body}\n")
        if n > 1:
            assert feedback  # previous failure / review issues were passed back
        return f"attempt {n}"

    def review(self, brief, plan, diff, check_summary):
        self.calls.append("review")
        self.reviews += 1
        if self.reviews == 1:
            return ReviewResult(approved=False, issues=["add a test"], summary="needs test")
        return ReviewResult(approved=True, summary="looks good")

    def docs(self, brief, plan, diff):
        self.calls.append("docs")
        return DocsResult(updated_files=[], commit_message="fix(demo): flip Ok\n\n- example")


def _settings(repo: Path, tmp_path: Path, max_attempts: int) -> Settings:
    return Settings(
        repo_root=repo, max_attempts=max_attempts,
        worktree_root=tmp_path / "wts", runs_dir=tmp_path / "runs",
        checks=[Check("fake build", "server", "server", "! grep -rq BROKEN .")],
    )


def _run(repo, tmp_path, max_attempts):
    brief = tmp_path / "flip-ok.md"
    brief.write_text("# Flip Ok\n\nMake Ok return false.\n")
    holder = {}

    def factory(s, sb):
        holder["r"] = FakeRunner(s, sb)
        return holder["r"]

    flow = DevFlow.build(_settings(repo, tmp_path, max_attempts), factory)
    run(flow, str(brief))
    return flow, holder["r"]


def test_flow_retries_until_ready(repo, tmp_path):
    flow, runner = _run(repo, tmp_path, max_attempts=4)
    st = flow.state
    # attempt1: checks fail -> attempt2: pass, review rejects -> attempt3: pass, approved
    assert runner.calls == ["plan", "implement", "implement", "review", "implement", "review", "docs"]
    assert st.outcome == "ready" and st.attempt == 3
    run_dir = Path(st.run_dir)
    assert (run_dir / "COMMIT_MSG.txt").read_text().startswith("fix(demo)")
    report = (run_dir / "report.md").read_text()
    assert "git commit -F" in report and st.branch in report
    # main checkout untouched, work lives on the crew branch
    assert "return true" in (repo / TARGET).read_text()


def test_flow_gives_up(repo, tmp_path):
    flow, runner = _run(repo, tmp_path, max_attempts=1)
    assert flow.state.outcome == "failed"
    assert runner.calls == ["plan", "implement"]
    assert "worktree remove --force" in (Path(flow.state.run_dir) / "report.md").read_text()


def test_llm_sends_no_temperature_by_default(monkeypatch):
    from dev_crew.crews import llm_options

    monkeypatch.delenv("TEMPERATURE", raising=False)
    assert llm_options() == {}
    monkeypatch.setenv("TEMPERATURE", "0.2")
    assert llm_options() == {"temperature": 0.2}
