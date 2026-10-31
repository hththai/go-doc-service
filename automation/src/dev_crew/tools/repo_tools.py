"""Sandboxed repo tools given to the agents. Every path goes through Sandbox."""

from __future__ import annotations

import fnmatch
import functools
import re
from pathlib import Path
from typing import Any, Type

from crewai.tools import BaseTool
from pydantic import BaseModel, ConfigDict, Field

from ..checks import run_checks
from ..gitops import full_diff
from ..sandbox import Sandbox, SandboxError
from ..settings import Check
from ..tracing import tool_span

MAX_READ_LINES = 800
MAX_FILE_BYTES = 1_000_000


class _RepoTool(BaseTool):
    model_config = ConfigDict(arbitrary_types_allowed=True)
    sandbox: Any = Field(exclude=True)

    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        if "_run" in cls.__dict__:  # wrap every tool's _run in a TOOL span for Phoenix
            inner = cls.__dict__["_run"]

            @functools.wraps(inner)
            def traced(self, *args, **kw):
                with tool_span(self.name, kw or args) as span:
                    result = inner(self, *args, **kw)
                    span.finish(result)
                    return result

            cls._run = traced

    @property
    def sb(self) -> Sandbox:
        return self.sandbox

    def _safe(self, fn, *args, **kwargs) -> str:
        try:
            return fn(*args, **kwargs)
        except SandboxError as exc:
            return f"ERROR: {exc}"
        except (OSError, UnicodeDecodeError, re.error) as exc:
            return f"ERROR: {type(exc).__name__}: {exc}"


# ---------- read-only tools ----------

class _ListArgs(BaseModel):
    path: str = Field(".", description="Folder relative to the repo root, e.g. 'server/internal/category'")
    max_entries: int = Field(300, description="Maximum number of files to list")


class ListFilesTool(_RepoTool):
    name: str = "list_files"
    description: str = "List files under a folder of the repository (recursive, protected folders hidden)."
    args_schema: Type[BaseModel] = _ListArgs

    def _run(self, path: str = ".", max_entries: int = 300) -> str:
        def go() -> str:
            start = self.sb.resolve(path)
            if not start.is_dir():
                return f"ERROR: '{path}' is not a folder"
            files = []
            for f in self.sb.iter_files(start):
                files.append(self.sb.rel(f))
                if len(files) >= max_entries:
                    files.append(f"... (stopped at {max_entries} entries)")
                    break
            return "\n".join(files) or "(empty)"
        return self._safe(go)


class _ReadArgs(BaseModel):
    path: str = Field(..., description="File path relative to the repo root")
    start_line: int = Field(1, description="First line to return (1-based)")
    end_line: int = Field(0, description="Last line to return; 0 means up to 800 lines from start")


class ReadFileTool(_RepoTool):
    name: str = "read_file"
    description: str = "Read a text file with line numbers. Use start_line/end_line for large files."
    args_schema: Type[BaseModel] = _ReadArgs

    def _run(self, path: str, start_line: int = 1, end_line: int = 0) -> str:
        def go() -> str:
            p = self.sb.resolve(path)
            if not p.is_file():
                return f"ERROR: '{path}' does not exist"
            lines = p.read_text(encoding="utf-8").splitlines()
            start = max(start_line, 1)
            end = end_line if end_line > 0 else start + MAX_READ_LINES - 1
            end = min(end, len(lines), start + MAX_READ_LINES - 1)
            body = "\n".join(f"{i:5d}  {lines[i - 1]}" for i in range(start, end + 1))
            more = f"\n... ({len(lines) - end} more lines)" if end < len(lines) else ""
            return f"{path} (lines {start}-{end} of {len(lines)})\n{body}{more}"
        return self._safe(go)


class _SearchArgs(BaseModel):
    pattern: str = Field(..., description="Python regular expression to search for")
    path: str = Field(".", description="Folder to search in, relative to the repo root")
    file_glob: str = Field("*", description="Filename filter, e.g. '*.go' or '*.tsx'")


class SearchCodeTool(_RepoTool):
    name: str = "search_code"
    description: str = "Search the code with a regex. Returns up to 100 'path:line: text' matches."
    args_schema: Type[BaseModel] = _SearchArgs

    def _run(self, pattern: str, path: str = ".", file_glob: str = "*") -> str:
        def go() -> str:
            rx = re.compile(pattern)
            start = self.sb.resolve(path)
            hits: list[str] = []
            for f in self.sb.iter_files(start):
                if not fnmatch.fnmatch(f.name, file_glob) or f.stat().st_size > MAX_FILE_BYTES:
                    continue
                try:
                    text = f.read_text(encoding="utf-8")
                except UnicodeDecodeError:
                    continue
                for n, line in enumerate(text.splitlines(), 1):
                    if rx.search(line):
                        hits.append(f"{self.sb.rel(f)}:{n}: {line.strip()[:200]}")
                        if len(hits) >= 100:
                            return "\n".join(hits) + "\n... (stopped at 100 matches)"
            return "\n".join(hits) or "(no matches)"
        return self._safe(go)


class _NoArgs(BaseModel):
    pass


class GitDiffTool(_RepoTool):
    name: str = "git_diff"
    description: str = "Show every change made so far in this worktree (including new files)."
    args_schema: Type[BaseModel] = _NoArgs

    def _run(self) -> str:
        return self._safe(lambda: full_diff(self.sb.root))


# ---------- write tools ----------

class _WriteArgs(BaseModel):
    path: str = Field(..., description="File path relative to the repo root")
    content: str = Field(..., description="Complete new file content")


class WriteFileTool(_RepoTool):
    name: str = "write_file"
    description: str = (
        "Create a new file or fully replace an existing one. Prefer edit_file for small changes."
    )
    args_schema: Type[BaseModel] = _WriteArgs
    allowed_suffixes: tuple[str, ...] = ()

    def _run(self, path: str, content: str) -> str:
        def go() -> str:
            p = self.sb.resolve(path, write=True)
            if self.allowed_suffixes and not path.endswith(self.allowed_suffixes):
                return f"ERROR: this agent may only write {', '.join(self.allowed_suffixes)} files"
            p.parent.mkdir(parents=True, exist_ok=True)
            existed = p.exists()
            p.write_text(content if content.endswith("\n") else content + "\n", encoding="utf-8")
            return f"{'Updated' if existed else 'Created'} {path} ({len(content.splitlines())} lines)"
        return self._safe(go)


class _EditArgs(BaseModel):
    path: str = Field(..., description="File path relative to the repo root")
    old_text: str = Field(..., description="Exact text to replace; must appear exactly once")
    new_text: str = Field(..., description="Replacement text (empty string deletes old_text)")


class EditFileTool(_RepoTool):
    name: str = "edit_file"
    description: str = (
        "Replace one exact, unique snippet in a file. Include enough surrounding lines "
        "to make old_text unique. Whitespace and tabs must match exactly."
    )
    args_schema: Type[BaseModel] = _EditArgs
    allowed_suffixes: tuple[str, ...] = ()

    def _run(self, path: str, old_text: str, new_text: str) -> str:
        def go() -> str:
            p = self.sb.resolve(path, write=True)
            if self.allowed_suffixes and not path.endswith(self.allowed_suffixes):
                return f"ERROR: this agent may only edit {', '.join(self.allowed_suffixes)} files"
            if not p.is_file():
                return f"ERROR: '{path}' does not exist (use write_file to create it)"
            text = p.read_text(encoding="utf-8")
            count = text.count(old_text)
            if count == 0:
                return "ERROR: old_text not found. Re-read the file and copy the text exactly."
            if count > 1:
                return f"ERROR: old_text appears {count} times. Add surrounding lines to make it unique."
            p.write_text(text.replace(old_text, new_text, 1), encoding="utf-8")
            return f"Edited {path}"
        return self._safe(go)


class _ChecksArgs(BaseModel):
    area: str = Field("auto", description="'auto' (areas you changed), 'server', 'client' or 'all'")


class RunChecksTool(_RepoTool):
    name: str = "run_checks"
    description: str = (
        "Run the project's real checks (gofmt, go vet, go test ./..., tsc, client tests) "
        "and return pass/fail with failure output. Run this before you finish."
    )
    args_schema: Type[BaseModel] = _ChecksArgs
    checks: list[Check] = Field(default_factory=list, exclude=True)
    timeout: int = 900

    def _run(self, area: str = "auto") -> str:
        areas = None if area == "auto" else (["server", "client"] if area == "all" else [area])
        report = run_checks(self.sb.root, self.checks, areas=areas, timeout=self.timeout)
        if report.passed:
            return "ALL CHECKS PASSED\n" + report.summary()
        return report.summary() + "\n\n" + (report.failures() or "")


def read_tools(sandbox: Sandbox) -> list[BaseTool]:
    return [
        ListFilesTool(sandbox=sandbox),
        ReadFileTool(sandbox=sandbox),
        SearchCodeTool(sandbox=sandbox),
    ]


def write_tools(
    sandbox: Sandbox, checks: list[Check], timeout: int = 900, md_only: bool = False
) -> list[BaseTool]:
    suffixes = (".md",) if md_only else ()
    tools: list[BaseTool] = [
        WriteFileTool(sandbox=sandbox, allowed_suffixes=suffixes),
        EditFileTool(sandbox=sandbox, allowed_suffixes=suffixes),
        GitDiffTool(sandbox=sandbox),
    ]
    if not md_only:
        tools.append(RunChecksTool(sandbox=sandbox, checks=checks, timeout=timeout))
    return tools
