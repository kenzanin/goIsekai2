# Design: detail-pagination

## Context

See `proposal.md`. `renderDetail` renders every chapter row; the largest observed title carries
1,199 of them, dominating the 8.8 s response. The store layer is already fast (`/api/detail`
0.001 s), so the fix is bounded rendering, not caching.

## Goals / Non-Goals

- Goals: bounded HTML size and render time on the detail route; pager identical to `/chapters`;
  full-list paths preserved.
- Non-Goals: no API payload changes, no client-side paging, no configurable page size, no
  reader-page changes.

## Decisions

### D1 — Server-side slice in `renderDetail`, page size 20

`renderDetail` reads `?page=` (default 1; non-numeric or out-of-range clamps to 1) and renders rows
`[(page-1)*20, page*20)`. The fixed size of 20 matches the reader chapter-list convention, so one
number governs both lists.

Alternative: client-side collapse — rejected, the response still ships all 1,199 rows of HTML.

### D2 — Pager is the `/chapters` pager

The same prev/next plus `n/total` partial as the chapters page, omitted entirely when the list fits
one page. A second pager variant would drift; reuse keeps both lists identical.

### D3 — Full list stays one click away

`/api/chapters/:plugin/:slug` and `/chapters/:plugin/:slug` are untouched and the detail page keeps
no hidden rows, so nothing observable is lost by the cap.

## Risks / Trade-offs

- [Deep link lands past the end] → out-of-range pages clamp to 1; every row stays reachable via the
  full-list page.
- [Pager drift between routes] → single shared partial exercised by both routes' tests.

## Migration Plan

One lane: clamp+slice and pager in `renderDetail`, fixture tests for 45→20 rows/1/3 and
12→no pager, `openspec validate detail-pagination --strict` at close-out.

## Open Questions

None.
