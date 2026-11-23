# Proposal: detail-pagination

## Why

`GET /view/manga/atsumaru/sVC2A` returns 2,043,045 bytes in 8.8 s because `renderDetail` emits all
1,199 chapter rows on every request, while `/api/detail` serves the same title from the store in
0.001 s. The HTML shell, not the data layer, is the cost.

## What Changes

- MODIFIED `detail` — the detail page renders at most 20 chapter rows per request, selected by
  `?page=`, with the `/chapters` pager (prev/next, n/total) hidden when one page covers the list.
- The full chapter list stays reachable: `/api/chapters/:plugin/:slug` (JSON) and
  `/chapters/:plugin/:slug` (HTML) are unchanged.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `detail` — caps rendered chapter rows at 20 per request behind `?page=` with the shared pager.

## Impact

- Code: `renderDetail` in `internal/http/server.go` (page parsing, row slicing, pager emission) and
  the detail template.
- No new dependencies; `CGO_ENABLED=0 go build ./...` gate unchanged; no API payload changes.

## Baseline

cppIsekai serves the full chapter list from the dedicated list route; the detail page carrying a
bounded window plus the unchanged full-list routes preserves that behavior.
