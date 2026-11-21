---
type: Fact
title: "JS runtime host bindings: envelope via __host_call, throwing typed wrappers, bound only with Tools"
description: "JSRuntime binds __host_call plus a host shim per Call when Tools is set; host.call returns the ok/data/error envelope and typed wrappers throw, host undefined without tools."
tags: [internal/plugin, bindings_quickjs.go, js.go, plugin-host, quickjs]
generated: { by: agent/mcp, at: "2026-11-21T19:23:09Z" }
status: stable
---

## Evidence

`bindJS` (internal/plugin/bindings_quickjs.go) registers `__host_call(name, jsonArgsJSON)` via `quickjs.VM.RegisterHostFunc`; it returns the frozen `{ok,data,error}` envelope (shared `hostResultJSON`) and never throws into JS. A JS shim evals the global `host`: `call` returns the parsed envelope, `now`/`log.info`/`log.error`/`storage.get`/`storage.set` throw `new Error(res.error)` on `!ok` and yield `res.data`.

`JSRuntime.Call` binds only when `Tools != nil` (mirrors LuaRuntime). `TestJSHostBindings` (real `store.Open(tempdir)`): storage set/get round-trips `42`, absent key throws, `host.call("nope")` envelope has `bad_ok=false` with `unknown tool` error -> `ok internal/plugin 0.231s`, `CGO_ENABLED=0 go test ./internal/plugin/...` exit 0. `TestJSRuntimeWithoutToolsHasNoHost`: without tools `typeof host === "undefined"` -> `{"h":"undefined"}`.

## Consequence

JS plugins get parity with the Lua host surface; tool errors surface as catchable JS exceptions, and `host.call` stays the envelope escape hatch. Scripts running without a bound ToolSet must not touch `host`.

# Related Concepts
- [tools.ToolSet: storage delegates to plugin_state, missing keys return gorm not-found](plugin-tools-surface.md): JS host surface mirrors the Lua host surface over the same frozen tool set.
- [Go filename suffix trap: *_js.go is an implicit GOOS=js constraint](go-filename-goos-suffix-trap.md): The binding file lives as bindings_quickjs.go because the brief's bindings_js.go hit the GOOS suffix rule.
