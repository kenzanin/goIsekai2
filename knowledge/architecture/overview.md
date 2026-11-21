---
type: architecture
title: "Architecture overview: layers, import rules and frozen external contracts"
description: "goIsekai2 shape as one local daemon: layer map and import matrix, model as the only shared package, store as the only SQL speaker, request paths for chapters and images, frozen baselines (routes, plugin ABI, status vocabulary, cache tree, config precedence), data directory and the open stack decisions."
tags: [architecture, layers, dependencies, http, plugins, sqlite, stack]
generated: { by: agent/cli, at: "2026-11-21T12:19:11Z" }
status: draft
---

## Shape

One binary, one data directory, one SQLite file, plugin folder per site. No external services: the daemon
serves SSR pages plus a small JSON API, scrapes through plugins, caches images on disk, and persists
everything locally. Porting target is cppIsekai, so every externally visible contract (routes, payloads,
status vocabulary, cache paths, plugin ABI) is fixed by the baseline; only internals are ours.

```mermaid
graph TD
  cmd[cmd/goisekai2] --> http[internal/http]
  http --> library[internal/library]
  http --> reader[internal/reader]
  http --> api[internal/http/api]
  library --> store[internal/store]
  library --> plugin[internal/plugin]
  reader --> store
  reader --> cache[internal/cache]
  api --> jobs[internal/jobs]
  jobs --> store
  jobs --> plugin
  plugin --> hosttool[internal/plugin/hosttool]
  hosttool --> fetch[internal/fetch]
  hosttool --> store
  hosttool --> cache
  http --> events[internal/events]
  http --> stream[internal/stream]
  jobs --> events
  plugin --> events
  plugin --> fsm[internal/fsm]
  jobs --> fsm
  events --> stream
  store --> sqlite[(SQLite)]
  cache --> disk[(data/cache)]
```

## Layers and import rules

| Layer | Package | May import | Must not import |
|---|---|---|---|
| Entry | `cmd/goisekai2` | everything (wiring only) | -- |
| Transport | `internal/http` | library, reader, jobs, config, events, stream, model | plugin internals, sqlite driver |
| Domain services | `internal/library`, `internal/reader`, `internal/jobs` | store, plugin (hub interface), cache, events, fsm, model | http, templates |
| Plugin host | `internal/plugin` (hub, runtimes) | fsm, events, hosttool, model | store directly, http, cache internals |
| Host tools | `internal/plugin/hosttool` | fetch, store interface, cache interface, crypto/text helpers | http, templates, plugin runtimes |
| Infrastructure | `internal/store`, `internal/cache`, `internal/fetch` | model, config | any service, http, plugin |
| Leaves | `internal/events`, `internal/stream`, `internal/fsm`, `internal/model` | stdlib (+ `gookit/event`, `stateless` inside their wrappers) | everything else |

Hard rules:

1. `internal/model` holds shared structs and enums only; it imports stdlib and nothing from this repo. Every
   other package may import it, so it can never depend upward.
2. `internal/store` is the only package that speaks SQL. Plugins reach data through host tools
   (`storage.get/set`), never through the driver -- the baseline enforces the same rule by routing plugin
   storage through the db worker (`src/main.cpp:106`).
3. No `init()` side effects and no package-level mutable singletons. All wiring happens in `cmd`.
4. `events`, `stream` and `fsm` are leaves: importing them never drags in the store or HTTP.
5. Guest plugin code (Lua/JS/WASM) never imports Go. It only sees the host-tool table, payloads as JSON text.

## Request paths (baseline parity)

```mermaid
sequenceDiagram
  participant B as Browser
  participant H as http handler
  participant S as service
  participant P as plugin host
  participant C as cache
  participant D as SQLite
  B->>H: GET /api/chapters/:plugin/:slug
  H->>S: Chapters(ctx, key)
  S->>D: read chapters + pages (single connection)
  alt stale and online
    S->>P: call(chapters)
    P-->>S: JSON envelope
    S->>D: upsert, commit
    S->>H: fresh data
  end
  H-->>B: JSON
  B->>H: GET /cache/img?url=...
  H->>C: ensure(key)
  C-->>H: file (miss: fetch, transcode, sidecar)
  H-->>B: image bytes
```

Reads are store-first. Network work only happens on a miss, a stale hit or an explicit refresh, and every
write is committed before any event is published.

## External contracts (frozen by the baseline)

| Contract | Source |
|---|---|
| Routes and payload shapes | `src/http/*`, `openspec/specs/*` in cppIsekai |
| Plugin ABI, host tools, method vocabulary | `src/plugin/*`, `knowledge/architecture/plugin-system.md` |
| Status vocabulary | `Ongoing|Completed|Hiatus|Dropped|Upcoming`, tolerate upstream `Ongoin` |
| Cache tree | `img/<plugin-ns>/cover/<sha1-16>.<ext>`, `img/<plugin-ns>/manga/<slug-dir>/ch<n>-<id8>/pNNN.<ext>` (see `architecture/cache-layout`) |
| Config knobs and precedence | `src/config/registry.hpp`: default < ini < env < flag, `--dump-config` |
| Reader window | limit 10, strictly-after position, per-chapter slices |

Everything else -- package layout, event bus, FSM, worker model -- is internal and free to diverge, provided
the tables above do not move.

## Data directory

```
data/
  library.db          SQLite (WAL), backups as library-YYYYmmdd-HHMMSS.db
  cache/img/...       v3 tree, LRU over bytes
  plugins/<site>/     main.json + entry (lua|js|wasm)
  enrichment/<plug>/  optional enrichment plugins
  logs/goisekai.log   append-only
```

## Stack decisions

| Area | Choice | Status |
|---|---|---|
| HTTP | `go-chi/chi/v5` v5.5.0 + templ v0.3.1118 + Tailwind v4.3.5 + Alpine/htmx | decided (`decisions/ui-stack`) |
| Events | `gookit/event` behind one actor | decided (`decisions/event-dispatch`) |
| FSM | `qmuntal/stateless` behind typed wrappers | decided (`decisions/state-machine-engine`) |
| SQLite driver | `modernc.org/sqlite` v1.62.0 (pure Go, BSD-3) | decided |
| HTTP client | `bogdanfinn/tls-client` v1.18.1 (pure Go via utls, BSD-3) | decided |
| Lua runtime | `github.com/yuin/gopher-lua` v1.1.2 (pure Go, Lua 5.1) | decided |
| JS runtime | `modernc.org/quickjs` v0.27.0 (pure Go, ES2023, V1-proven) | decided (`decisions/dependency-stack`) |
| WASM runtime | `github.com/tetratelabs/wazero` v1.12.0 (pure Go, Apache-2.0) | decided |
| Real-time | `coder/websocket` v1.8.15 over `internal/stream` (drop-oldest) | decided: WS `GET /ws/log` parity, not SSE |
| Tray | `fyne.io/systray` v1.11.0 in a separate tray command | decided (`decisions/tray`) |

Hard constraint: **no cgo anywhere**. The build gate is `CGO_ENABLED=0 go build ./...`, so any dependency
that needs a C toolchain is rejected outright -- that excludes `mattn/go-sqlite3`, `buke/quickjs-go` (cgo,
45 vendored C files, self-declared not production ready) and `wasm3` (C). Rationale and the QuickJS-on-WASI
escape hatch are in `decisions/dependency-stack`.

Remaining compatibility gates (not blockers, but they decide behaviour):

1. **gopher-lua is Lua 5.1, not LuaJIT.** The baseline embeds LuaJIT; 5.1 semantics match, LuaJIT extras (`ffi`,
   `bit`) do not. Any plugin needing them is a runtime bug we must surface explicitly.
2. **JS engine is ES2023 (`modernc.org/quickjs`), but the ABI stays synchronous.** The engine has a real job
   pump (`ExecutePendingJobs`), `SetEvalTimeout` and `Interrupt`, so async is a host decision, not a wall. The
   plugin contract is still `call(method, json) -> json`; a plugin that returns a `Promise` is an error unless
   the host later adds a settlement step.

## Known divergences to record

1. **Realtime transport** (resolved): WebSocket `GET /ws/log` with snapshot replay is kept for parity, served by `coder/websocket` over the drop-oldest stream hub; no SSE route. See `decisions/http-framework`.
2. **Job table**: added for durable cancel/retry (baseline keeps retry in `update_attempts` only).
3. **Plugin state columns**: added for FSM persistence; API still exposes boolean `enabled`.

# Related Concepts
- [Concurrency and ownership: one owner per mutable resource, bounded everything](concurrency-ownership.md): layer boundaries are enforced by the single-owner concurrency model
- [Event-driven core: domain events on gookit/event, stream plane separate](event-driven-core.md): transport and services publish domain facts through the bus
- [State machines: plugin and job lifecycle on qmuntal/stateless; cache entry and daemon deliberately not](state-machines.md): plugin and job lifecycles live in the domain service layer
- [Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, goja, wazero)](../decisions/dependency-stack.md): names the concrete pure-Go libraries behind each stack slot
- [chi + coder/websocket for HTTP and WebSocket](../decisions/http-framework.md): router and WebSocket transport for the HTTP layer
- [Stack: libraries vs hand-written code](stack.md): one-page review table of every library row and every hand-written piece
- [Data flows: five ordered sequences and the invariants that keep them reproducible](data-flows.md): each request path in the overview is expanded here into a full ordered sequence
- [Boot and shutdown: ordered startup, dependency-ordered teardown, and the test that proves both](boot-shutdown.md): startup and teardown order for the layer map
- [Image cache: v3 layout, namespace by display name, slug rules, LRU and pure-Go transcode](cache-layout.md): expands the cache tree entry of the frozen contracts
