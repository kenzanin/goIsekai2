---
type: decision
title: "Lua runtime: modernc.org/lua v0.1.0 for Lua 5.5.1 plugins"
description: "GoIsekai2 Lua plugins run on modernc.org/lua v0.1.0 (PUC-Rio Lua 5.5.1 translated to pure Go by ccgo), replacing gopher-lua v1.1.2 so plugins can use Lua 5.5 syntax while CGO_ENABLED=0 remains the build gate."
tags: [decision, lua, plugins, no-cgo, modernc]
generated: { by: agent/mcp, at: "2026-11-22T14:48:38Z" }
status: stable
---

## Decision

Use `modernc.org/lua` v0.1.0 as the only Lua runtime in `internal/plugin`. It is PUC-Rio Lua 5.5.1
translated to Go by ccgo: BSD-3-Clause, no cgo, official Lua test suite passes including internal tests.
`gopher-lua` v1.1.2 is removed from go.mod.

## Why

The plugin corpus is growing and Lua 5.1 caps plugin authors below current Lua syntax. The modernc engine is
from the same pure-Go family already used by sqlite and quickjs, so the no-cgo rule and the release pipeline
do not change. Measured against C Lua 5.5.1, the package docs report roughly 2.5-3x slower execution but
2-3x faster than gopher-lua; plugin calls are short compared with network fetches, so this is not the bottleneck.

## Probe evidence

Ran a throwaway `CGO_ENABLED=0 go run` module against v0.1.0:

- `lua.Version()` -> `Lua 5.5.1`; Lua `_VERSION` -> `Lua 5.5`
- `local x <const> = 5` compiled and returned, proving 5.5 syntax
- `SetGlobal("os", nil)` made `os.getenv` fail with `attempt to index a nil value`
- Nested host bindings (`host.text.clean`) worked after creating tables with `NewTable`/`Value.Set`:
  `RegisterFunc("host.now", ...)` alone does not create the parent table

## Porting contract

- `LuaRuntime` keeps its exported surface: `SetTools`, `SetDir`, `NewLuaRuntime`, `NewLuaRuntimeWithTools`, `Call`.
- Every call creates a fresh `lua.VM`; `Close` is deferred; timeouts use `SetEvalTimeout` plus a context
  watcher that calls `Interrupt`.
- Host bindings are built as `*lua.Value` tables with `Value.Set` for nested functions; the frozen JSON
  envelope is unchanged.
- `require` is a Go-backed global resolving `<name>.lua` only inside the plugin directory, with a per-VM
  cache; module return values are converted from `*lua.Value` to Go maps so `require("genres").map` keeps working.
- Sandbox disables `os`, `io`, `debug`, `dofile`, `loadfile` and the `ffi`/`bit` modules.
- Lua-to-Go conversion handles `int64` and `float64`, and numeric tables become JSON arrays as before.

## Alternatives

| Option | Verdict |
|---|---|
| Keep gopher-lua v1.1.2 | Rejected: Lua 5.1 forever; user asked for 5.5 |
| LuaJIT via cgo/purego | Rejected: cgo or per-platform native libraries violate the no-cgo rule |
| Hand-written Lua engine | Rejected: wasteful when a pure-Go reference implementation exists |
| wasm3 from modernc | Rejected: no published module; see `notes/modernc-wasm3-unavailable` |

## Consequences

- go.mod gains `modernc.org/lua` v0.1.0 and newer `modernc.org/cc*`/`libc` versions; all remain pure Go.
- Plugin ABI, cache layout, routes and status vocabulary are unchanged.
- Lua 5.1-only globals (`unpack`, `loadstring`, `setfenv`) are gone; the shipped corpus does not use them.
- A conformance probe runs every shipped Lua plugin through search/detail/chapters/pages before archive.

# Related Concepts
- [Dependency stack: pure Go only (modernc sqlite, tls-client, modernc lua, quickjs, wazero)](dependency-stack.md): Updates the Lua slot in the frozen dependency stack.
- [Stack: libraries vs hand-written code](../architecture/stack.md): Runtime change reflected in the stack review.
- [No published modernc.org/wasm3 module; keep wazero](../notes/modernc-wasm3-unavailable.md): Contrasts the available modernc Lua module with the unavailable modernc wasm3 module.
