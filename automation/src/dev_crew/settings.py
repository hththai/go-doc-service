"""Configuration: repo paths, verification commands and run limits."""

from __future__ import annotations

import os
from dataclasses import dataclass, field
from pathlib import Path

AUTOMATION_DIR = Path(__file__).resolve().parents[2]
DEFAULT_REPO_ROOT = AUTOMATION_DIR.parent

# Top-level path prefix -> area. An area decides which checks run.
AREA_PREFIXES: dict[str, tuple[str, ...]] = {
    "server": ("server/", "ocr/", "go.work"),
    "client": ("client-vite/",),
}


@dataclass(frozen=True)
class Check:
    """One verification command, mirroring lefthook.yml / Jenkinsfile."""

    name: str
    area: str  # "server" | "client"
    cwd: str  # relative to the worktree root
    cmd: str
    fail_on_output: bool = False  # e.g. `gofmt -l` prints unformatted files


def default_checks() -> list[Check]:
    client_test = os.getenv("CLIENT_TEST_CMD", "pnpm test")
    return [
        Check("gofmt", "server", "server", "gofmt -l .", fail_on_output=True),
        Check("go vet", "server", "server", "go vet ./..."),
        Check("go test", "server", "server", "go test ./..."),
        Check("tsc", "client", "client-vite/client-vite", "npx tsc --noEmit"),
        Check("client test", "client", "client-vite/client-vite", client_test),
    ]


@dataclass
class Settings:
    repo_root: Path = DEFAULT_REPO_ROOT
    model: str = "anthropic/claude-sonnet-5-5"
    coder: str = "crew"  # "crew" | "claude-code"
    max_attempts: int = 3
    base_ref: str = "HEAD"
    worktree_root: Path = DEFAULT_REPO_ROOT.parent / f"{DEFAULT_REPO_ROOT.name}-worktrees"
    runs_dir: Path = AUTOMATION_DIR / "runs"
    checks: list[Check] = field(default_factory=default_checks)
    check_timeout: int = 900
    verbose: bool = False

    @classmethod
    def from_env(cls, **overrides) -> "Settings":
        repo_root = Path(os.getenv("REPO_ROOT", DEFAULT_REPO_ROOT)).resolve()
        values = dict(
            repo_root=repo_root,
            model=os.getenv("MODEL", cls.model),
            coder=os.getenv("CODER", cls.coder),
            max_attempts=int(os.getenv("MAX_ATTEMPTS", cls.max_attempts)),
            worktree_root=Path(
                os.getenv("WORKTREE_ROOT", repo_root.parent / f"{repo_root.name}-worktrees")
            ),
        )
        values.update({k: v for k, v in overrides.items() if v is not None})
        return cls(**values)
