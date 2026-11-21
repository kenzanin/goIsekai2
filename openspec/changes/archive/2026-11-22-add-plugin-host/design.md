# Design: add-plugin-host

## Context

See `proposal.md`. The `plugin-host` capability spec freezes the ABI, method vocabulary and visible
JSON. cppIsekai `src/plugin*` is the behavioral baseline; V2 diverges internally with an explicit FSM
and typed events, but keeps the visible contract. Host tools (network/storage/log/crypto/text) are a
separate capability and are only referenced through an interface here.

## Goals / Non-Goals

- Goals: deterministic discovery, one runtime interface, no cgo, no plugin access to store or event
  bus, single call per instance, boot/reload parity, typed management page using the archived UI
  stack.
- Non-Goals: implementing the actual host tools, scraping site logic, library/reader routes.

## Decisions

### D1 — Manifest and discovery

One folder per plugin under `paths.plugins`; `main.json` declares `name`, `version`, `runtime`,
`capabilities`, `api` and optional entry/logo. Folder name is the stable id. Broken manifests are
skipped with a warning; duplicate ids resolve by scan order. Baseline: cppIsekai plugin folders and
`main.json`.

Alternative: a single registry file — rejected because the baseline folder layout is frozen.

### D2 — One runtime interface, three backends

`Runtime.Call(ctx, envelope) ([]byte, error)` is the only path to plugin code. Lua uses
`gopher-lua` (Lua 5.1, `ffi`/`bit` fail clearly), JS uses `modernc.org/quickjs` with
`ExecutePendingJobs`, WASM uses `wazero` with WASI preview1 and exported `handle`. A Promise return
is an error because the ABI is synchronous. Baseline: V1 already proved these three engines under
`CGO_ENABLED=0`.

Alternative: quickjs via cgo/purego — rejected by the no-cgo gate.

### D3 — Lifecycle FSM

`qmuntal/stateless` v1.8.0 drives `Disabled`, `Enabled{Idle,Busy}`, `Failed`. One owning actor fires
triggers; callbacks only persist/enqueue. Every transition writes the plugin row, commits, then
publishes `plugin.lifecycle`. Illegal triggers return a typed error and increment a counter, never
panic. Baseline: cppIsekai's per-runtime mutex and implicit state; V2 makes it explicit.

Alternative: hand-rolled table — rejected because future engine features (guards, hierarchy,
external storage) would have to be written by us.

### D4 — Single call per instance

A per-plugin semaphore/FIFO queue admits one guest call at a time; queued callers honour their own
context deadlines. Reload waits for the in-flight call. Baseline: cppIsekai's "one call at a time"
per-runtime mutex and `plugin_pool(10)`.

Alternative: a shared worker pool — rejected by the concurrency-pool policy.

### D5 — Budget and sandbox

Lua gets an instruction budget, JS an interrupt deadline, WASM a context deadline plus fuel. Lua
exposes only base/string/table/math/json and read-only io; `os`/`debug` are absent. JS has no FS,
network or process globals. WASM gets WASI preview1 only. No runtime receives the store handle or
the event bus. Baseline: cppIsekai host-tool surface, inverted so plugins see only declared tools.

Alternative: OS-level sandbox — rejected as cgo/platform-specific.

### D6 — Host tool interface

The runtime receives a `HostTools` interface (`http.fetch`, `storage.*`, `log.*`, `crypto.*`,
`text.*`) implemented in a later capability. This change defines the interface and injects a test
double only; it never opens the database from plugin code. Baseline: cppIsekai `Host::call`.

### D7 — Boot, reload and purge

Boot order: store migrations → plugin discovery → lifecycle reconciliation → HTTP listen. Reload
reuses the same pipeline. A scan that finds no directory logs an error and short-circuits before
pruning; a missing folder for a known id purges its data rows in one transaction, plugin row last.
Baseline: cppIsekai `main.cpp` boot ladder and folder pruning.

### D8 — Management page

The `/plugins` page is a templ component using the archived Tailwind/Alpine/htmx stack, rendering
rows with status/runtime/capabilities/title count and enable/disable/reload actions. Baseline:
cppIsekai plugins page; the visible JSON keeps the boolean `enabled` field.

## Risks / Trade-offs

- [plugin code can hang] → per-runtime budget plus context deadline; never an unbounded call.
- [FSM external state diverges from disk] → persist before publish; boot reconciliation is authoritative.
- [runtime error leaks host paths] → logs truncate raw output and name only plugin id + method.
- [reload races a call] → reload waits for the current call; old code serves the in-flight request.

## Migration Plan

Greenfield: no plugins shipped yet. Add a Lua fixture plugin in tests only. Rollback is reverting
the change commit; no production data is touched.

## Open Questions

None. ABI, runtimes and lifecycle rules are fixed by the capability spec.
