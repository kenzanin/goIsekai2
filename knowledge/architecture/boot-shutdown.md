---
type: architecture
title: "Boot and shutdown: ordered startup, dependency-ordered teardown, and the test that proves both"
description: "Fifteen-step boot order derived from cppIsekai main.cpp with failure modes, the ordering constraints behind it, the shutdown ladder (stop accepting, workers, scheduler, plugins, browser, log drain), deliberate divergences, and one integration test contract."
tags: [architecture, boot, shutdown, lifecycle, testing]
generated: { by: agent/mcp, at: "2026-11-21T12:17:40Z" }
status: stable
---

## Boot sequence (baseline `src/main.cpp`, reproduced with intent)

| # | Step | Fails how |
|---|---|---|
| 1 | Arm crash tracing (baseline: backward-cpp) before anything heavy | n/a |
| 2 | Parse CLI; `--dump-config` prints the template and exits 0 without touching data | exit 0 |
| 3 | Require `--config`; missing prints usage and exits 2 with no `data/` created | exit 2 |
| 4 | Resolve config: default < ini < env < flag (`CPPISEKAI_*`/`GOISEKAI_*`) | invalid value falls back to default, logged |
| 5 | Apply log level to both owners (app logger and HTTP server logger) | bad value keeps `info` |
| 6 | Create `data_dir` | fatal if uncreatable |
| 7 | Boot browser backend (baseline: own Lightpanda child); abort loud if unavailable | exit 3 |
| 8 | Open store: migrations, `MaxOpenConns(1)`, WAL, busy timeout | fatal, never serve on a half-migrated DB |
| 9 | Start log drain worker | n/a |
| 10 | Build worker pools: db=1, plugin=10, log=1 | allocation only |
| 11 | Plugin host: discovery, `load_all` from DB, host tools wired to the db worker | per-plugin failure records `failed`, does not abort boot |
| 12 | Image cache: config + namespace snapshot from `store.plugins()` | fatal if cache root uncreatable |
| 13 | Start scheduler (reads its knobs fresh each tick) | n/a |
| 14 | Start tray (detached, no-op when headless or disabled) | non-fatal |
| 15 | Bind HTTP and block (`concurrency(4)` in baseline) | port busy = fatal |

Ordering constraints that are not negotiable:

1. **Config before everything.** Every later step reads resolved values, never the environment directly.
2. **Store before plugins and cache.** `load_all` reads plugin rows; the cache namespace map is derived from
   `store.plugins()`. Building the cache first would key files on stale or missing display names.
3. **Browser before store in the baseline** only because a dead browser makes a whole plugin family useless and the
   failure should be loud at boot, not on first click. In Go the equivalent is: start the browser supervisor before
   serving, fail boot only if the configured backend is required.
4. **Pools before scheduler.** The scheduler submits work through them; starting it first races the queues.
5. **Bind HTTP last.** Until the port is open, a supervisor sees the daemon as down, which is the correct signal.
6. **Log drain starts before the components that log.** Otherwise the first lines are lost.

While boot runs, nothing serves. There is no half-ready state: the port opens only after step 15.

## Shutdown ladder

The baseline's `serve()` blocks; everything after it is teardown, and the order is by dependency, not by
convenience:

```mermaid
sequenceDiagram
  participant O as signal / tray restart
  participant H as http server
  participant W as in-serve workers
  participant S as scheduler
  participant P as plugin host
  participant B as browser child
  participant L as log drain
  O->>H: SIGTERM / app.stop()
  H->>H: stop accepting, drain in-flight requests
  H->>W: scope exit -- page prefetch + cover workers stop
  Note over W: must stop before the cache and store they call
  H->>S: scheduler.stop() (join the tick thread)
  H->>P: plugin manager shutdown (persist state, stop runtimes)
  H->>B: stop only the child we own; adopted externals survive
  H->>L: final drain, then exit loop
```

Rules:

1. **Stop accepting before stopping workers.** A request arriving after the server stops would find a half-torn
   component.
2. **In-flight requests get a deadline**, not an arbitrary wait. Baseline has no deadline; we add one and log the
   stragglers.
3. **The scheduler joins before the pools are destroyed.** A tick firing after pool destruction is a crash.
4. **Persist plugin state before stopping runtimes**, so a crash mid-shutdown leaves the DB consistent with what the
   user saw.
5. **The log drain stops last** (baseline sets its flag at the end of `main`), because every other step still logs.
6. **A tray-initiated restart respawns only after teardown completes**, otherwise two processes race for the SQLite
   file and the port.

## Deliberate divergences (documented, not accidental)

| Baseline behaviour | V2 behaviour | Why |
|---|---|---|
| Log drain thread is detached and never joined; the last lines can be lost on exit | drain is joined with a bounded wait, then flushed | silent data loss in the only diagnostic channel |
| Plugin state has no persisted lifecycle column | state persisted via FSM, reconciled at boot | enables fast-fail on illegal transitions |
| No explicit bus drain | bus actor drains its queue before exit, with a deadline | events published at shutdown must not be dropped silently |
| Browser child stopped after scheduler, before log drain | same order, but with a supervisor health check | parity kept; shutdown gets a timeout |

## Testable contract

One integration test, no framework beyond stdlib:

1. Boot with a temp `data/` and `:0` (ephemeral port). Assert `/health` returns ready only after the store answered a
   query.
2. Save a manga and assert the row exists **and** the event was observed after the commit (recorded bus, not a sleep).
3. Send SIGTERM. Assert: port closed, DB handle closed, log file flushed, `goleak.VerifyNone` clean.
4. Boot a second instance against the same data dir while the first still runs. It must fail fast on the SQLite lock,
   not hang and not migrate.

These four assertions are what make the boot order and the shutdown ladder real rather than aspirational.
