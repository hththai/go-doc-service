"""Collects the project rules the planner should always see."""

from __future__ import annotations

from pathlib import Path

DOCS = [
    ("CLAUDE.md", 4000),
    ("server/CLAUDE.md", 4000),
    ("client-vite/client-vite/CLAUDE.md", 4000),
    ("server/internal/SERVICE_LAYER.md", 9000),
]
LAYOUT_DIRS = ["server/api/v1", "server/internal", "server/middleware", "client-vite/client-vite/src"]
SKIP = {"node_modules", ".DS_Store"}


def _layout(root: Path, rel: str, limit: int = 150) -> str:
    base = root / rel
    if not base.is_dir():
        return ""
    files = sorted(
        p.relative_to(root).as_posix()
        for p in base.rglob("*")
        if p.is_file() and not (set(p.parts) & SKIP)
    )
    extra = f"\n... ({len(files) - limit} more)" if len(files) > limit else ""
    return "\n".join(files[:limit]) + extra


def collect_context(worktree: Path) -> str:
    parts = []
    for rel, limit in DOCS:
        p = worktree / rel
        if p.is_file():
            text = p.read_text(encoding="utf-8")
            if len(text) > limit:
                text = text[:limit] + "\n... [truncated, use read_file for the rest]"
            parts.append(f"--- {rel} ---\n{text}")
    layout = "\n".join(f"[{d}]\n{_layout(worktree, d)}" for d in LAYOUT_DIRS)
    parts.append(f"--- file layout ---\n{layout}")
    return "\n\n".join(parts)
