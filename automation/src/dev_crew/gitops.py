"""Git worktree handling and diff collection. The crew never commits or pushes."""

from __future__ import annotations

import re
import subprocess
from pathlib import Path

NODE_MODULES_REL = Path("client-vite/client-vite/node_modules")


def git(*args: str, cwd: Path) -> str:
    result = subprocess.run(
        ["git", *args], cwd=cwd, capture_output=True, text=True, check=False
    )
    if result.returncode != 0:
        raise RuntimeError(f"git {' '.join(args)} failed: {result.stderr.strip()}")
    return result.stdout


def slugify(text: str, max_len: int = 40) -> str:
    slug = re.sub(r"[^a-z0-9]+", "-", text.lower()).strip("-")
    return slug[:max_len].rstrip("-") or "task"


def create_worktree(repo_root: Path, worktree_root: Path, branch: str, base: str = "HEAD") -> Path:
    """Create an isolated worktree on a new branch so the user's checkout is untouched."""
    worktree_root.mkdir(parents=True, exist_ok=True)
    path = worktree_root / branch.replace("/", "-")
    git("worktree", "add", "-b", branch, str(path), base, cwd=repo_root)
    # Reuse the main checkout's node_modules so client checks run without a reinstall.
    src = repo_root / NODE_MODULES_REL
    dst = path / NODE_MODULES_REL
    if src.is_dir() and dst.parent.is_dir() and not dst.exists():
        dst.symlink_to(src, target_is_directory=True)
    return path


def changed_paths(worktree: Path) -> list[str]:
    """Paths changed (tracked or untracked) relative to HEAD, ignoring node_modules."""
    out = git("status", "--porcelain", "--untracked-files=all", cwd=worktree)
    paths: list[str] = []
    for line in out.splitlines():
        if len(line) < 4:
            continue
        path = line[3:]
        if " -> " in path:
            path = path.split(" -> ", 1)[1]
        path = path.strip().strip('"')
        if "node_modules" in Path(path).parts:
            continue
        paths.append(path)
    return paths


def full_diff(worktree: Path, max_chars: int = 60_000) -> str:
    """Tracked changes plus the contents of new files."""
    parts = [git("diff", "HEAD", cwd=worktree)]
    untracked = git("ls-files", "--others", "--exclude-standard", cwd=worktree).splitlines()
    for rel in untracked:
        if "node_modules" in Path(rel).parts:
            continue
        result = subprocess.run(
            ["git", "diff", "--no-index", "--", "/dev/null", rel],
            cwd=worktree, capture_output=True, text=True, check=False,
        )
        parts.append(result.stdout)
    diff = "\n".join(p for p in parts if p.strip())
    if len(diff) > max_chars:
        diff = diff[:max_chars] + f"\n... [diff truncated, {len(diff) - max_chars} more chars]"
    return diff or "(no changes)"


def diff_stat(worktree: Path) -> str:
    stat = git("diff", "HEAD", "--stat", cwd=worktree).strip()
    new = [p for p in git("ls-files", "--others", "--exclude-standard", cwd=worktree).splitlines()
           if "node_modules" not in Path(p).parts]
    if new:
        stat += ("\n" if stat else "") + "\n".join(f" {p} (new)" for p in new)
    return stat or "(no changes)"
