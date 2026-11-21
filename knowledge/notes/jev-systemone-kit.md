---
type: Fact
title: "systemone/jev is a KIT MCP tool, not a KIT model"
description: "The jev classifier works inside KIT through the local MCP server when KIT_JEV=1; the systemone protocol cannot replace KIT's chat-completion provider."
tags: [kit, jev, systemone, mcp, tooling]
generated: { by: agent/cli, at: "2026-11-21T14:51:50Z" }
status: stable
---

## Probe results

- `POST http://localhost:20128/v1/systemone` with model `oc/jev-1.13-free`,
  `state` and a `noul` question returned HTTP 200:
  `{"model":"jev-1.13-free","answers":{"is_urgent":{"type":"noul","noul":0.96}},...}`.
- `POST /v1/chat/completions` with the same model returned HTTP 400
  `ModelProtocolUnsupported` — the model is a classifier, not a chat model.
- `kit --help` lists `--provider-wire openai|openai-compat|anthropic|google`;
  there is no systemone wire, so it cannot be set as KIT's `--model`.
- A local MCP server already exists at
  `~/.config/opencode/vendor/jev-mcp/dist/index.js` (env: `JEV_API_KEY`,
  `JEV_BASE_URL`, `JEV_MODEL`, `JEV_TIMEOUT_MS`). Passing
  `--mcp "jev=$(command -v node) <that path>"` to KIT exposed `jev.classify`.
- A KIT lane called `jev.classify` successfully with the tool's own schema:
  `{texts: string[], instructions: string, options?: {label: description}}`,
  returning `{"model":"oc/jev-1.13-free","results":[{"index":0,"choice":"yes",...}]}`
  and confidence `0.91`. The raw systemone shape (`state`/`questions`) fails
  MCP input validation.

## Wiring

`scripts/kit-worker` exposes this as opt-in `KIT_JEV=1`: it adds the `--mcp`
flag, exports the router key as `JEV_API_KEY` (never committed), derives
`JEV_BASE_URL` from `KIT_PROVIDER_URL` by stripping `/v1`, and appends the tool
schema to the brief. Default lanes are unchanged. Full instructions: the
repo-local kit-worker skill.

# Related Concepts
- [KIT headless worker harness (.kit.yml + scripts/kit-worker + kit-worker skill)](../references/kit-worker.md): records the opt-in jev MCP tool added to the KIT worker harness
