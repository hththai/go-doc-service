import subprocess
from pathlib import Path

import pytest


def _git(cwd: Path, *args: str) -> None:
    subprocess.run(["git", *args], cwd=cwd, check=True, capture_output=True)


@pytest.fixture()
def repo(tmp_path: Path) -> Path:
    """A tiny repo shaped like 2_Go: server/ and client-vite/client-vite/."""
    root = tmp_path / "2_Go"
    (root / "server/internal/demo").mkdir(parents=True)
    (root / "client-vite/client-vite/src").mkdir(parents=True)
    (root / "server/internal/demo/service.go").write_text("package demo\n\nfunc Ok() bool { return true }\n")
    (root / "server/go.mod").write_text("module demo\n")
    (root / "server/.env").write_text("SECRET=1\n")
    (root / "client-vite/client-vite/src/app.ts").write_text("export const x = 1\n")
    (root / "CLAUDE.md").write_text("# rules\n")
    (root / ".gitignore").write_text(".env\nnode_modules\n")
    _git(root, "init", "-q", "-b", "main")
    _git(root, "config", "user.email", "t@example.com")
    _git(root, "config", "user.name", "t")
    _git(root, "add", "-A")
    _git(root, "commit", "-qm", "init")
    return root
