# Design: add-library

## Context

See `proposal.md`. The `library` capability spec freezes routes, status-flag semantics and the grid
contract. cppIsekai `library-save`/`library-page` are the baseline; V2 reuses the archived store,
plugin host, host tools and UI stack.

## Goals / Non-Goals

- Goals: frozen JSON shapes, commit-before-event, bit-preserving flag ops, FTS5 ranking with LIKE
  fallback, SSR grid with client-side filtering, no new dependency.
- Non-Goals: reader pages, detail enrichment, scheduler, site plugins.

## Decisions

### D1 — Repository over typed store

`internal/library.Service` owns all mutations and calls typed store methods. No raw SQL outside
`internal/store/sql`. Baseline: cppIsekai store callers; V2 keeps SQL at the boundary.

### D2 — Save and placeholder protection

`UpsertSaved` skips `name` when it is empty or equal to slug and a good name is already stored.
Baseline: archived add-config parity fix for `library-save`.

### D3 — Flags are bit operations

`is_new`, `is_done`, category bits and `is_dim` are set/cleared with bit ops; dim/pin toggles
preserve every other bit. Baseline: cppIsekai `status_flags` table.

### D4 — Events after commit

`library.saved` and `library.removed` are typed events on the archived bus, published only after a
successful commit. Baseline: cppIsekai publish-after-write ordering.

### D5 — Search

Search form lists only enabled plugins from `Host.Rows`. Local search calls the store FTS5 helper
(BM25, prefix) with LIKE fallback; results are cached through `api_cache`. Baseline: cppIsekai
local search and FTS5 dedup.

### D6 — Grid

Templ component renders cards with `data-dup`, `data-status-flags`, `data-status`, 12-per-page Alpine
pagination, empty state and `?plugin=` client filter. Saved mode ignores `?q=`/`?page=` server-side.
Baseline: cppIsekai home grid.

## Risks / Trade-offs

- [FTS5 absent] → LIKE fallback and `dup_group = 0`, no user-visible error.
- [Event ordering race] → publish inside the service only after the store Tx returns.
- [Grid state drift] → SSR emits all flags; Alpine only filters/paginates, never mutates.

## Migration Plan

Greenfield: demo plugin and empty library. Rollback is reverting the change commit.

## Open Questions

None. Routes and behavior are fixed by the capability spec.
