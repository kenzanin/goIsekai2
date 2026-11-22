# Design: add-reader

## Context

See `proposal.md`. The `reader` capability spec freezes modes, window math and position semantics.
cppIsekai `reader-page` is the baseline. V2 reuses the archived store, plugin host, host tools and UI
stack.

## Goals / Non-Goals

- Goals: exact chapter-id identity, four modes with direction-aware pairing, persist-before-prefetch,
  in-place navigation, stale-list fallback, no new dependency.
- Non-Goals: image cache route (image-cache capability), site plugins, scheduler.

## Decisions

### D1 — Chapter identity is the plugin id

Numbers are labels only. `ch_id` wins when it matches; otherwise the first entry with matching number
is used. Neighbours come from a stable number-ordered list of entries. Baseline: cppIsekai exact
chapter selection.

### D2 — Page-list resolver

`internal/reader.Service.Pages(ctx, plugin, slug, number, id)` calls the plugin `pages` method via
the archived host, caches the list by `(plugin, slug, chapter_id)` in the store, and falls back to the
stored list on failure. URLs are `/cache/img?plugin=...&slug=...&ch_id=...&p=...`.
Baseline: cppIsekai pages API with 10-url window.

### D3 — Position upsert

`POST /reader/position` keeps stored fields for absent keys via a typed store method, persists first,
then starts prefetch. Baseline: cppIsekai reader position table.

### D4 — Prefetch

Adjacent page lists are fetched with direct calls in a bounded goroutine: next first, then previous.
Results go into an in-memory cache keyed by chapter id; no event is published. Baseline: cppIsekai
prefetch semantics, minus the event.

### D5 — Reader surface

Templ + Alpine render four modes; window math is pure Go helpers (`pairs`, `window`) with table tests.
Overlay, keyboard and history are client-side only; SSR stays stateless. Baseline: cppIsekai fullscreen
overlay and keyboard map.

### D6 — Chapter list controls

Detail page chapter rows read `chapters.flags` from the store; skip/read toggles call
`POST /chapter/flags` with bit ops. Pagination is fixed at 20. Baseline: cppIsekai chapter list.

## Risks / Trade-offs

- [plugin page call slow] → context deadline from config; stale list fallback.
- [duplicate chapter numbers] → exact id always wins; neighbours are entry-based.
- [prefetch race] → one bounded goroutine per reader session, results keyed by id.

## Migration Plan

Greenfield. Rollback is reverting the change commit.

## Open Questions

None. Modes and window math are fixed by the capability spec.
