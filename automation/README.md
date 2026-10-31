# dev-crew: crewAI automation for the Purchase Record app

A [crewAI](https://docs.crewai.com) Flow that takes a Markdown **task brief** and
produces a reviewed, tested change on its own git branch, plus a suggested
commit message. **It never commits or pushes.** You review and commit.

```
brief.md
  -> prepare    new branch crew/<brief>-<time> in an isolated git worktree
  -> plan       Planner agent reads CLAUDE.md, SERVICE_LAYER.md and the code -> plan.json
  -> implement  Go developer (+ Frontend developer if the plan touches the client)
  -> verify     real checks, no LLM: gofmt, go vet, go test ./..., tsc, client tests
                  fail -> back to implement with the error output (max MAX_ATTEMPTS)
  -> review     Reviewer agent: correctness, clean architecture, tests, security
                  changes requested -> back to implement with the issues
  -> finalize   Docs agent updates SERVICE_LAYER.md / TODO.md, writes COMMIT_MSG.txt
  -> report     runs/<time>-<brief>/report.md with the exact commands to commit or discard
```

## Setup (once)

Requires Python 3.10-3.13 and [uv](https://docs.astral.sh/uv/), plus the tools
the checks use: Go 1.25 and Node/pnpm (with `client-vite/client-vite/node_modules`
installed in your main checkout; worktrees reuse it through a symlink).

```bash
cd automation
uv sync
cp .env.example .env      # set MODEL and the API key
uv run pytest             # offline tests, no API key needed
```

## Run

```bash
cd automation
uv run dev-crew briefs/maintenance-cleanup.md
uv run dev-crew briefs/security-jwt-token-type.md --max-attempts 4 --verbose
uv run dev-crew briefs/my-task.md --coder claude-code   # Claude Code writes the code
uv run dev-crew briefs/my-task.md --base release         # branch from another ref
```

Then open `runs/<latest>/report.md`. If the outcome is **ready**:

```bash
cd ../../2_Go-worktrees/crew-<brief>-<time>
git diff
git add -A && git commit -F <run_dir>/COMMIT_MSG.txt
# merge the crew/... branch into your feature branch, then:
git worktree remove <that folder>
```

## Watch runs in a UI (Arize Phoenix)

Phoenix runs locally, so prompts and code never leave your machine.

```bash
cd automation
uv sync --extra tracing                               # OpenInference + Phoenix exporter
docker compose -f phoenix/docker-compose.yml up -d    # Phoenix at http://localhost:6006
echo "TRACING=phoenix" >> .env
uv run dev-crew briefs/maintenance-cleanup.md
```

No Docker? `uvx --from arize-phoenix phoenix serve` starts the same UI.

Open http://localhost:6006, then the **dev-crew** project. Each run is one trace:

```
dev-crew.run                      outcome, branch, attempts, report path
├─ dev-crew.plan                  brief in, plan JSON out
│   └─ Crew / Agent / LLM / tool spans (planner reading files)
├─ dev-crew.implement #1          feedback in, developer summary out
│   └─ ... every edit_file / run_checks call and LLM call, with tokens
├─ dev-crew.verify #1   (red)     gofmt / go vet / go test output, route=retry
├─ dev-crew.implement #2
├─ dev-crew.verify #2
├─ dev-crew.review #2             approved / issues
└─ dev-crew.docs                  updated files + commit message
```

Red spans mark failed checks or rejected reviews. Click any LLM span to see the exact
prompt, response and token count. Spans are flushed when the run ends.
Leave `TRACING` empty to turn tracing off.

## Writing briefs

Copy `briefs/_TEMPLATE.md`. Good briefs name the files, state constraints and
say how "done" is checked. They are the same as the issue lists you already write in
module `CLAUDE.md` files. Ideas from the backlog:

| Brief | Source |
|---|---|
| `maintenance-cleanup.md` (included) | `server/internal/maintenance/CLAUDE.md` |
| `security-jwt-token-type.md` (included) | Strix vuln-0001 |
| Category service/repository unit tests | `server/TODO.md` #5, #11 |
| Generic login error (username enumeration) | Strix vuln-0002 |
| Block SVG uploads / sanitise (stored XSS) | Strix vuln-0005 |
| Error boundaries in the client | `client-vite/client-vite/TODO.md` |

## Safety rails

- Work happens in a separate **git worktree and branch**; your checkout is untouched.
- Agents can only reach files inside the worktree. `.env*`, `*.pem`, `certs/`, `sigs/`,
  `filedata/`, `sess/`, `input/`, `.git/` and `node_modules/` are hidden.
- Dependency, CI and infra files (`go.mod`, `go.sum`, `package.json`, lockfiles,
  `Jenkinsfile*`, `docker-compose*`, `Dockerfile`, `lefthook.yml`) and `automation/`
  are read-only to agents.
- The docs agent can only write `.md` files.
- Pass/fail comes from running real commands, not from an agent's opinion.

## Layout

```
automation/
  briefs/                 task briefs (inputs)
  src/dev_crew/
    main.py               DevFlow + CLI
    crews.py              agents/tasks from YAML, structured outputs (Plan, ReviewResult...)
    config/agents.yaml    roles, goals, backstories
    config/tasks.yaml     task prompts
    tools/repo_tools.py   sandboxed list/read/search/edit/write/diff/run_checks tools
    checks.py             verification commands (see settings.default_checks)
    gitops.py             worktree + diff helpers
    sandbox.py            path rules
    claude_code.py        optional `claude -p` coder
    tracing.py            optional Phoenix tracing (TRACING=phoenix)
  phoenix/                docker-compose for the local Phoenix UI
  tests/                  offline tests (fake runner drives the whole flow)
  runs/                   per-run plan, check logs, review, report (gitignored)
```

## Tuning

- Edit prompts in `config/*.yaml`, not in Python.
- Change check commands in `settings.default_checks()` (for example `CLIENT_TEST_CMD`).
- Costs: every attempt re-runs the developer and the reviewer. Start with
  `MAX_ATTEMPTS=3` and keep briefs small.
