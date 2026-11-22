# reader Specification

## Purpose

Defines the reading experience: the reader route and its four modes, the pages API that resolves a
chapter and hands the browser host-owned image URLs, position persistence with resume, in-place
chapter switching keyed by plugin chapter id, adjacent prefetch, the full-screen overlay surface, and
the detail page's chapter list with its read/skip controls. Chapter numbers are always labels;
identity is always the plugin's chapter id.

### Modes

| Mode | Rendering |
|---|---|
| `webtoon` | vertical scroll, one page per row, lazy-loaded |
| `paged` (`ltr`/`rtl`) | one page per screen |
| `spread-h` (`ltr`/`rtl`) | two pages side by side, direction-aware pairing |
| `spread-v` | two pages stacked, paged downward |

## Requirements

### Requirement: Reading modes

`GET /read/:plugin/:slug?ch=<n>` SHALL render the chapter in one of the four modes above. Mode and
direction SHALL persist per manga and be restored on reopen. Default SHALL be `webtoon` for
long-strip content and `paged rtl` otherwise, and an explicit user choice SHALL beat the plugin's
long-strip hint.

#### Scenario: Spread pairing with an odd page count
- **WHEN** pages 1..5 are read in `spread-h`
- **THEN** page 1 renders alone and pages (2,3) and (4,5) pair, with the lower-numbered page on the right in `rtl`

#### Scenario: Mode persists per manga
- **WHEN** a manga is switched to `spread-v` and reopened later
- **THEN** the reader restores `spread-v` and the last position

#### Scenario: User override wins
- **WHEN** the plugin flags long-strip but the user chose `paged`
- **THEN** the reader opens in `paged`

### Requirement: Pages API

`GET /api/pages/:plugin/:slug/:ch` SHALL resolve the chapter entry, obtain its page list through the
plugin (serving the last known list when a refresh fails), and return host-owned image URLs
(`/cache/img?...`) built from that entry's chapter id. The response SHALL carry `chapter`, `pages[]`,
and neighbours as both numbers and ids (`prev_ch`, `next_ch`, `prev_ch_id`, `next_ch_id`).

#### Scenario: Pages in reading order
- **WHEN** the plugin returns a page list for the resolved chapter
- **THEN** `pages[]` is emitted in plugin order, each entry a `/cache/img` URL carrying the chapter id

#### Scenario: Unknown chapter
- **WHEN** no entry matches the number and no `ch_id` matches
- **THEN** the route answers 404 with `{"ok":false,"error":"unknown chapter"}`

#### Scenario: Stale list beats a failed refresh
- **WHEN** the plugin call for the chapter fails but the store holds a page list for that chapter id
- **THEN** the stored list is served and the failure is logged

### Requirement: Exact chapter selection

The route SHALL accept an optional `ch_id`: when it matches a chapter entry, that entry is the target
(its id is requested from the plugin, carried by the image URLs and returned); when absent or
unmatched, the target SHALL be the first entry whose number equals `<ch>`. Neighbours SHALL be derived
from position in a stable number-ordered list, so duplicated numbers yield one navigation stop per
entry.

#### Scenario: Chapter id selects the duplicate
- **WHEN** two entries share number 24 and the request carries the second entry's `ch_id`
- **THEN** the plugin is asked for that id and those pages are served

#### Scenario: Sibling is the neighbour
- **WHEN** the target is the first of two entries numbered 24
- **THEN** `next_ch` is 24 with `next_ch_id` of the second entry, and the reverse holds for `prev`

#### Scenario: Refresh keeps the release
- **WHEN** the reader reloads while on the second entry numbered 24
- **THEN** the same entry's pages are served, not the first entry's

### Requirement: Position persistence and resume

`POST /reader/position` SHALL upsert `{plugin, slug, chapter, page, mode, direction}` per manga,
keeping stored values for absent fields, and SHALL trigger prefetch. Reopening a manga SHALL resume
the persisted chapter and page, and the detail page SHALL show a "continue ch N p M" hint when a
position exists.

#### Scenario: Resume after restart
- **WHEN** a manga is left at chapter 12 page 3 and the daemon restarts
- **THEN** reopening resumes at chapter 12 page 3

#### Scenario: Partial update keeps mode
- **WHEN** a position POST carries only chapter and page
- **THEN** the stored mode, direction and zoom are unchanged

### Requirement: In-place chapter switch

Advancing past the last page or going back from the first SHALL request `/api/pages` with
`next_ch_id`/`prev_ch_id` and render the result in place, without a document reload. The address bar
SHALL keep both `ch` (label) and `ch_id` (exact release) via `history.pushState`, and `popstate` SHALL
restore the matching chapter and page in place. Mode and zoom SHALL survive the switch.

#### Scenario: Next chapter without reload
- **WHEN** the reader advances past the last page
- **THEN** the next chapter renders in place with no document navigation

#### Scenario: Back button restores in place
- **WHEN** the user switched chapters in place and presses back
- **THEN** the previous chapter and page restore in place

#### Scenario: Refresh after switch
- **WHEN** the user refreshes after switching to chapter 5 page 2 in place
- **THEN** chapter 5 page 2 loads directly

### Requirement: Adjacent prefetch

After a chapter renders, the reader SHALL nudge the prefetch worker immediately, then continue to
prefetch the adjacent chapters' page lists in the background, next first then previous, through a direct
call (not an event). Opening a chapter SHALL therefore warm the sliding image window without waiting for a
manual page change. Switching to a prefetched chapter SHALL render without a new page-list fetch.

#### Scenario: Prefetch on open
- **WHEN** a chapter opens and renders its first page
- **THEN** a position post triggers the sliding-window prefetch for that chapter

#### Scenario: Instant switch
- **WHEN** the user advances after the next chapter was prefetched
- **THEN** no page-list request is issued for it

#### Scenario: Prefetch order
- **WHEN** a chapter finishes rendering and both neighbours exist
- **THEN** the next chapter's page list is fetched first, then the previous one's

### Requirement: Full-screen surface and overlay

The reader SHALL be a full-screen surface with the overlay navigation hidden by default. A tap in the
centre third of the image SHALL toggle the overlay; taps on the left/right thirds SHALL page
(webtoon scrolls instead of paging on a centre tap). The overlay top bar SHALL contain a back-to-detail
control, a chapter `<select>` listing every chapter and a page `<select>` listing every page of the current
chapter. The bottom bar SHALL contain mode, direction and zoom selectors plus numbered page-jump buttons
1..N; for chapters larger than 24 pages the numbered row SHALL collapse to a single current-page jump
control while the top page `<select>` still lists every page. While the nav is hidden the reader SHALL show
a dim, non-interactive `p N / M` indicator.

#### Scenario: Centre tap toggles
- **WHEN** the nav is hidden and the user taps the centre third
- **THEN** the overlay appears; tapping again hides it

#### Scenario: Chapter select switches in place
- **WHEN** the user picks a chapter in the top bar select
- **THEN** the reader requests that chapter's exact id, renders it without a reload, and updates the URL and position

#### Scenario: Page jump
- **WHEN** a chapter has 12 pages and the user picks page 7 in the bar
- **THEN** the reader shows page 7 and records the position

#### Scenario: Numbered page buttons
- **WHEN** a chapter has 12 pages and the nav is shown
- **THEN** the bottom bar shows buttons 1..12 with page 7 highlighted when current

#### Scenario: Large chapter compacts navigation
- **WHEN** a chapter has 80 pages and the nav is shown
- **THEN** the bottom bar shows one current-page control instead of 80 buttons and the top page select lists all 80

#### Scenario: Webtoon centre tap
- **WHEN** the reader is in webtoon mode and the user taps the centre third
- **THEN** the overlay toggles and no page change occurs

### Requirement: Cross-chapter completion

Tapping past the last page of a chapter SHALL advance to the next chapter; at the last chapter the
reader SHALL show the done overlay instead of a dead tap. The chapters envelope SHALL carry each
entry's page count so the overlay and read-completion logic do not need an extra call.

#### Scenario: Advance from the last page
- **WHEN** the user taps forward on the final page of a non-final chapter
- **THEN** the next chapter opens at its first page

#### Scenario: Done overlay
- **WHEN** the user taps forward on the final page of the final chapter
- **THEN** the done overlay appears

### Requirement: Keyboard navigation

The reader SHALL keep keyboard navigation: arrows and page-up/page-down move by one page (or one
screen in webtoon), horizontal arrows follow the reading direction in `rtl`, and `Escape` closes the
overlay or returns to the detail page when no overlay is open. Keys targeted at `INPUT`, `SELECT` or
`TEXTAREA` controls SHALL NOT move the page.

#### Scenario: Arrow respects direction
- **WHEN** the mode is `paged rtl` and the user presses the left arrow
- **THEN** the next page is shown, not the previous one

#### Scenario: Form control keeps its key
- **WHEN** the page-jump input is focused and the user presses an arrow key
- **THEN** the input receives the key and the reader page does not change

### Requirement: Chapter list controls

The detail page chapter list SHALL paginate at a fixed size with prev/next and a position indicator,
hiding the pager when the list fits. Each row SHALL show a skip control (open/closed eye) reflecting
`chapters.flags` bit1 and toggling it through `POST /chapter/flags`, a read check lit iff bit0 is set,
and a `read pages / total pages` counter sourced from stored progress. The header SHALL offer
mark-all-read, clear-read-status and mark-previous-read actions; mark-previous SHALL confirm a range naming
chapter numbers and SHALL issue one flags request for the exact ids in that range. Bulk actions SHALL set or
clear bit0 through the same route without erasing progress counters.

#### Scenario: Long list paginates
- **WHEN** the list holds 45 entries at 20 per page
- **THEN** page 1 shows 20 rows and the indicator reads 1/3

#### Scenario: Skip persists
- **WHEN** the eye is clicked and the page reloads
- **THEN** the toggled state is still shown

#### Scenario: Read check
- **WHEN** stored progress for a chapter equals its page count
- **THEN** the host sets bit0 and the row renders the lit check

#### Scenario: Read progress counter
- **WHEN** a chapter has 12 pages with 5 recorded as read
- **THEN** its row shows 5/12

#### Scenario: Bulk mark previous
- **WHEN** mark-previous is invoked on chapter 10 of 30
- **THEN** a confirm names chapters 1-10 and one flags request sets bit0 for exactly those ten ids

#### Scenario: Clear keeps progress
- **WHEN** clear-read is confirmed
- **THEN** every check is unlit while per-row progress counts remain displayed

### Requirement: Reader route wiring

The daemon SHALL register the reader routes before it opens the HTTP port. Page-list prefetch SHALL be
a direct service call, never an event, and SHALL never block the rendering request. A failed plugin
refresh SHALL fall back to the last stored page list for that exact chapter id while logging the
failure with plugin and chapter id. Position writes SHALL keep absent fields and SHALL be persisted
before any prefetch starts.

#### Scenario: Prefetch is a direct call
- **WHEN** a chapter finishes rendering
- **THEN** the host calls the page-list resolver directly for next then previous chapter, without publishing an event

#### Scenario: Stale list survives a failed refresh
- **WHEN** the plugin page call fails and a stored page list exists for that chapter id
- **THEN** the stored list is returned and the failure is logged with plugin and chapter id

#### Scenario: Position fields are preserved
- **WHEN** a position POST carries only chapter and page
- **THEN** stored mode, direction and zoom stay unchanged before prefetch begins
