---
type: Decision
title: "No generic worker pool: job runners plus semaphores"
description: "goIsekai V2 does not build a general worker pool: only the scheduler has bounded job runners, while plugin calls and fetches use semaphores and SQLite serialization comes from database/sql."
tags: [concurrency, pool, semaphore, jobs, plugins, decision]
generated: { by: agent/cli, at: "2026-11-21T14:38:39Z" }
status: stable
---

## Context

The cppIsekai baseline uses three `BS::thread_pool` instances: `db_pool(1)`
because `sqlite_orm`'s storage is not thread-safe, `plugin_pool(10)` for
blocking guest calls behind crow's `concurrency(4)`, and `log_pool(1)`.
The user asked whether that worker-pool model survives the Go rewrite now that
the store runs on GORM and FTS5/BM25 is confirmed available.

## Decision

Do **not** build a generic application worker pool. Use these bounds instead:

| Resource | Mechanism | Limit |
|---|---|---|
| SQLite | `database/sql` via GORM, `SetMaxOpenConns(1)` | one writer, driver-serialized |
| Plugin call | goroutine per request + per-plugin semaphore | one guest call per runtime |
| Fetcher | global semaphore + per-host semaphore | configurable in-flight caps |
| Scheduler jobs | bounded queue + N job runners | N = 2 by default |
| Reader/cover prefetch, log pump, bus, stream hub | one owner goroutine each | 1 |

The only real worker pool is the scheduler's bounded set of job runners. If
cross-plugin parallelism is needed later, add **runtime instances** (slots),
not a generic pool of goroutines: each guest runtime remains single-owner and
its state machine (`Enabled/Idle/Busy`) is the source of truth.

## Evidence

- Baseline `src/http/http.cpp:72-98` passes `db_pool` and `plugin_pool` into the
  HTTP layer; `knowledge/decisions/db-read-parallel` explicitly defers
  read-parallelism and keeps single-writer DB access.
- `knowledge/architecture/concurrency-ownership.md` rows 22-32 and its
  backpressure table already encode this model: one owner per resource,
  semaphores for plugin/fetch, bounded runners for jobs, single goroutines for
  prefetch and log.
- GORM manages connections, not work: `SetMaxOpenConns(1)` replaces the C++
  `db_pool(1)` entirely. FTS5/BM25 works on that same connection (verified in
  `notes/fts5-purego-probe`), so it adds no pool requirement.

## Alternatives considered

| Option | Verdict |
|---|---|
| Generic fixed-size goroutine pool for all work | rejected: hides backpressure, duplicates `net/http` scheduling, and makes owner-per-resource harder to see |
| Baseline-style `plugin_pool(10)` | rejected: a fixed thread count is a C++ scheduling workaround; Go uses request goroutines plus semaphores |
| Unbounded goroutines (`go fn()` per event) | rejected by the existing forbidden-patterns list |
| More reader DB connections for FTS5 scans | deferred, same as the baseline decision; revisit only if dedup/grid scans become hot |

## Consequences and review triggers

- No new package for a pool; the job runner set belongs to the scheduler work.
- Tests must prove the semaphores bound in-flight calls and that job runners
  drain without growing the queue (`go test -race`, goroutine-leak checks).
- Review if a guest runtime turns out to be safely concurrent (then one shared
  slot could serve many calls), if FTS5 reads become a measured bottleneck, or
  if job latency requires tuning N above 2.

# Related Concepts
- [Concurrency and ownership: one owner per mutable resource, bounded everything](../architecture/concurrency-ownership.md): locks the pool policy already implied by the ownership and backpressure tables
- [Store access: GORM with one raw-SQL boundary](store-access.md): GORM's connection pool replaces the baseline db_pool; no work pool is added
- [Use qmuntal/stateless as the FSM engine behind typed domain wrappers](state-machine-engine.md): plugin runtime slots map to Enabled/Idle/Busy states, not to generic workers
