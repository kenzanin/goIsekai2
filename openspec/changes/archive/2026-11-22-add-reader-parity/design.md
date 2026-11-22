# Design: add-reader-parity

## D1. Reader chapter options are server-rendered

`renderRead` obtains the full ordered chapter list from `reader.Service.Chapters` and passes
`[]components.ChapterOption{ID, Number, Title}` into `components.Reader`. The boot JSON carries these options
under `chapters`. Selecting one calls the same `switchTo({ch, id})` path as prev/next, so there is one in-place
switch implementation.

## D2. Page controls stay client-side

The page `<select>` is generated from `pages[]` (no extra endpoint), and the bottom numbered buttons are
generated from the same list. For more than 24 pages the numbered row collapses to the existing big current-page
jump control; the top `<select>` remains the full list. Both jump through `setPage`, so positions and history
stay consistent.

## D3. Keyboard ignores form controls

`onKey` returns early when `event.target` is `INPUT`, `SELECT` or `TEXTAREA`, matching the baseline's
form-control exclusion while keeping arrows working on the image surface.

## D4. Prefetch on open

`init()` calls the debounced `save()` once after boot. That POST already drives `prefetchWindow`, so opening a
chapter now warms the sliding image window without adding a second worker or event.

## D5. Read progress storage

Use the existing `plugin_state` table with `PluginID = "host"` and key
`read/<plugin>/<slug>/<chapterID>`; value is the highest page reached. Store methods
`PutReadProgress`, `ReadProgress` and `ReadProgresses` keep SQL/GORM usage in the store package. When
`POST /reader/position` records a page count equal to the stored page count, set `ChapterRead` (bit0).

## D6. Bulk flags in one request

`POST /chapter/flags` accepts repeated `ch_id` form values (htmx default) and a JSON `chapter_ids` array.
The handler loops the ids through `Store.SetChapterFlags`; a missing id list keeps the current single-id behavior.
Bulk mark-all sets bit0, clear-read clears bit0, and progress labels are untouched by either.

## D7. Detail page data

`renderChapters` loads all chapter options before slicing the 20-row page, and loads progress per row via
`ReadProgresses`. `ChapterRow` gains `Read int`; the row renders `Read/Pages` when a progress record exists and
`0/Pages` otherwise. A JSON script carries all chapter options so "mark previous" can resolve ids client-side
after the confirm names the range.

## D8. Tests

Cover: reader page contains chapter/page selects; large chapters collapse numbered buttons; keyboard ignored in
selects; flags handler applies multiple ids; progress reaches page count and sets read bit; chapter rows show
counters. End-to-end checks the live reader controls on the demo daemon.
