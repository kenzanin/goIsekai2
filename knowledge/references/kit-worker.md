---
type: Reference
title: KIT headless worker harness (.kit.yml + scripts/kit-worker + kit-worker skill)
description: "How goIsekai2 drives the local KIT CLI as a bounded fixer with the free router model, which MCP servers it gets, the serial-dispatch and no-cgo guardrails, and the main-session verification gate that a worker run must pass."
tags: [tooling, delegation, kit, worker, harness]
generated: { by: agent/cli, at: "2026-11-21T13:25:59Z" }
status: stable
---

# KIT headless worker

Ported from cppIsekai. Three files, all in this repo:

| File | Role |
|---|---|
| `.kit.yml` | worker config: `model: openai/free`, MCP servers `codebase-memory-mcp` and `okf` (`okf mcp ./knowledge`) |
| `scripts/kit-worker` | wrapper: sets `KIT_PROVIDER_API_KEY` (from `OMNIROUTE_API_KEY`, else the 9router key in `~/.config/opencode/opencode.json`), `KIT_PROVIDER_URL=http://localhost:20128/v1`, `KIT_PROVIDER_WIRE=openai-compat`, then `kit --quiet` |
| `.opencode/skills/kit-worker/SKILL.md` | the delegation contract (this doc is the project-side record of it) |

```bash
./scripts/kit-worker "<bounded prompt>"
./scripts/kit-worker --continue "<follow-up>"   # resume the same lane, do not respawn
./scripts/kit-worker --json "<prompt>" | jq .
```

Sessions persist as tree JSONL under `~/.kit/sessions/<cwd>/`. No API key is stored in the repo.

## Verified

- `kit --version` → 0.143.0; `/v1/models` → 200; model alias `free` resolves (smoke answered as
  `GLM-4.7`).
- Shell tool works headless: a worker run of `CGO_ENABLED=0 go build ./...` reported `EXIT=0`.

## Guardrails for every brief

- **Serial dispatch** — one worker at a time; no overlapping file scopes with the main session.
- **No commit/push, no branch work, no `openspec new/archive`, no daemon restart** — main session owns those.
- **No cgo, no new dependency** — `CGO_ENABLED=0 go build ./...` is the gate and the stack is fixed by
  `decisions/dependency-stack`.
- **Specs and OKF are read-only to the worker** unless it is told to record one specific fact.
- **`git diff` + gate command + exit status** in the finding; a worker's "done" is a claim, not evidence.

## Known limits

- goIsekai2 is not indexed in codebase-memory, so graph queries return nothing — the worker must fall
  back to reading source and say so.
- `openai/free` (GLM-4.7 via 9router) wanders on vague multi-step prompts; keep a job to one file or
  one shape fix.
- No UI/visual verification through KIT.

# Related Concepts
- [Capability porting order](../process/porting-order.md)
- [Dependency stack decision](../decisions/dependency-stack.md)
- [Capability porting order: one OpenSpec change per capability, config first](../process/porting-order.md): worker lanes accelerate the capability implementation sequence
