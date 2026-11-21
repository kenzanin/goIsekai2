# Proposal: add-library

## Why

`host-tools` is archived, so plugins can now return real detail data. The library is the next
capability in the porting order and the consumer of saved manga rows, FTS5 search and plugin calls.
This change implements the save/list/filter/search routes and the SSR grid.

## What Changes

- Add `POST /api/library`, `POST /api/library/remove` and `GET /api/library` with the frozen JSON
  shapes, placeholder-name protection and post-commit `library.saved` / `library.removed` events.
- Add status-flag operations (new/done/dim/pin) that preserve unrelated bits, and category/duplicate
  filters driven by stored `status_flags` and `dup_group`.
- Add the SSR home/library grid with 12-per-page client-side pagination, `data-*` attributes, empty
  state and `?plugin=` deep-link filtering.
- Add enabled-plugin-only search form and FTS5-ranked local title search with API-cache write-through.
- Reuse the archived store, plugin-host, host-tools and UI stack; no new runtime dependency.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `library` — adds the daemon-level wiring requirement (routes registered before listen, events after
  commit) on top of the frozen library contracts.

## Impact

- Code: new `internal/library` package plus HTTP routes and templ components for the grid/search.
- Dependencies: none beyond the archived stack.
- No cgo: `CGO_ENABLED=0 go build ./...` stays the gate.
- Out of scope: reader/view pages, detail enrichment routes, update scheduler, site plugins.

## Baseline

cppIsekai `library-save` and `library-page` specs define routes, status flags and grid behavior;
V2 keeps the JSON contract byte-compatible.
