---
type: architecture
title: "Stack: libraries vs hand-written code"
description: "One page answering what goIsekai2 takes from libraries and what it hand-writes, with the open slots where no decision exists yet."
tags: [stack, libraries, dependencies, handwritten, review]
generated: { by: agent/cli, at: "2026-11-21T12:35:23Z" }
status: stable
---

## Purpose

Review page. Every durable decision about a dependency lives in `decisions/*`; this concept is the single
table that answers "library or hand-written?" without reading five concepts.

## Libraries (all pure Go, no cgo, `CGO_ENABLED=0 go build ./...` is the gate)

| Area | Library | Version | Decision |
|---|---|---|---|
| SQLite | `modernc.org/sqlite` | v1.62.0 | `decisions/dependency-stack` |
| HTTP client (scraping) | `bogdanfinn/tls-client` | v1.18.1 | `decisions/dependency-stack` |
| Lua | `yuin/gopher-lua` | v1.1.2 | `decisions/dependency-stack` |
| JS | `modernc.org/quickjs` | v0.27.0 | `decisions/dependency-stack` |
| WASM | `tetratelabs/wazero` | v1.12.0 | `decisions/dependency-stack` |
| Router / framework | `go-chi/chi/v5` | v5.5.0 | `decisions/http-framework` |
| WebSocket | `coder/websocket` | v1.8.15 | `decisions/http-framework` |
| Event bus | `gookit/event` | v1.2.0 | `decisions/event-dispatch` |
| FSM engine | `qmuntal/stateless` | v1.8.0 | `decisions/state-machine-engine` |
| WebP encode | `gen2brain/webp` | v0.6.4 | `decisions/image-pipeline` |
| AVIF encode | `gen2brain/avif` | v0.6.0 | `decisions/image-pipeline` |
| JPEG XL encode | `gen2brain/jpegxl` | unpinned | `decisions/image-pipeline` |
| Decode / resize | `golang.org/x/image` | v0.48.0 | `decisions/image-pipeline` |
| Tray | `fyne.io/systray` | v1.11.0 | `decisions/tray` |
| INI config file | `gopkg.in/ini.v1` (repo `github.com/go-ini/ini`) | v1.68.0 | `decisions/config-file-format` |
| Native `.so` loading | `ebitengine/purego` | only if native plugins are adopted | `decisions/tray` |

Stdlib for everything else: `net/http`, `html/template`, `log/slog`, `encoding/json`, `crypto/*`,
`database/sql`, `os/exec`, `flag`.

## Hand-written (deliberate, no library)

| Piece | Why not a library |
|---|---|
| Config registry: flag + env + default + provenance + strict/lenient resolution | the baseline precedence rules are ours; the ini format is handled by `gopkg.in/ini.v1` (`decisions/config-file-format`) |
| Events facade + single actor over `gookit/event` | typed payloads are the point; the library is untyped `map[string]any` |
| Stream hub (bounded, drop-oldest) | nobody ships drop-oldest fan-out that never blocks the publisher |
| FSM layer: typed enums + `fire()` wrappers + SQLite external state | compile-time safety the engine cannot give (`State = any`) |
| Update scheduler, candidate selection, backoff, pacing, backup tick | product rules from `specs/updates`; the SQL window query is written by hand |
| Store: SQL, migrations, `schema_version`, anti-nested-Tx helper | no ORM by rule; V1 used `go-jet`, so this row is open for review (see below) |
| Plugin loader + 3-runtime bridge + host tools | ABI is frozen by the baseline (`specs/host-tools`, `specs/plugin-host`) |
| Image cache tree, slug, `cache_ns`, disk LRU, sidecar, GC | baseline cache layout is a frozen contract (`specs/image-cache`) |
| Daemon routes, SSR handlers, `/cache/img`, config save | thin layer over chi + `html/template` |
| Reader window math, prefetch, position | small, tested against the baseline maths |
| Browser host supervision (`os/exec`, single respawn, owned shutdown) | Go gives this directly; see `decisions/browser-host-transport` |
| Plugin scraping logic | no Go HTML parser: plugins use their own patterns/regex, as the baseline does |

## Open slots (no decision yet)

1. **`golang.org/x/sys`** for a single-instance file lock: probably needed, not yet recorded as an approved
   dependency (it arrives transitively through godbus anyway).
2. **SQL builder**: V1 used `go-jet/jet/v2`. V2's rule is hand-written SQL over `database/sql`; revisit only
   if the store package's tests start duplicating scan logic.

## Resolved since the first pass

1. **Tray.** `fyne.io/systray` v1.11.0 in a separate `cmd/goisekai-tray` command that spawns the daemon as a
   child. Pure Go on Linux (dbus StatusNotifierItem) and Windows; macOS is the only cgo surface and only for
   that binary. Evidence: `CGO_ENABLED=0 go build ./cmd/goisekai-tray` in V1 exits 0. The daemon never links
   a tray. `decisions/tray`
2. **JS engine.** `modernc.org/quickjs` v0.27.0, C QuickJS transpiled to Go: pure Go, ES2023, host funcs,
   `ExecutePendingJobs`, `SetEvalTimeout`, `Interrupt`, memory and stack limits. V1 ships it and builds with
   `CGO_ENABLED=0`, so the goja plan and the WASI build step are both dropped.
3. **Realtime transport.** `coder/websocket` over `internal/stream`, WS `GET /ws/log` parity, not SSE
   (`decisions/http-framework`).

## Rule

A new dependency is allowed only when this table gains a row and a `decisions/*` concept names the
alternative that was rejected. A hand-written piece is allowed only when its row explains why a library does
not fit.

# Related Concepts
- [Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, quickjs, wazero)](../decisions/dependency-stack.md): stack page is the review index over the per-area dependency decisions
- [Image transcode without libvips: gen2brain WASM-transpiled webp/avif/jxl, pure-Go resize](../decisions/image-pipeline.md): image codecs are listed as library rows on the stack page
- [Adopt gookit/event behind a single-actor bus; separate stream plane](../decisions/event-dispatch.md): event bus and stream hub rows point at the dispatch decision
- [chi + coder/websocket for HTTP and WebSocket](../decisions/http-framework.md): router and WebSocket transport rows
- [Tray as a separate pure-Go command on fyne.io/systray](../decisions/tray.md): tray and purego rows
