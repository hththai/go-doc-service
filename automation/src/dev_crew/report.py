"""Human-readable run report with the exact commands to commit or discard."""

from __future__ import annotations

from pathlib import Path


def write_report(state, stat: str) -> Path:
    run_dir = Path(state.run_dir)
    review = state.review or {}
    issues = "\n".join(f"- {i}" for i in review.get("issues", [])) or "- none"
    plan = state.plan or {}
    steps = "\n".join(f"{i}. `{s['file']}` ({s['layer']}): {s['change']}"
                      for i, s in enumerate(plan.get("steps", []), 1)) or "(no plan)"

    if state.outcome == "ready":
        next_steps = f"""Review, then commit (the crew never commits for you):

```bash
cd "{state.worktree}"
git diff
git add -A
git commit -F "{run_dir / 'COMMIT_MSG.txt'}"
```

Merge `{state.branch}` into your feature branch, then clean up:

```bash
git worktree remove "{state.worktree}"
```"""
    else:
        next_steps = f"""The crew stopped after {state.attempt} attempt(s). Inspect the work, fix it by hand
or refine the brief and run again. To discard everything:

```bash
git worktree remove --force "{state.worktree}" && git branch -D {state.branch}
```

Last feedback given to the developer:

{state.feedback or '(none)'}"""

    commit = f"```\n{state.commit_message}\n```" if state.commit_message else "(not written)"
    text = f"""# dev-crew run: {state.title}

- **Outcome:** {state.outcome}
- **Branch:** `{state.branch}`
- **Worktree:** `{state.worktree}`
- **Attempts:** {state.attempt}/{state.max_attempts}

## Plan
{plan.get('summary', '')}

{steps}

## Checks (last run)
```
{state.check_summary}
```

## Review
{review.get('summary', '(not reviewed)')}

Issues:
{issues}

## Changed files
```
{stat}
```

## Docs updated
{chr(10).join(f'- {f}' for f in state.docs_files) or '- none'}

## Suggested commit
{commit}

## Next steps
{next_steps}
"""
    path = run_dir / "report.md"
    path.write_text(text, encoding="utf-8")
    return path
