# Proposal: add-reader

## Why

`library` is archived, so saved chapters and cached pages are available. The reader is next in the
porting order and is the main consumer of chapter lists, page lists and reader positions. This change
implements the reading surface, pages API, position/resume and prefetch.

## What Changes

- Add `GET /read/:plugin/:slug?ch=<n>&ch_id=<id>` with four modes (webtoon, paged, spread-h, spread-v),
  direction-aware pairing and per-manga persisted mode/zoom.
- Add `GET /api/pages/:plugin/:slug/:ch` resolving chapters by number or exact id, returning host-owned
  `/cache/img` URLs, stale-list fallback and neighbour numbers/ids.
- Add `POST /reader/position` partial upsert with resume and prefetch trigger, plus adjacent page-list
  prefetch via direct calls.
- Add in-place chapter switch with `history.pushState`/`popstate`, overlay controls, keyboard
  navigation and cross-chapter completion.
- Add detail-page chapter list controls (skip/read flags, pagination, bulk actions).

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `reader` — adds daemon-level wiring (routes registered before listen, prefetch as direct call not
  event) on top of the frozen reader contracts.

## Impact

- Code: new `internal/reader` package, HTTP routes and templ reader/detail components.
- Dependencies: none beyond the archived stack.
- No cgo: `CGO_ENABLED=0 go build ./...` stays the gate.
- Out of scope: image cache route implementation (owned by `image-cache`), site plugins, update scheduler.

## Baseline

cppIsekai `reader-page` and `reader-settings` specs define modes, window math and position semantics;
V2 keeps routes and JSON shapes byte-compatible.
