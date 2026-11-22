# Proposal: add-reader-parity

## Why

The V2 reader already has the core engine (four modes, exact chapter ids, resume, in-place switch, history,
image prefetch), but several visible controls promised by the baseline and by our own spec are still missing:
a chapter picker inside the overlay, a real page list in the top bar, numbered page buttons in the bottom bar,
per-chapter read progress counters, and the bulk read actions on the detail page. This change closes those gaps
against cppIsekai's `reader-page`, `reader-fullscreen` and `chapter-list-controls` specs and keeps every V2
enhancement (zoom, keyboard, done overlay, htmx flags, per-manga chrome state).

## What Changes

- Overlay top bar gains a chapter `<select>` built from all chapters and a page `<select>` built from the loaded pages.
- Overlay bottom bar gains numbered page buttons 1..N with a compact fallback for large chapters; current page is highlighted.
- Hidden-nav indicator becomes a dim `p N / M`; keyboard navigation ignores keys targeted at form controls.
- Reader open nudges prefetch immediately instead of waiting for the first manual page change.
- Detail chapter rows show `read pages / total pages`; bulk mark-all, clear-read and mark-previous use one flags request with a confirm.
- `/chapter/flags` accepts multiple chapter ids; read progress is stored under host plugin state and reaching the page count sets the read bit.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `reader` — fills the already-specified overlay, keyboard, prefetch and chapter-list contracts with concrete controls and backend support.

## Impact

- Code: `internal/http/components/reader.templ`, `internal/http/components/chapters.templ`, matching `*_templ.go`, `internal/http/server.go`, `internal/http/server_test.go`, `internal/reader/service.go` and its tests, `internal/store/progress.go`.
- API: `POST /chapter/flags` accepts repeated `ch_id` form values or a JSON `chapter_ids` array; responses keep `{ok:true}`.
- No cgo, no new dependency, no route removal, no change to plugin ABI or cache layout.

## Baseline

cppIsekai `openspec/specs/reader-page`, `reader-fullscreen` and `chapter-list-controls`. Divergences that are V2
enhancements (zoom, keyboard hints, htmx toggles, per-chapter id navigation) remain; nothing is removed.
