## 2026-11-21
* **Update**: Linked `notes/plugin-tools-crypto.md` to `notes/plugin-tools-surface.md` (crypto helpers are part of the same host tool surface; storage/http/text/regex/cookies facts hang off it too.).
* **Creation**: Documented concept `notes/plugin-tools-crypto.md` (tools crypto: lowercase-hex digests; Sign = HMAC-SHA256 alias; vectors openssl-verified).
* **Update**: Linked `notes/plugin-tools-regex.md` to `notes/plugin-tools-surface.md` (regex helpers extend the plugin host tool surface; same package and gate).
* **Creation**: Documented concept `notes/plugin-tools-regex.md` (tools regex: Capture falls back to whole match, Replace expands $1, no compile cache).
* **Update**: Linked `notes/plugin-tools-text.md` to `notes/plugin-tools-surface.md` (text helpers extend the host tool surface next to the ToolSet core.).
* **Creation**: Documented concept `notes/plugin-tools-text.md` (tools text helpers: Clean keeps newline/tab as whitespace, DateEpoch dual-layout, JSONBlob sorted-key stable).
* **Update**: Linked `notes/plugin-tools-cookies.md` to `notes/plugin-tools-surface.md` (Cookie persistence is the third leg of the same ToolSet host surface, beside storage and HTTP fetch.).
* **Creation**: Documented concept `notes/plugin-tools-cookies.md` (tools cookies: per-plugin TSV jar; SaveJar needs a LoadJar jar (stdlib jar has no enumeration)).
* **Update**: Linked `notes/tls-client-v1-api-shape.md` to `decisions/dependency-stack.md` (Pinpoints the exact NewHttpClient/option/fhttp call shape of the tls-client version recorded in the dependency decision.).
* **Update**: Linked `notes/plugin-tools-http.md` to `notes/plugin-tools-surface.md` (HTTP tools extend the same ToolSet surface as storage/log; same package, same gates.).
* **Creation**: Documented concept `notes/tls-client-v1-api-shape.md` (tls-client v1.18.1 API: functional options + fhttp types, DefaultClientProfile=Chrome_150).
* **Creation**: Documented concept `notes/plugin-tools-http.md` (tools HTTP surface: non-2xx is data, FetchBody errors on non-200, bare map means headers).
* **Update**: Linked `notes/plugin-tools-surface.md` to `notes/config-zero-value-string.md` (ToolSet.ConfigGet carries the explicit not-found branch because a zero config.Value renders "0".).
* **Update**: Linked `notes/plugin-tools-surface.md` to `architecture/overview.md` (Implements the host-tools slice of the plugin layer; ToolSet is the per-plugin capability bundle the runtimes bind.).
* **Creation**: Documented concept `notes/config-zero-value-string.md` (Zero config.Value renders "0", not "" — absent-knob paths must early-return).
* **Creation**: Documented concept `notes/plugin-tools-surface.md` (tools.ToolSet: storage delegates to plugin_state, missing keys return gorm not-found).
* **Update**: Linked `notes/go-mod-tidy-prunes-unimported-deps.md` to `decisions/dependency-stack.md` (dependency pin maintenance finding from the plugin runtime lanes).
* **Update**: Linked `notes/js-quickjs-thrown-string-errors.md` to `decisions/dependency-stack.md` (QuickJS error-shape finding from the plugin JS runtime built on the pinned pure-Go engine).
* **Update**: Updated concept `notes/plugin-envelope-v1.md`.
* **Creation**: Documented concept `notes/go-mod-tidy-prunes-unimported-deps.md` (go mod tidy prunes not-yet-imported deps; hand-edit for single promotions).
* **Creation**: Documented concept `notes/js-quickjs-thrown-string-errors.md` (QuickJS thrown-string errors keep exact text; match wrapped sentinels with Contains).
* **Update**: Linked `notes/plugin-envelope-v1.md` to `architecture/overview.md` (Implements the frozen plugin ABI baseline recorded in the architecture overview (envelope contract, method vocabulary).).
* **Creation**: Documented concept `notes/plugin-envelope-v1.md` (Plugin host contract v1 implemented and gated in internal/plugin).
* **Update**: Linked `notes/tmp-probe-go-test.md` to `notes/tmp-probe-go-build.md` (Same scratch package: the build probe proved compilation, this test probe proves the test gate too.).
* **Creation**: Documented concept `notes/tmp-probe-go-test.md` (tmp probe test green: go test ./tmp -count=1 exit 0, TestOK passes against OK() sentinel).
* **Update**: Linked `notes/tmp-probe-go-build.md` to `notes/tmp-dummy-probe.md` (Sibling scratch verification in tmp/: this probe extends the fixture chain with a compilable Go package sentinel.).
* **Creation**: Documented concept `notes/tmp-probe-go-build.md` (tmp/probe.go builds: package tmp with OK() sentinel, go build ./tmp exit 0).
* **Update**: Linked `notes/tmp-a-txt-fixture.md` to `notes/tmp-nested-b-fixture.md` (Sibling tmp/ fixture probes with identical verification pattern (exact payload, no trailing newline, exit 0).).
* **Creation**: Documented concept `notes/tmp-a-txt-fixture.md` (tmp/a.txt fixture: exact ALPHA payload, no newline; wc -c exits 0).
* **Update**: Linked `notes/tmp-nested-b-fixture.md` to `notes/tmp-dummy-probe.md` (Sibling scratch-probe fact: same tmp/ fixture pattern, byte-exact payload plus exit-code proof.).
* **Creation**: Documented concept `notes/tmp-nested-b-fixture.md` (tmp/nested/b.txt fixture: exact BETA payload, no newline; find exits 0).
* **Update**: Linked `notes/tmp-dummy-probe.md` to `references/kit-worker.md` (KIT_OK sentinel probe used to verify worker lane completion markers.).
* **Creation**: Documented concept `notes/tmp-dummy-probe.md` (tmp/dummy.txt sentinel probe: exact KIT_OK payload, ls exits 0).
* **Update**: Linked `notes/sqlite-pure-go-stack.md` to `notes/sqlite-pragma-return-values.md` (The pragma return-value gotcha was observed in the same Open verification that proved the pure-Go stack.).
* **Creation**: Documented concept `notes/sqlite-pragma-return-values.md` (SQLite pragma readback forms: lowercase wal, numeric foreign_keys).
* **Creation**: Documented concept `notes/sqlite-pure-go-stack.md` (Store DB: pure-Go gorm + glebarez/sqlite verified under CGO_ENABLED=0).
* **Update**: Updated concept `process/porting-order.md`.
* **Update**: Linked `decisions/ui-stack.md` to `architecture/stack.md` (UI stack rows belong in the consolidated dependency review).
* **Update**: Linked `decisions/ui-stack.md` to `decisions/http-framework.md` (replaces the html/template + htmx fragment note with the full four-piece UI stack).
* **Creation**: Documented concept `decisions/ui-stack.md` (UI stack: templ + Tailwind + Alpine + htmx).
* **Update**: Linked `notes/config-projection-and-mergesave.md` to `notes/config-resolution-gotchas.md` (Projection and MergeSave build upon configuration resolution rules.).
* **Creation**: Documented concept `notes/config-projection-and-mergesave.md` (Projection and MergeSave verified for internal/config).
* **Update**: Linked `notes/events-bus-implementation.md` to `architecture/event-driven-core.md` (Refined implementation of the event bus conforms to the event-driven core architecture.).
* **Update**: Updated concept `notes/events-bus-implementation.md`.
* **Update**: Linked `notes/events-bus-implementation.md` to `architecture/event-driven-core.md` (Implementation of the event bus conforms to the event-driven core architecture.).
* **Creation**: Documented concept `notes/events-bus-implementation.md` (Events bus implementation verified with FIFO, panic recovery, and recorder).
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
