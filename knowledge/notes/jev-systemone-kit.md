---
type: Fact
title: "systemone/jev is a KIT MCP tool, not a KIT model"
description: "KIT has a native systemone decision wire: --decision-model typesafe/oc/jev-1.13-free targets /v1/systemone and returns probabilities; the MCP tool route is also available."
tags: [kit, jev, systemone, mcp, tooling]
generated: { by: agent/cli, at: "2026-11-21T15:05:57Z" }
status: stable
---

## Probe results

- `POST http://localhost:20128/v1/systemone` with model `oc/jev-1.13-free`,
  `state` and a `noul` question returned HTTP 200:
  `{"model":"jev-1.13-free","answers":{"is_urgent":{"type":"noul","noul":0.96}},...}`.
- `POST /v1/chat/completions` with the same model returned HTTP 400
  `ModelProtocolUnsupported` — it is a classifier, not a chat model.
- The local MCP server at
  `~/.config/opencode/vendor/jev-mcp/dist/index.js` exposes `jev.classify` with
  schema `{texts: string[], instructions: string, options?: {label: description}}`.
  A KIT lane called it and got `choice: "yes"` with confidence `0.91`. The raw
  systemone shape (`state`/`questions`) fails MCP input validation.

## Native decision wire (the better route)

KIT itself speaks systemone for decision models (`internal/models/decision_wire.go`:
`DecisionWireSystemOne` → `POST {base}/systemone`; the `typesafe` provider override
accepts any model). Using KIT's own client with
`ProviderConfig{ModelString: "typesafe/oc/jev-1.13-free", ProviderURL: "http://localhost:20128/v1"}`:

```
RESOLVED wire=systemone url=http://localhost:20128/v1/systemone model=oc/jev-1.13-free
DECIDE_OK probability=0.96
```

So `scripts/kit-worker` now passes
`--decision-model "${KIT_DECISION_MODEL:-typesafe/oc/jev-1.13-free}"` with the
existing router URL/key. An extension must call `ctx.Decide(...)` to use it;
chat models cannot serve decision requests and `SetModel` rejects them.
`KIT_JEV=1` remains as an opt-in route for ad-hoc classification via the MCP tool.

## Wiring notes

The launcher no longer passes `--quiet --bare --no-skills --no-agents
--no-extensions` (user direction): lane activity is visible in the herdr tab.
`codebase-memory-mcp` is back in `.kit.yml`; goIsekai2 is not graph-indexed, so
lanes must query the `cppIsekai` project explicitly.

# Related Concepts
- [KIT headless worker harness (.kit.yml + scripts/kit-worker + kit-worker skill)](../references/kit-worker.md): records both the native decision wire and the opt-in MCP route in the KIT launcher
