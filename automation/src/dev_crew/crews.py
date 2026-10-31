"""Builds the crewAI agents/tasks from YAML and runs each stage of the flow."""

from __future__ import annotations

import json
from datetime import date
from pathlib import Path
from typing import Protocol

import yaml
from crewai import LLM, Agent, Crew, Process, Task
from pydantic import BaseModel, Field

from .claude_code import run_claude_code
from .sandbox import Sandbox
from .settings import Settings
from .tools import read_tools, write_tools

CONFIG_DIR = Path(__file__).parent / "config"


# ---------- structured outputs ----------

class PlanStep(BaseModel):
    layer: str = Field(description="model | repository | service | handler | test | client | docs")
    file: str = Field(description="Path relative to the repo root")
    change: str = Field(description="What to change in this file, specifically")


class Plan(BaseModel):
    summary: str
    touches_client: bool = Field(description="True only if client-vite needs changes")
    steps: list[PlanStep]
    tests: list[str] = Field(description="Tests to add or update, with file paths")
    risks: list[str] = Field(default_factory=list)


class ReviewResult(BaseModel):
    approved: bool
    issues: list[str] = Field(default_factory=list, description="Concrete, file-specific problems")
    summary: str


class DocsResult(BaseModel):
    updated_files: list[str] = Field(default_factory=list)
    commit_message: str


# ---------- runner interface (the flow depends only on this) ----------

class StageRunner(Protocol):
    def plan(self, brief: str, repo_context: str) -> Plan: ...
    def implement(self, brief: str, plan: Plan, feedback: str) -> str: ...
    def review(self, brief: str, plan: Plan, diff: str, check_summary: str) -> ReviewResult: ...
    def docs(self, brief: str, plan: Plan, diff: str) -> DocsResult: ...


def _load(name: str) -> dict:
    return yaml.safe_load((CONFIG_DIR / name).read_text(encoding="utf-8"))


class CrewRunner:
    """Real implementation: one small crew per stage, all sharing one LLM."""

    def __init__(self, settings: Settings, sandbox: Sandbox):
        self.settings = settings
        self.sandbox = sandbox
        self.agents_cfg = _load("agents.yaml")
        self.tasks_cfg = _load("tasks.yaml")
        self.llm = LLM(model=settings.model, temperature=0.2)

    def _agent(self, name: str, tools: list) -> Agent:
        cfg = self.agents_cfg[name]
        return Agent(
            role=cfg["role"],
            goal=cfg["goal"],
            backstory=cfg["backstory"],
            tools=tools,
            llm=self.llm,
            allow_delegation=False,
            max_iter=cfg.get("max_iter", 25),
            verbose=self.settings.verbose,
        )

    def _task(self, name: str, agent: Agent, **kwargs) -> Task:
        cfg = self.tasks_cfg[name]
        return Task(description=cfg["description"], expected_output=cfg["expected_output"],
                    agent=agent, **kwargs)

    def _run(self, agents: list[Agent], tasks: list[Task], inputs: dict):
        crew = Crew(agents=agents, tasks=tasks, process=Process.sequential,
                    verbose=self.settings.verbose)
        return crew.kickoff(inputs=inputs)

    # --- stages ---

    def plan(self, brief: str, repo_context: str) -> Plan:
        planner = self._agent("planner", read_tools(self.sandbox))
        task = self._task("plan_task", planner, output_pydantic=Plan)
        out = self._run([planner], [task], {"brief": brief, "repo_context": repo_context})
        return out.pydantic if isinstance(out.pydantic, Plan) else Plan.model_validate_json(out.raw)

    def implement(self, brief: str, plan: Plan, feedback: str) -> str:
        plan_json = plan.model_dump_json(indent=2)
        if self.settings.coder == "claude-code":
            return run_claude_code(self.sandbox.root, brief, plan_json, feedback)
        tools = read_tools(self.sandbox) + write_tools(
            self.sandbox, self.settings.checks, self.settings.check_timeout)
        agents, tasks = [], []
        server_work = any(not s.file.startswith("client-vite/") for s in plan.steps)
        if server_work or not plan.touches_client:
            dev = self._agent("go_developer", tools)
            agents.append(dev)
            tasks.append(self._task("backend_task", dev))
        if plan.touches_client:
            fe = self._agent("frontend_developer", tools)
            agents.append(fe)
            tasks.append(self._task("frontend_task", fe, context=list(tasks)))
        out = self._run(agents, tasks, {
            "brief": brief, "plan": plan_json,
            "feedback": feedback or "None, this is the first attempt.",
        })
        return "\n\n".join(t.raw for t in out.tasks_output) if out.tasks_output else out.raw

    def review(self, brief: str, plan: Plan, diff: str, check_summary: str) -> ReviewResult:
        reviewer = self._agent("reviewer", read_tools(self.sandbox))
        task = self._task("review_task", reviewer, output_pydantic=ReviewResult)
        out = self._run([reviewer], [task], {
            "brief": brief, "plan": plan.model_dump_json(indent=2),
            "diff": diff, "check_summary": check_summary,
        })
        if isinstance(out.pydantic, ReviewResult):
            return out.pydantic
        return ReviewResult.model_validate_json(out.raw)

    def docs(self, brief: str, plan: Plan, diff: str) -> DocsResult:
        tools = read_tools(self.sandbox) + write_tools(
            self.sandbox, self.settings.checks, md_only=True)
        writer = self._agent("docs_writer", tools)
        task = self._task("docs_task", writer, output_pydantic=DocsResult)
        out = self._run([writer], [task], {
            "brief": brief + f"\n\n(Today is {date.today().isoformat()}.)",
            "plan": plan.model_dump_json(indent=2), "diff": diff,
        })
        if isinstance(out.pydantic, DocsResult):
            return out.pydantic
        return DocsResult.model_validate(json.loads(out.raw))
