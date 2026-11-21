## 2026-11-21
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
