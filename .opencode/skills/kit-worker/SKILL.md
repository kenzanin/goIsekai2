---
name: kit-worker
description: Delegate bounded implementation tasks (fixer/worker) to the local KIT CLI headless with a free router model and read/write access to this Go project plus codebase-memory-mcp and okf MCP. Use for scoped Go edits, test fixes, and mechanical multi-file changes; keep review, spec/OKF writes, commits and final verification in the main session.
---

# KIT as a goIsekai2 worker

KIT (`~/.local/bin/kit`, v0.143.0) is a Go coding-agent CLI. This repo has a permanent worker
config: `.kit.yml` (MCP servers + `openai/free` model) and `scripts/kit-worker` (provider env from
`OMNIROUTE_API_KEY`, falling back to the local 9router key in
`~/.config/opencode/opencode.json`). Use it for **bounded** implementation tasks; the main session
stays the reviewer, the spec author and the final verifier.

## Mandatory: every lane runs in its own herdr tab

`./scripts/kit-worker` creates a new herdr tab (`kit:<prompt prefix>`), starts KIT inside it, waits for the
lane's `__KIT_EXIT=` sentinel, prints the tail, and leaves the tab open for inspection. Rules:

- **Never run `kit` directly** and never launch a lane from a bare background shell — the tab is how the
  user watches a lane, and how a stalled lane is diagnosed.
- **Outside herdr the wrapper refuses** (exit 64) unless `KIT_ALLOW_HEADLESS=1` is set. Do not set it for
  real lanes; it exists for CI or a broken herdr server only.
- The wrapper prints `kit-worker: lane tab <tab> pane <pane>` on stderr. Quote that id to the user, and use
  `herdr pane read <pane> --source recent-unwrapped --lines 80` to inspect without disturbing the lane.
- A lane that ends without the sentinel is **not** a success: read the tab, then resume with `--continue`.
  Never respawn a fresh session for the same lane.
- Keep the tab until the lane is reviewed; close it with `herdr tab close <tab>` afterwards.

## Invoke

```bash
cd /home/kenzanin/tmp/go/goIsekai2
./scripts/kit-worker "<bounded prompt>"            # runs in a new herdr tab, prints the tail
./scripts/kit-worker --json "<prompt>" | jq .      # JSON envelope
./scripts/kit-worker --continue "<follow-up>"      # resume the lane's session
```

The wrapper resolves the key from `OMNIROUTE_API_KEY`, else the 9router entry in
`~/.config/opencode/opencode.json`, then exports `KIT_PROVIDER_API_KEY`, `KIT_PROVIDER_URL` (default
`http://localhost:20128/v1`) and `KIT_PROVIDER_WIRE=openai-compat`. Sessions persist as tree-based JSONL
under `~/.kit/sessions/<cwd>/`, so a lane that ran out of steps is continued, not restarted. Override
`KIT_PROVIDER_URL` / `KIT_BIN` / `KIT_WAIT_MS` if needed.

**Dispatch serially.** One worker at a time — the user requires serial fixer delegation. Never run two
`kit-worker` invocations whose file scopes could overlap.

## MCP tools (from `.kit.yml`)

- `codebase-memory-mcp` — `search_graph`, `trace_path`, `get_code_snippet`, `check_index_coverage`,
  `search_code`, `index_repository`. **goIsekai2 is not indexed as a graph project** (and has little
  code yet): expect empty results and require the worker to read source directly and say so in its
  finding instead of guessing.
- `okf` — `okf_search`, `okf_show`, `okf_validate`. Read-only use by default; see the guardrails.

Built-in tools (`read`, `grep`, `find`, `ls`, `write`, `edit`, `shell`) work headless with **no
approval prompts**.

## Delegation contract

Give the worker a bounded, checkable change: exact file(s), symptom, expected behaviour, and the gate
it must pass. Good:

```bash
./scripts/kit-worker "In internal/config/ini.go only: make Parse preserve duplicate keys verbatim.
Add table-driven cases in internal/config/ini_test.go. Gate: gofmt -l internal/config empty and
CGO_ENABLED=0 go test ./internal/config. Do not touch any other file, do not commit."
```

Bad: "implement the config package" — that is a lane the main session owns.

### Guardrails (put them in every brief)

- **One lane, non-overlapping files.** The worker and the main session never write the same file.
- **No `git commit`/`push`, no branch operations, no `openspec new/archive`, no daemon restarts** —
  the main session owns those.
- **No cgo, ever.** `CGO_ENABLED=0 go build ./...` is the gate; a worker that reaches for a cgo
  library has misread the project.
- **No new dependency** without a decision from the main session (the stack is fixed in
  `knowledge/decisions/dependency-stack.md`).
- **Specs and knowledge are read-only to the worker** unless it was explicitly told to record one
  fact; the main session writes `openspec/` and OKF concepts.
- **Tests next to the package**, stdlib `testing` only, no new test framework.
- **Report `git diff --stat`** at the end plus the exact gate command and its exit status.

## Verification (mandatory, main session)

```bash
git diff                                  # the diff is the evidence, not the worker's summary
CGO_ENABLED=0 go build ./... && go vet ./... && go test ./... && gofmt -l .   # empty output = clean
```

Then the project's own proof for the capability being built (HTTP route via `httptest` or curl
against the daemon, ini bytes on disk, `config.changed` recorder output). A worker's "done" is a
claim, not evidence — amend the same session with `--continue` instead of spawning a fresh one.

## Known limits

- Headless KIT executes tools with no approval prompt and no OS sandbox; a vague brief can write the
  wrong file. Keep jobs to one file or one shape fix.
- The `openai/free` router model (GLM-4.7) is decent at tool discipline but wanders on multi-step
  ambiguity, and sometimes ends a response with nothing at all. That shows up as a lane with no
  `__KIT_EXIT=` sentinel: read the tab, then `--continue` with a shorter, more imperative brief
  ("write these 5 files now, then run these gates") instead of re-reading context.
- No UI/visual verification through KIT; the main session verifies rendering.
- KIT session history is local JSONL — durable knowledge belongs in OKF + OpenSpec + Magic Context,
  not in a worker's transcript.
