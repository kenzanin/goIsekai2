---
type: decision
title: "Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, goja, wazero)"
description: "No-cgo constraint with CGO_ENABLED=0 as build gate; chosen modernc.org/sqlite v1.62.0, bogdanfinn/tls-client v1.18.1, yuin/gopher-lua v1.1.2, dop251/goja and tetratelabs/wazero v1.12.0; measured goja ES2020 probe results, rejection of buke/quickjs-go and wasm3, QuickJS-on-WASI escape hatch."
tags: [decision, dependencies, no-cgo, sqlite, tls, lua, js, wasm]
generated: { by: agent/mcp, at: "2026-11-21T12:11:06Z" }
status: stable
---

## Hard constraint

**No cgo.** The build gate is `CGO_ENABLED=0 go build ./...` and `CGO_ENABLED=0 go test ./...`. Any
dependency that needs a C toolchain is rejected, no exceptions. This also removes the musl/glibc deployment
class of problems and keeps cross-compilation trivial.

## Decisions

| Area | Decision | Evidence |
|---|---|---|
| SQLite | `modernc.org/sqlite` v1.62.0 | pure Go (BSD-3), active (2026-11), used through `database/sql`; no cgo |
| HTTP client | `bogdanfinn/tls-client` v1.18.1 | pure Go via `bogdanfinn/utls` (BSD-3), active, provides curl-impersonate-class TLS fingerprints; stdlib cannot match site fingerprints |
| Lua | `github.com/yuin/gopher-lua` v1.1.2 | pure Go (MIT), active, Lua 5.1 |
| JS | `github.com/dop251/goja` | pure Go (MIT), active; V1 already ran its JS plugins on goja |
| WASM | `github.com/tetratelabs/wazero` v1.12.0 | pure Go (Apache-2.0), compiler + interpreter, WASI preview1, active |

## goja: measured, not assumed

Probe run against current goja (`go run` with the real module, not a doc claim):

| Feature | Result |
|---|---|
| optional chaining `?.` | PASS (executed) |
| nullish `??` | PASS (executed) |
| BigInt | PASS (executed) |
| spread, template literals, classes | PASS (executed) |
| `Promise` object | exists |
| `async`/`await` settling | does **not** settle: promise jobs are not pumped, so `r` stays `"pending"` |

The last row is not an engine verdict, it is an ABI verdict: the plugin contract is synchronous
`call(method, paramsJSON) -> JSON text`, so an async plugin cannot produce a return value under **any**
runtime, QuickJS included. Accepting async plugins is a host change (job pump + settlement deadline), not a
runtime swap. Both JS plugins in the V1 corpus are plain ES5.

## wazero over wasm3

wasm3 is a C interpreter (MIT, active) and needs a cgo binding, which the hard constraint forbids. wazero is
pure Go, ships both a compiler (amd64/arm64) and an interpreter, supports WASI preview1, and is the only
runtime that satisfies the constraint. Startup cost is acceptable because plugin calls are long-lived
instances, not per-request processes.

## QuickJS escape hatch (not adopted)

`buke/quickjs-go` is rejected: it imports C, vendors 45 C/H files, and its own README says it is "not ready
for production use". If real QuickJS semantics are ever required, the cgo-free route is compiling
quickjs-ng to a WASI reactor and running it under wazero: quickjs-ng ships `qjs-wasi-reactor.c` but does not
wire it into its Makefile, so we would own a `wasi-sdk`-pinned build step plus a checked-in `.wasm`
artifact. Revisit only when a plugin actually needs it.

## Consequences

1. Driver choice means `database/sql` with one connection; see `architecture/concurrency-ownership`.
2. gopher-lua is Lua 5.1, not LuaJIT: `ffi` and `bit` are unavailable. A plugin using them must fail with a
   clear error, not silently misbehave.
3. The wasm runtime is the same engine that would host the QuickJS escape hatch, so the investment is shared.
4. Any new dependency must prove `CGO_ENABLED=0` compatibility before its first import.
