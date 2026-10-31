"""Optional coder: delegate the implementation step to Claude Code running headless."""

from __future__ import annotations

import shutil
import subprocess
from pathlib import Path

ALLOWED_TOOLS = ",".join([
    "Read", "Edit", "Write", "Glob", "Grep",
    "Bash(go build:*)", "Bash(go vet:*)", "Bash(go test:*)", "Bash(gofmt:*)",
    "Bash(npx tsc:*)", "Bash(pnpm test:*)", "Bash(npm run test:*)",
])

PROMPT = """You are implementing one planned change in this repository (a git worktree).
Follow CLAUDE.md files and server/internal/SERVICE_LAYER.md. Clean architecture:
api/v1 handler -> internal/<module>/service -> repository -> database.

Rules:
- Do not run git commit/push. Do not touch .env*, certs, go.mod/go.sum, package.json, CI files.
- Add or update tests in the module's test/ folder and keep them passing.
- Run gofmt, `go vet ./...` and `go test ./...` in server/ (and client checks if you touched the client).

## Brief
{brief}

## Plan
{plan}

## Feedback from previous attempt
{feedback}

When done, reply with a short summary of what you changed and the test results."""


def run_claude_code(worktree: Path, brief: str, plan: str, feedback: str, timeout: int = 1800) -> str:
    exe = shutil.which("claude")
    if not exe:
        return "ERROR: `claude` CLI not found on PATH. Install Claude Code or use CODER=crew."
    prompt = PROMPT.format(brief=brief, plan=plan, feedback=feedback or "None, first attempt.")
    proc = subprocess.run(
        [exe, "-p", prompt, "--permission-mode", "acceptEdits", "--allowedTools", ALLOWED_TOOLS],
        cwd=worktree, capture_output=True, text=True, timeout=timeout, check=False,
    )
    out = proc.stdout.strip()
    if proc.returncode != 0:
        out += f"\n[claude exited {proc.returncode}] {proc.stderr.strip()[-2000:]}"
    return out
