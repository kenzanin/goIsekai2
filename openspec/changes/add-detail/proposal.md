# Proposal: add-detail

## Why

`image-cache` is archived, so saved metadata and covers can be enriched and served offline. Detail is
next in the porting order and is the read/write surface for manga metadata, chapter lists, refresh and
user-selected enrichment.

## What Changes

- Add `GET /view/manga/:plugin/:slug` shell with per-field placeholders and read/cbz chapter rows.
- Add `GET /api/detail/:plugin/:slug` saved-row read, cold materialization, stale-serve and 404 guard.
- Add `POST /api/refresh/:plugin/:slug` with all-or-nothing cache overwrite and `update_attempts`
  stamping.
- Add enrichment registry under `data/enrichment/<id>/`, two fetch paths (normal/force), user-selected
  candidate merge, cover wizard and reset actions.
- Add related-reading view and accessible icon controls.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `detail` — adds daemon wiring (routes before listen, dedicated JSON routes instead of the baseline
  action-dispatch endpoint) on top of the frozen detail contracts.

## Impact

- Code: new `internal/detail` package, HTTP routes, templ detail/enrichment components.
- Dependencies: none beyond archived stack.
- No cgo: `CGO_ENABLED=0 go build ./...` stays the gate.
- Out of scope: site plugins, scheduler internals, reader UI changes.

## Baseline

cppIsekai `manga-detail` and `detail-enrichment` specs plus the recorded dedicated-route divergence.
