# Spec Delta

## MODIFIED Requirements

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
