## 2026-11-21
* **Update**: Linked `notes/jev-systemone-kit.md` to `references/kit-worker.md` (records both the native decision wire and the opt-in MCP route in the KIT launcher).
* **Update**: Updated concept `notes/jev-systemone-kit.md`.
* **Update**: Linked `notes/jev-systemone-kit.md` to `references/kit-worker.md` (records the opt-in jev MCP tool added to the KIT worker harness).
* **Creation**: Documented concept `notes/jev-systemone-kit.md` (systemone/jev is a KIT MCP tool, not a KIT model).
* **Update**: Linked `notes/probe-2026-11-21.md` to `process/porting-order.md` (Probe snapshot verifying the tree is green at the current porting stage).
* **Creation**: Documented concept `notes/probe-2026-11-21.md` (Probe 2026-11-21: tree builds and tests green).
* **Update**: Linked `notes/fts5-purego-probe.md` to `architecture/concurrency-ownership.md` (FTS5 adds no worker-pool requirement; the policy is recorded separately).
* **Update**: Linked `decisions/concurrency-pool-policy.md` to `decisions/state-machine-engine.md` (plugin runtime slots map to Enabled/Idle/Busy states, not to generic workers).
* **Update**: Linked `decisions/concurrency-pool-policy.md` to `decisions/store-access.md` (GORM's connection pool replaces the baseline db_pool; no work pool is added).
* **Update**: Linked `decisions/concurrency-pool-policy.md` to `architecture/concurrency-ownership.md` (locks the pool policy already implied by the ownership and backpressure tables).
* **Creation**: Documented concept `decisions/concurrency-pool-policy.md` (No generic worker pool: job runners plus semaphores).
* **Update**: Linked `notes/fts5-purego-probe.md` to `decisions/store-access.md` (proves FTS5 and BM25 work on the chosen pure-Go GORM stack).
* **Update**: Updated concept `notes/fts5-purego-probe.md`.
* **Update**: Linked `notes/fts5-purego-probe.md` to `decisions/store-access.md` (proves the chosen pure-Go GORM stack supports the FTS5 feature the baseline relies on).
* **Creation**: Documented concept `notes/fts5-purego-probe.md` (FTS5 works on the pure-Go GORM stack, but V2 specs omit it).
* **Update**: Updated concept `conventions/sql-boundary.md`.
* **Update**: Updated concept `decisions/store-access.md`.
* **Update**: Linked `conventions/sql-boundary.md` to `architecture/stack.md` (the review page points at the gate that keeps store SQL in one place).
* **Update**: Linked `decisions/store-access.md` to `conventions/sql-boundary.md` (the decision's enforcement mechanism is the raw-SQL boundary rule).
* **Update**: Linked `decisions/store-access.md` to `decisions/dependency-stack.md` (extends the dependency stack with GORM on the pure-Go SQLite driver).
* **Creation**: Documented concept `conventions/sql-boundary.md` (Raw SQL boundary).
* **Creation**: Documented concept `decisions/store-access.md` (Store access: GORM with one raw-SQL boundary).
* **Update**: Linked `notes/config-resolution-gotchas.md` to `notes/go-ini-v2-behaviour.md` (both record verified behaviour of the config layer built for add-config).
* **Update**: Linked `notes/config-resolution-gotchas.md` to `decisions/config-file-format.md` (records implementation traps found while applying the config-file-format decision).
* **Update**: Linked `notes/go-ini-v2-behaviour.md` to `decisions/config-file-format.md` (records the library behaviour the config-file-format decision depends on).
* **Creation**: Documented concept `notes/config-resolution-gotchas.md` (Config resolution gotchas found while building add-config).
* **Creation**: Documented concept `notes/go-ini-v2-behaviour.md` (unknwon.dev/ini/v2 round-trip behaviour).
* **Update**: Linked `decisions/config-file-format.md` to `decisions/dependency-stack.md` (adds gopkg.in/ini.v1 to the approved pure-Go dependency set).
* **Update**: Linked `decisions/config-file-format.md` to `architecture/stack.md` (supplies the ini parser row in the stack review and replaces the hand-written INI plan).
* **Creation**: Documented concept `conventions/language.md` (English for all artifacts, Indonesian only in chat).
* **Update**: Updated concept `references/kit-worker.md`.
* **Update**: Linked `conventions/godoc.md` to `architecture/overview.md` (applies to every Go package described by the layer map).
* **Creation**: Documented concept `conventions/godoc.md` (Godoc comment rule for all Go source).
* **Update**: Linked `references/kit-worker.md` to `process/porting-order.md` (worker lanes accelerate the capability implementation sequence).
* **Creation**: Documented concept `references/kit-worker.md` (KIT headless worker harness (.kit.yml + scripts/kit-worker + kit-worker skill)).
* **Update**: Linked `process/porting-order.md` to `architecture/overview.md` (orders the implementation of the layers described here).
* **Update**: Linked `process/porting-order.md` to `architecture/capability-map.md` (defines the execution order and per-change gates for the capabilities mapped here).
* **Creation**: Documented concept `process/porting-order.md` (Capability porting order: one OpenSpec change per capability, config first).
* **Update**: Linked `decisions/tray.md` to `architecture/stack.md` (resolves the tray slot listed as open on the stack page).
* **Creation**: Documented concept `decisions/tray.md` (Tray as a separate pure-Go command on fyne.io/systray).
* **Update**: Linked `architecture/stack.md` to `decisions/event-dispatch.md` (event bus and stream hub rows point at the dispatch decision).
* **Update**: Linked `architecture/stack.md` to `decisions/image-pipeline.md` (image codecs are listed as library rows on the stack page).
* **Update**: Linked `architecture/stack.md` to `decisions/dependency-stack.md` (stack page is the review index over the per-area dependency decisions).
* **Update**: Linked `decisions/http-framework.md` to `architecture/stack.md` (names the router and websocket libraries on the stack page).
* **Creation**: Documented concept `architecture/stack.md` (Stack: libraries vs hand-written code).
* **Creation**: Documented concept `decisions/http-framework.md` (chi + coder/websocket for HTTP and WebSocket).
* **Update**: Linked `decisions/browser-host-transport.md` to `architecture/overview.md` (browser backend and its supervision are part of the plugin layer described in the overview).
* **Creation**: Documented concept `knowledge/decisions/browser-host-transport.md` (Browser host transport and supervision in Go).
* **Update**: Linked `architecture/capability-map.md` to `architecture/overview.md` (the capability layer split that mirrors the architecture layers).
* **Creation**: Documented concept `architecture/capability-map.md` (Capability map: 25 baseline specs collapsed into 11 V2 capabilities with a porting order).
* **Update**: Linked `decisions/dependency-stack.md` to `decisions/image-pipeline.md` (applies the no-cgo rule to the image pipeline).
* **Update**: Linked `architecture/cache-layout.md` to `decisions/image-pipeline.md` (implements the transcode step with cgo-free codecs).
* **Update**: Linked `architecture/data-flows.md` to `architecture/cache-layout.md` (the image-serve flow resolves against this tree).
* **Update**: Linked `architecture/overview.md` to `architecture/cache-layout.md` (expands the cache tree entry of the frozen contracts).
* **Creation**: Documented concept `decisions/image-pipeline.md` (Image transcode without libvips: gen2brain WASM-transpiled webp/avif/jxl, pure-Go resize).
* **Creation**: Documented concept `architecture/cache-layout.md` (Image cache: v3 layout, namespace by display name, slug rules, LRU and pure-Go transcode).
* **Update**: Linked `architecture/state-machines.md` to `architecture/data-flows.md` (the plugin toggle flow is the FSM transition path).
* **Update**: Linked `architecture/event-driven-core.md` to `architecture/data-flows.md` (defines where each flow publishes and why nothing publishes inside a transaction).
* **Update**: Linked `architecture/overview.md` to `architecture/boot-shutdown.md` (startup and teardown order for the layer map).
* **Update**: Linked `architecture/overview.md` to `architecture/data-flows.md` (each request path in the overview is expanded here into a full ordered sequence).
* **Creation**: Documented concept `architecture/boot-shutdown.md` (Boot and shutdown: ordered startup, dependency-ordered teardown, and the test that proves both).
* **Creation**: Documented concept `architecture/data-flows.md` (Data flows: five ordered sequences and the invariants that keep them reproducible).
* **Update**: Linked `architecture/event-driven-core.md` to `decisions/dependency-stack.md` (fixes the event bus implementation behind the actor wrapper).
* **Update**: Linked `architecture/overview.md` to `decisions/dependency-stack.md` (names the concrete pure-Go libraries behind each stack slot).
* **Creation**: Documented concept `decisions/dependency-stack.md` (Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, goja, wazero)).
* **Update**: Linked `architecture/concurrency-ownership.md` to `architecture/event-driven-core.md` (the bus actor is one of the single owners in this model).
* **Update**: Linked `architecture/overview.md` to `architecture/state-machines.md` (plugin and job lifecycles live in the domain service layer).
* **Update**: Linked `architecture/overview.md` to `architecture/event-driven-core.md` (transport and services publish domain facts through the bus).
* **Update**: Linked `architecture/overview.md` to `architecture/concurrency-ownership.md` (layer boundaries are enforced by the single-owner concurrency model).
* **Creation**: Documented concept `architecture/concurrency-ownership.md` (Concurrency and ownership: one owner per mutable resource, bounded everything).
* **Creation**: Documented concept `architecture/overview.md` (Architecture overview: layers, import rules and frozen external contracts).
* **Update**: Linked `architecture/event-driven-core.md` to `architecture/state-machines.md` (domain events originate from persisted FSM transitions).
* **Update**: Linked `architecture/state-machines.md` to `decisions/state-machine-engine.md` (implements the engine choice and wrapper rules).
* **Update**: Linked `architecture/state-machines.md` to `architecture/event-driven-core.md` (plugin and job transitions are the source of domain events; persist-then-publish ordering).
* **Creation**: Documented concept `decisions/state-machine-engine.md` (Use qmuntal/stateless as the FSM engine behind typed domain wrappers).
* **Creation**: Documented concept `architecture/state-machines.md` (State machines: plugin and job lifecycle on qmuntal/stateless; cache entry and daemon deliberately not).
* **Update**: Updated concept `decisions/event-dispatch.md`.
* **Update**: Linked `architecture/event-driven-core.md` to `decisions/event-dispatch.md`.
* **Creation**: Documented concept `decisions/event-dispatch.md` (Adopt gookit/event behind a single-actor bus; separate stream plane).
* **Creation**: Documented concept `architecture/event-driven-core.md` (Event-driven core: domain events on gookit/event, stream plane separate).
* **Creation**: Documented concept `references/reference-sources.md` (Reference source map (cppIsekai baseline, goIsekai V1 corpus)).
* **Creation**: Documented concept `decisions/goisekai-v2-bootstrap.md` (goIsekai2 is a greenfield Go rewrite against the cppIsekai baseline).
* **Creation**: Initialized OKF v0.2 knowledge bundle.
