"""Deterministic verification: the same commands lefthook and Jenkins run."""

from __future__ import annotations

import subprocess
import time
from dataclasses import dataclass, field
from pathlib import Path

from .gitops import changed_paths
from .settings import AREA_PREFIXES, Check


@dataclass
class CheckResult:
    name: str
    cmd: str
    passed: bool
    output: str
    seconds: float


@dataclass
class CheckReport:
    areas: list[str]
    results: list[CheckResult] = field(default_factory=list)

    @property
    def passed(self) -> bool:
        return bool(self.results) and all(r.passed for r in self.results)

    def summary(self) -> str:
        if not self.results:
            return "No checks ran (no changed area)."
        lines = [f"{'PASS' if r.passed else 'FAIL'}  {r.name}  ({r.cmd}, {r.seconds:.1f}s)"
                 for r in self.results]
        return "\n".join(lines)

    def failures(self, max_chars_each: int = 6000) -> str:
        blocks = []
        for r in self.results:
            if r.passed:
                continue
            out = r.output.strip() or "(no output)"
            if len(out) > max_chars_each:
                out = "...[earlier output trimmed]\n" + out[-max_chars_each:]
            blocks.append(f"### {r.name} failed: `{r.cmd}`\n```\n{out}\n```")
        return "\n\n".join(blocks)

    def full_text(self) -> str:
        body = "\n\n".join(f"### {r.name} ({'PASS' if r.passed else 'FAIL'})\n$ {r.cmd}\n{r.output}"
                           for r in self.results)
        return f"Areas: {', '.join(self.areas) or '-'}\n\n{self.summary()}\n\n{body}"


def areas_for(paths: list[str]) -> list[str]:
    areas = set()
    for path in paths:
        for area, prefixes in AREA_PREFIXES.items():
            if path.startswith(prefixes):
                areas.add(area)
    return sorted(areas)


def run_checks(
    worktree: Path,
    checks: list[Check],
    areas: list[str] | None = None,
    timeout: int = 900,
) -> CheckReport:
    if areas is None:
        areas = areas_for(changed_paths(worktree))
    report = CheckReport(areas=areas)
    for check in checks:
        if check.area not in areas:
            continue
        cwd = worktree / check.cwd
        start = time.monotonic()
        if not cwd.is_dir():
            report.results.append(CheckResult(check.name, check.cmd, False, f"missing folder {check.cwd}", 0))
            continue
        try:
            proc = subprocess.run(
                check.cmd, shell=True, cwd=cwd, capture_output=True, text=True, timeout=timeout
            )
            output = (proc.stdout + ("\n" + proc.stderr if proc.stderr else "")).strip()
            passed = proc.returncode == 0 and not (check.fail_on_output and proc.stdout.strip())
            if check.fail_on_output and proc.stdout.strip():
                output = f"These files need `{check.cmd.split()[0]} -w`:\n{output}"
        except subprocess.TimeoutExpired:
            passed, output = False, f"timed out after {timeout}s"
        report.results.append(
            CheckResult(check.name, check.cmd, passed, output, time.monotonic() - start)
        )
    return report
