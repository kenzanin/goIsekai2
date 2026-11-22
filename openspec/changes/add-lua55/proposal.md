# Proposal: add-lua55

## Why

Lua plugins are currently pinned to Lua 5.1 by `github.com/yuin/gopher-lua`. The shipped corpus is growing
and plugin authors should not be capped below current Lua syntax. `modernc.org/lua` v0.1.0 is a pure-Go
translation of PUC-Rio Lua 5.5.1 — from the same no-cgo family already used by sqlite and quickjs — so the
runtime can move to Lua 5.5 without changing the deployment model.

## What Changes

- Replace the Lua runtime implementation with `modernc.org/lua` v0.1.0.
- Port the frozen `host.*` bindings, `json` module, `require` shim and sandbox to the new VM API.
- Keep the exported `LuaRuntime` surface, JSON envelope and all routes unchanged.
- Remove `gopher-lua` from go.mod and verify `CGO_ENABLED=0` gates.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `plugin-host` — the `lua` runtime implementation and its LuaJIT-module failure wording change; the plugin
  ABI, discovery rules, lifecycle machine and host-tool surface do not.

## Impact

- Code: `internal/plugin/lua.go`, `internal/plugin/bindings_lua.go`, their tests, and go.mod/go.sum.
- Dependencies: add `modernc.org/lua` v0.1.0; upgrade transitive `modernc.org/cc*`/`libc` to the versions it requires.
- Docs: `decisions/lua-runtime`, `decisions/dependency-stack`, `architecture/stack`.
- No cgo: `CGO_ENABLED=0 go build ./...` remains the gate.
- Out of scope: JS/WASM runtimes, host tools, plugin corpus sources, UI and routes.

## Baseline

cppIsekai runs Lua plugins through its native runtime; the frozen contract is the JSON envelope, not the
engine version. This change is an internal implementation divergence recorded in `decisions/lua-runtime`.
