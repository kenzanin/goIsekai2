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

**Dispatch at most two lanes in parallel.** The user allows a maximum of two active KIT lanes;
never run a third, and never let two lanes write overlapping files. For a failed lane, continue the
same session with `--continue` instead of respawning a fresh one. Never run two `--continue` lanes
at once: both pick the most recent session for the cwd and fight over the JSONL file. For parallel
follow-ups use fresh sessions or `--session=<jsonl-path>`.

## Language

Every artifact this project produces is **English**: source, comments, godoc, test names, docs, specs,
knowledge, commit messages, worker briefs and worker reports. Indonesian is allowed only in chat replies
to the user. Upstream values kept verbatim (status vocabulary, scraped titles, quoted third-party error
strings) are data, never translated. Full rule: `knowledge/conventions/language.md`.

## MCP tools (from `.kit.yml`)

- `codebase-memory-mcp` — `search_graph`, `trace_path`, `get_code_snippet`, `check_index_coverage`,
  `search_code`, `index_repository`. **goIsekai2 is not indexed as a graph project** (and has little
  code yet): query the `cppIsekai` project explicitly; goIsekai2 queries come back empty. Verified in a
  KIT lane: `search_code(project: "cppIsekai", pattern: "fts5")` returned `src/store/store_dedup.cpp`.
  Restored to `.kit.yml` on the user's instruction.
- `okf` — `okf_search`, `okf_show`, `okf_validate`, `okf_create`, `okf_update`. Workers **must** write back
  what they verified: `scripts/kit-worker` appends an OKF duty to every brief (search first, at most 3
  concepts of type `Fact`/`Decision` under `notes/<slug>`, evidence in the body, then
  `okf validate knowledge --strict --drift`). The main session reviews those concepts like any other
  diff: unverified or duplicated notes get fixed or dropped.

Built-in tools (`read`, `grep`, `find`, `ls`, `write`, `edit`, `shell`) work headless with **no
approval prompts**.

## Decision model (systemone)

`scripts/kit-worker` passes `--decision-model typesafe/oc/jev-1.13-free` (override with
`KIT_DECISION_MODEL`). KIT has a **native** `systemone` decision wire: with the wrapper's
`--provider-url` (`http://localhost:20128/v1`) and router key it resolves to
`POST /v1/systemone`, not chat completions. Probed through KIT's own decision client:
`RESOLVED wire=systemone url=http://localhost:20128/v1/systemone model=oc/jev-1.13-free`,
`DECIDE_OK probability=0.96`. An extension must call `ctx.Decide(...)` to use it; KIT warns at
startup if the configured value is not a decision model.

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
- **Specs stay read-only** (`openspec/` is the main session's); knowledge is written **only** through the
  appended OKF duty — never free-form edits to existing concepts, never a note without the evidence that
  proved it.
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
- Default chat model is `custom/adv/msi/GLM-5.3-Flash-EXL3`, with fallback `custom/agnes/agnes-3.0-flash`.
  This follows the documented KIT rule for `--provider-url`: with an explicit model KIT routes through
  the `custom/` OpenAI-compatible wire and strips the prefix from the model id. The earlier
  `openai/` prefix happened to work only because it was stripped too; `adv/msi/...` without a prefix
  failed with `No active credentials for provider: msi`. `KIT_MODEL` overrides the default.
  Curl against `/v1/chat/completions` answered `KIT_OK` (HTTP 200) for the full `adv/msi/...` id and
  for both `agnes/agnes-3.0-flash` and `agnes/agnes-2.5-flash`; KIT with `custom/` answered `KIT_OK`
  for all three. `openai/free` is a **combo alias**, not a single model: each request is routed to a
  different upstream, some of which answer with an empty body — KIT then exits 0 with no files.
  `scripts/kit-worker` retries an empty answer on the pinned fallback model.
- Do not route lanes to `oc/*-free` models: through the router they answer `FreeUsageLimitError`
  (free models only work inside OpenCode itself), and `oc/muse-spark-1.3-contributor-free` answers
  `Model is unavailable`. Direct `adv/*-free` ids return 401/503 through the router as well. The
  Working pinned choices are `custom/adv/msi/GLM-5.3-Flash-EXL3` and
  `custom/agnes/agnes-{3.0,2.5}-flash`.
- The launcher deliberately passes no `--bare`/`--no-skills`/`--no-agents` flags (user direction):
  lane activity stays visible in the herdr tab, at the cost of a larger system prompt.
- No UI/visual verification through KIT; the main session verifies rendering.
- KIT session history is local JSONL — durable knowledge belongs in OKF + OpenSpec + Magic Context,
  not in a worker's transcript.

## Optional jev / systemone tool (opt-in)

`KIT_JEV=1 ./scripts/kit-worker "..."` adds the local jev MCP server with
`--mcp "jev=<node> ~/.config/opencode/vendor/jev-mcp/dist/index.js"`, exporting
`JEV_API_KEY` from the wrapper's resolved router key (never committed), plus
`JEV_BASE_URL`, `JEV_MODEL=oc/jev-1.13-free` and `JEV_TIMEOUT_MS`. The lane brief
gains one line with the tool's real schema:
`{texts: string[], instructions: string, options?: {label: description}}`.
Default lanes do not pay for those tool schemas.

Verified: a KIT lane called `jev.classify` through this MCP server and received
`{"model":"oc/jev-1.13-free","results":[{"index":0,"choice":"yes",...}]}`
with a `0.91` confidence. A wrong first call using the raw `/v1/systemone`
schema (`state`/`questions`) failed input validation; the tool schema above is
the one that works.

systemone is **not** a drop-in KIT model: `--provider-wire` accepts only
`openai`, `openai-compat`, `anthropic` and `google`, and a chat-completions
request with `oc/jev-1.13-free` returns `400 ModelProtocolUnsupported`. Use it
as a tool for classification/decision calls, not as the lane's reasoning model.
