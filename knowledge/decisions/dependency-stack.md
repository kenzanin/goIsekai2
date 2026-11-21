---
type: decision
title: "Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, quickjs, wazero)"
description: "No-cgo constraint with CGO_ENABLED=0 as build gate; chosen modernc.org/sqlite v1.62.0, bogdanfinn/tls-client v1.18.1, yuin/gopher-lua v1.1.2, modernc.org/quickjs v0.27.0 (V1 evidence supersedes goja) and tetratelabs/wazero v1.12.0; rejection of buke/quickjs-go, wasm3 and the purego native-lib route."
tags: [decision, dependencies, no-cgo, sqlite, tls, lua, js, quickjs, wasm]
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
| JS | `modernc.org/quickjs` v0.27.0 | pure Go (transpiled C QuickJS on `modernc.org/libc`), ES2023, host funcs, job pump, timeout/interrupt; V1 builds with `CGO_ENABLED=0`; supersedes goja, see the revision section |
| WASM | `github.com/tetratelabs/wazero` v1.12.0 | pure Go (Apache-2.0), compiler + interpreter, WASI preview1, active |

## goja probe: why it was the first choice, and why it lost

When the stack was first surveyed, goja looked like the obvious pick (pure Go, MIT, V1 was believed to run
its JS plugins on it). Probe run against current goja (`go run` with the real module, not a doc claim):

| Feature | Result |
|---|---|
| optional chaining `?.` | PASS (executed) |
| nullish `??` | PASS (executed) |
| BigInt | PASS (executed) |
| spread, template literals, classes | PASS (executed) |
| `Promise` object | exists |
| `async`/`await` settling | does **not** settle: promise jobs are not pumped, so `r` stays `"pending"` |

The last row is an engine limit on top of the ABI limit: goja does not pump promise jobs unless the host
calls its job queue explicitly. Both JS plugins in the V1 corpus are plain ES5, so goja would have worked,
but it caps the plugin language at ES5-plus-syntax forever.

## wazero over wasm3

wasm3 is a C interpreter (MIT, active) and needs a cgo binding, which the hard constraint forbids. wazero is
pure Go, ships both a compiler (amd64/arm64) and an interpreter, supports WASI preview1, and is the only
runtime that satisfies the constraint. Startup cost is acceptable because plugin calls are long-lived
instances, not per-request processes.

## Revision 2026-11-21: V1 evidence selects `modernc.org/quickjs`

Earlier in this concept the JS slot was goja, with QuickJS reachable only through a WASI build or purego.
That was wrong about V1: V1 has no goja at all, it runs `modernc.org/quickjs` v0.27.0 -- C QuickJS
transpiled to Go through `modernc.org/libc`, so the engine is pure Go with no `.wasm` build step and no
native library. Verified, not assumed:

```bash
cd ~/tmp/goIsekai
grep -n goja go.mod                                        # no match
CGO_ENABLED=0 go build -o /tmp/gk-server ./cmd/goisekai    # exit 0, links internal/pluginmanager
```

| API (from the module source) | What it gives us |
|---|---|
| `NewVM()`, `Eval(src, EvalGlobal)` | ES2023 script evaluation |
| `HostFunc func(args []any) (any, error)` | the host-tool bridge, same shape as our plugin ABI |
| `ExecutePendingJobs()`, `promise.go` | a real job pump exists, so async is a host decision, not an engine wall |
| `SetEvalTimeout(d)`, `Interrupt()` | per-call timeout and kill for a runaway plugin |
| `SetMemoryLimit`, `SetMaxStackSize`, `Close()` | sandbox bounds and clean teardown |

Lua stays gopher-lua (5.1). wazero stays for wasm plugins; it is no longer needed as a QuickJS transport.

goja remains the fallback only if `modernc.org/quickjs` ever fails a gate (it pulls `modernc.org/libc`,
which is the same family already used by `modernc.org/sqlite`, so the dependency cost is shared).

### Superseded history: cgo-free QuickJS paths considered before the finding

| Path | Binary cgo-free | Extra runtime artifacts | Binding work | Verdict |
|---|---|---|---|---|
| quickjs-ng -> WASI reactor -> wazero | yes | one `.wasm` per build, all platforms | JSON-in/JSON-out host calls | superseded by `modernc.org/quickjs` |
| `purego` + `libquickjs.so` | yes | native lib **per GOOS/GOARCH** (linux/macOS/windows, amd64+arm64) | ~1k LOC unsafe ABI, manual `JS_DupValue`/`JS_FreeValue` | superseded; purego is kept for native plugin loading only |
| `buke/quickjs-go` (cgo) | no | linked in | already written | rejected by the no-cgo rule |

### purego assessment (still relevant for native plugin libs)

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

### WASI route detail (superseded)

quickjs-ng ships `qjs-wasi-reactor.c` but does not wire it into its Makefile, so the route would need a
`wasi-sdk`-pinned build step plus a checked-in `.wasm` artifact. Not needed now that `modernc.org/quickjs`
exists; keep the note only as a fallback if the transpiled engine is ever abandoned upstream.

## Consequences

1. Driver choice means `database/sql` with one connection; see `architecture/concurrency-ownership`.
2. gopher-lua is Lua 5.1, not LuaJIT: `ffi` and `bit` are unavailable. A plugin using them must fail with a
   clear error, not silently misbehave.
3. wazero is only the wasm plugin runtime; the QuickJS engine is the transpiled `modernc.org/quickjs`, so
   there is no `.wasm` build step in the release pipeline.
4. Any new dependency must prove `CGO_ENABLED=0` compatibility before its first import.
5. `modernc.org/libc` is now pulled twice (sqlite and quickjs) and deduplicated as one module version.

# Related Concepts
- [Image transcode without libvips: gen2brain WASM-transpiled webp/avif/jxl, pure-Go resize](image-pipeline.md): applies the no-cgo rule to the image pipeline
