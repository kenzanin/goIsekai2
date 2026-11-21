---
type: Reference
title: KIT headless worker harness (.kit.yml + scripts/kit-worker + kit-worker skill)
description: "How goIsekai2 drives the local KIT CLI as a bounded fixer with the free router model, which MCP servers it gets, the serial-dispatch and no-cgo guardrails, and the main-session verification gate that a worker run must pass."
tags: [tooling, delegation, kit, worker, harness]
generated: { by: agent/mcp, at: "2026-11-21T13:38:45Z" }
status: stable
---

# KIT headless worker

Ported from cppIsekai. Four pieces, all in this repo:

| File | Role |
|---|---|
| `.kit.yml` | worker config: `model: openai/free`, MCP servers `codebase-memory-mcp` and `okf` (`okf mcp ./knowledge`) |
| `scripts/kit-worker` | wrapper: creates a **herdr tab per lane**, runs KIT inside it, waits for the `__KIT_EXIT=` sentinel, prints the tail |
| `.opencode/skills/kit-worker/SKILL.md` | the delegation contract the worker itself follows |
| this concept | project-side record for review and for the next session |

## Mandatory: one herdr tab per lane

```bash
./scripts/kit-worker "<bounded prompt>"          # new herdr tab, prints tail, keeps tab open
./scripts/kit-worker --continue "<follow-up>"    # resume that lane's session
./scripts/kit-worker --json "<prompt>" | jq .
```

- The wrapper prints `kit-worker: lane tab <tab> pane <pane>` on stderr; that id is how a stalled lane
  is inspected (`herdr pane read <pane> --source recent-unwrapped --lines 80`).
- Outside herdr it refuses with exit 64 unless `KIT_ALLOW_HEADLESS=1` (CI / broken server only).
- No `__KIT_EXIT=` sentinel means the lane did **not** finish: read the tab, then `--continue`. Never
  respawn a fresh session for the same lane.
- Close the tab after review: `herdr tab close <tab>`.

## Context ceiling (measured 2026-11-21)

The free router model ends a lane **silently** — exit 0, no answer, no error — once the conversation
passes roughly 8k tokens. A lane brief that told the worker to read `tasks.md` + `design.md` +
`spec.md` accumulated 34,795 chars (~8.7k tokens) of tool output and died on the next request.

Rules that follow:

- The brief carries the contract inline. Do not send a lane to read a large spec; paste the exact
  struct, defaults and gate commands into the brief.
- At most one small file read per lane (~4k chars). Several small lanes beat one big lane.
- Split by file/shape, not by feature: "write registry.go + its test, run build" is a lane;
  "implement the config capability" is not.

## Guardrails for every brief

- **Serial dispatch** — one worker at a time; no overlapping write scope with the main session.
- **No commit/push, no branch work, no `openspec new/archive`, no daemon restart** — main session owns those.
- **No cgo, no new dependency** — `CGO_ENABLED=0 go build ./...` is the gate and the stack is fixed by
  `decisions/dependency-stack`.
- **Godoc rule applies** (`conventions/godoc.md`): package comment plus name-first doc comment on every
  exported symbol.
- **`git diff --stat` + gate commands + exit statuses** in the answer; a worker's "done" is a claim, and
  the main session re-runs the gates and reads the diff.
- The worker never ticks `openspec` task checkboxes; the main session does, after verification.

## Known limits

- goIsekai2 is not indexed in codebase-memory, so graph queries return nothing — the worker must read
  source and say so.
- `openai/free` (GLM-4.7 via 9router) wanders on vague multi-step prompts and sometimes returns an empty
  completion near its context limit.
- No UI/visual verification through KIT; no approval prompts (headless, non-interactive).

## Verified

- `kit --version` → 0.143.0; `/v1/models` → 200; alias `free` resolves (smoke answered as `GLM-4.7`).
- Shell tool works headless: a worker run of `CGO_ENABLED=0 go build ./...` reported `EXIT=0`.
- herdr tab path verified end to end: the wrapper created tab `w1:t3` (pane `w1:p3`), the worker wrote
  `internal/config/doc.go` inside that tab and reported `__KIT_EXIT=0`.

# Related Concepts
- [Capability porting order](../process/porting-order.md)
- [Dependency stack decision](../decisions/dependency-stack.md)
- [Godoc comment rule](../conventions/godoc.md)
