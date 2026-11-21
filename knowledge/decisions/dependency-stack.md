---
type: decision
title: "Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, goja, wazero)"
description: "No-cgo constraint with CGO_ENABLED=0 as build gate; chosen modernc.org/sqlite v1.62.0, bogdanfinn/tls-client v1.18.1, yuin/gopher-lua v1.1.2, dop251/goja and tetratelabs/wazero v1.12.0; measured goja ES2020 probe results, rejection of buke/quickjs-go and wasm3, QuickJS-on-WASI escape hatch."
tags: [decision, dependencies, no-cgo, sqlite, tls, lua, js, wasm]
generated: { by: agent/cli, at: "2026-11-21T12:19:11Z" }
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

## QuickJS paths (none adopted)

Three ways to get QuickJS semantics without cgo. Ranked by what they cost to own:

| Path | Binary cgo-free | Extra runtime artifacts | Binding work | Verdict |
|---|---|---|---|---|
| quickjs-ng -> WASI reactor -> wazero | yes | one `.wasm` per build, all platforms | JSON-in/JSON-out host calls | **preferred if parity is ever needed** |
| `purego` + `libquickjs.so` | yes | native lib **per GOOS/GOARCH** (linux/macOS/windows, amd64+arm64) | ~1k LOC unsafe ABI, manual `JS_DupValue`/`JS_FreeValue` | viable, worst to operate |
| `buke/quickjs-go` (cgo) | no | linked in | already written | rejected by the no-cgo rule |

### purego assessment (asked explicitly)

It genuinely works: `ebitengine/purego` v0.11.3 (Apache-2.0, active) can `dlopen` a shared library and call
C from a `CGO_ENABLED=0` binary, and Linux amd64/arm64 are Tier 1 with struct-by-value arguments **and
returns** supported -- which matters because QuickJS returns `JSValue` (16 bytes) by value. It is still
beta software.

What it costs us: the Go binary stays cgo-free, but the *product* stops being pure Go. We would ship
`libquickjs.so` / `.dylib` / `.dll` for every target, which reintroduces per-OS release artifacts, dynamic
loader paths and a C toolchain in the release pipeline -- exactly the deployment class the no-cgo rule
exists to avoid. On top of that we would hand-maintain ~1k LOC of `unsafe` binding with manual refcounting,
where a single missed `JS_FreeValue` is a leak and a double free is corruption, plus callback pinning for the
interrupt handler. The WASI route reaches the same semantics with one portable artifact and no native
loader, so purego is not the shortcut it looks like.

### WASI route detail

quickjs-ng ships `qjs-wasi-reactor.c` but does not wire it into its Makefile, so we would own a
`wasi-sdk`-pinned build step plus a checked-in `.wasm` artifact. Revisit only when a plugin actually needs
ES2020 semantics beyond goja.

## Consequences

1. Driver choice means `database/sql` with one connection; see `architecture/concurrency-ownership`.
2. gopher-lua is Lua 5.1, not LuaJIT: `ffi` and `bit` are unavailable. A plugin using them must fail with a
   clear error, not silently misbehave.
3. The wasm runtime is the same engine that would host the QuickJS escape hatch, so the investment is shared.
4. Any new dependency must prove `CGO_ENABLED=0` compatibility before its first import.

# Related Concepts
- [Image transcode without libvips: gen2brain WASM-transpiled webp/avif/jxl, pure-Go resize](image-pipeline.md): applies the no-cgo rule to the image pipeline
