"""Keeps agent file access inside the worktree and away from secrets."""

from __future__ import annotations

import os
from pathlib import Path
from typing import Iterator

# Never readable or writable by agents.
DENY_DIRS = {".git", "node_modules", "certs", "sigs", "filedata", "sess", "quarantine", "input"}
# Readable, but agents may not change them (dependencies, CI, infra, the crew itself).
LOCKED_FILES = {
    "go.mod", "go.sum", "go.work", "go.work.sum",
    "package.json", "package-lock.json", "pnpm-lock.yaml",
    "lefthook.yml", "Dockerfile", "entrypoint.sh", "traefik.yml",
}
LOCKED_PREFIXES = ("Jenkinsfile", "docker-compose")
LOCKED_TOP_DIRS = {"automation"}


class SandboxError(Exception):
    pass


def _is_secret(name: str) -> bool:
    if name.startswith(".env") and name != ".env.example":
        return True
    return name.endswith((".pem", ".key", ".crt"))


class Sandbox:
    def __init__(self, root: Path | str):
        self.root = Path(root).resolve()

    def resolve(self, rel: str, *, write: bool = False) -> Path:
        rel = (rel or ".").strip()
        path = (self.root / rel).resolve()
        if path != self.root and self.root not in path.parents:
            raise SandboxError(f"'{rel}' is outside the worktree")
        parts = path.relative_to(self.root).parts
        if any(p in DENY_DIRS for p in parts):
            raise SandboxError(f"'{rel}' is in a protected folder")
        if parts and _is_secret(parts[-1]):
            raise SandboxError(f"'{rel}' looks like a secret and is off limits")
        if write:
            if not parts:
                raise SandboxError("cannot write to the worktree root itself")
            name = parts[-1]
            if (
                parts[0] in LOCKED_TOP_DIRS
                or name in LOCKED_FILES
                or name.startswith(LOCKED_PREFIXES)
            ):
                raise SandboxError(
                    f"'{rel}' is locked (dependencies/CI/infra). Ask a human to change it."
                )
        return path

    def rel(self, path: Path) -> str:
        return path.resolve().relative_to(self.root).as_posix()

    def iter_files(self, start: Path) -> Iterator[Path]:
        for dirpath, dirnames, filenames in os.walk(start):
            dirnames[:] = sorted(d for d in dirnames if d not in DENY_DIRS and not d.startswith("."))
            for name in sorted(filenames):
                if name == ".DS_Store" or _is_secret(name):
                    continue
                yield Path(dirpath) / name
