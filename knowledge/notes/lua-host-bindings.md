---
type: Fact
title: "Lua runtime host bindings: host.call envelope + typed wrappers, bound per Call state"
description: "internal/plugin binds a global host table into each Lua call state: host.call JSON envelope over ToolSet.Dispatch plus direct now/log/storage wrappers, verified green under CGO_ENABLED=0."
tags: [internal/plugin, bindings_lua.go, lua.go, plugin-host]
generated: { by: agent/mcp, at: "2026-11-21T19:15:48Z" }
status: stable
---

## Evidence

Fixture plugin calls `host.now`, `host.log.info`, `host.storage.set/get` (including a missing key) and `host.call` (success + unknown-tool) against a real sqlite store; `LuaRuntime.SetTools` binds `bindLua` into every `Call` state.

```
$ CGO_ENABLED=0 go test ./internal/plugin/...
ok  	goisekai2/internal/plugin	0.217s
ok  	goisekai2/internal/plugin/tools	0.037s
EXIT=0
$ go run scripts/godoc-check.go
EXIT=0
```

## Shape

- `host.call(name, jsonArgs)` -> JSON envelope `{ok,data,error}` (`error` omitted on success); args are a JSON array spread into `tools.ToolSet.Dispatch`.
- Thin wrappers hit `ToolSet` methods directly: `host.now()` -> number; `host.log.info(msg)` / `host.log.error(msg)`; `host.storage.set(k,v)` raises on failure; `host.storage.get(k)` -> `value` or `nil, errmsg` (absent keys are gorm not-found, per notes/plugin-tools-surface).

## Consequence

Lua plugins reach every host tool through `host.call` and the common ones through typed wrappers; the `NewLuaRuntime` validation state has no tools bound, so top-level (non-`handle`) host calls fail there — bind via a toolset-carrying constructor if that ever matters.

# Related Concepts
- [tools.ToolSet: storage delegates to plugin_state, missing keys return gorm not-found](plugin-tools-surface.md): host.* wrappers delegate to the same ToolSet storage/log/clock surface; absent-key behaviour matches its gorm not-found fact.
