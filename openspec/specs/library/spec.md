# library Specification

## Purpose

Defines the saved-manga library and the plugin-backed search that feeds it: the JSON routes that
save, remove, list, pin, dim and filter entries, the `status_flags` bitfield that carries every
derived per-manga state, duplicate clustering over stored source metadata, the home grid contract,
and the search form that only offers enabled plugins. Persistence lives in the store capability; this
capability owns the routes and their observable behaviour.

### status_flags layout

| Bit | Value | Meaning |
|---|---|---|
| bit0 | 1 | `is_new` — a refresh observed chapter growth beyond `total_ch`; cleared by a reader touch |
| bit1 | 2 | `is_done` — `total_ch > 0` and `read_ch >= total_ch` |
| bit2 | 4 | category *ongoing* |
| bit3 | 8 | category *hiatus* |
| bit4 | 16 | category *canceled* / *cancelled* |
| bit5 | 32 | category *completed* |
| bit6 | 64 | `is_dim` — user toggled dim+blur on the card |

The source `status` string stays stored verbatim and remains the display source; the category bits
are its derived classification. An unrecognised status sets no category bit and never errors. Every
stat writer preserves bit6 except the dim toggle.

## Requirements

### Requirement: Save and unsave

`POST /api/library` SHALL upsert `{plugin, slug, name, cover}` keyed on `(plugin, slug)` and
`POST /api/library/remove` SHALL delete that row. Both SHALL respond with the resulting state
(`{"saved": true|false}`), SHALL be idempotent, and SHALL publish `library.saved` / `library.removed`
after the commit.

#### Scenario: Add to library
- **WHEN** a user adds an unsaved manga from the detail page
- **THEN** the row is persisted, the response reports `saved: true`, and the button reads "In library"

#### Scenario: Re-add refreshes
- **WHEN** `POST /api/library` targets an already-saved entry
- **THEN** name and cover refresh in place and no second row is created

#### Scenario: Remove
- **WHEN** the user clicks "In library" on a saved entry
- **THEN** the row is deleted, the response reports `saved: false`, and the button reads "Add to library"

### Requirement: Listing payload

`GET /api/library` SHALL return every saved entry ordered by `added_at` descending with `name`,
`cover`, `plugin`, `slug`, the stat fields (`status`, `total_ch`, `read_ch`, `last_read_at`,
`updated_at`, `status_flags`, `pinned`) and the stored display-detail fields (`authors`, `artists`,
`tags`). It SHALL NOT return a `has_update` field; the New state is bit0 of `status_flags`.

#### Scenario: Detail fields travel with the listing
- **WHEN** a saved entry has stored detail
- **THEN** `authors`, `artists` and `tags` appear in the same object as its name and cover

#### Scenario: New state without has_update
- **WHEN** an entry's chapter count grew since the last rollup
- **THEN** the response carries bit0 set in `status_flags` and no `has_update` field

### Requirement: Source duplicate signal

On save and on detail refresh, the saved row SHALL store the plugin's `synopsis` and `alt_titles`
(verbatim JSON array) from the detail envelope. Enrichment values SHALL NOT be used as the duplicate
signal.

#### Scenario: Synopsis stored on refresh
- **WHEN** refresh returns `detail.synopsis`
- **THEN** the saved row stores it and the clustering inputs update

### Requirement: Duplicate clustering

The store SHALL maintain an FTS5 index over saved `name`, `alt_titles` and `synopsis`, and on
save/unsave/refresh SHALL recompute `dup_group` (0 = unique) from FTS5 candidates scored by token
Dice similarity ≥ 0.6 in connected components. When FTS5 is unavailable `dup_group` SHALL stay 0 and
the duplicate filter SHALL report no duplicates, without crashing.

#### Scenario: Cross-plugin duplicates grouped
- **WHEN** two entries from different plugins share normalised alt-title tokens and a similar synopsis
- **THEN** both rows carry the same non-zero `dup_group`

#### Scenario: FTS5 unavailable
- **WHEN** the driver build lacks FTS5
- **THEN** every `dup_group` is 0, the duplicate filter renders empty, and no error surfaces to the user

### Requirement: Grid contract

The home/library grid SHALL render saved entries as cover cards linking to
`/view/manga/:plugin/:slug`, paginated client-side at 12 per page, with each card carrying
`data-dup`, `data-status-flags` and `data-status`. An empty library SHALL render the empty state, not
the search-cache grid, even when legacy `title` rows exist. In saved mode the grid SHALL ignore
`?q=`/`?page=` server-side, while `?plugin=` pre-filters client-side.

#### Scenario: Saved entry appears
- **WHEN** a manga is saved and the home page is opened
- **THEN** its cover card renders with its stat fields in the SSR context

#### Scenario: Library versus search cache
- **WHEN** `title` holds seed rows but nothing is saved
- **THEN** the empty state renders instead of a grid

#### Scenario: Deep-link plugin filter
- **WHEN** the home page opens with `?plugin=<id>`
- **THEN** the grid starts filtered to that plugin client-side

### Requirement: Category filter uses the stored bit

The grid's category filter (ongoing, hiatus, canceled, completed, done, new) SHALL gate on the
matching `status_flags` bit, never on string comparison of `data-status`. The "Complete" filter
SHALL use bit5.

#### Scenario: Complete filter
- **WHEN** the user activates "Complete"
- **THEN** only cards whose `data-status-flags` has bit5 set remain

#### Scenario: Unknown status renders neutral
- **WHEN** an entry's source status is outside the known categories
- **THEN** no category bit is set, the raw text is shown on the detail page, and the card renders a neutral chip

### Requirement: Duplicate filter clusters

The "duplicate" chip SHALL show only rows with non-zero `dup_group`, ordered by `dup_group` then
name, so each cluster renders contiguously.

#### Scenario: Adjacent duplicates
- **WHEN** three cluster members exist among fifty entries and the chip is active
- **THEN** exactly those three render consecutively

### Requirement: Dim toggle

`POST /api/display/dim-blur` SHALL set or clear only bit6 on a saved entry, `GET /api/display/dim-list`
SHALL return the saved entries with bit6 set, and a toggle for an unsaved manga SHALL be a no-op.

#### Scenario: Dim persists
- **WHEN** a saved entry is dimmed and later refreshed
- **THEN** bit6 stays set and the card keeps its dim treatment

### Requirement: Pin

`POST /api/library/pin` SHALL set or clear `pinned` on a saved entry. Pinned entries SHALL render in
a separate Pinned section of the grid, in `added_at` order, while remaining present in the main list
filtering rules.

#### Scenario: Pin moves the card
- **WHEN** a saved entry is pinned
- **THEN** the next render shows it under Pinned and the unpinned list is unchanged otherwise

### Requirement: Detail toggle hydration

The detail page SHALL initialise its save toggle from library membership (`GET /api/library`
matching plugin+slug), not from a server-rendered `saved` field. A failed toggle SHALL surface on the
page's existing error line and leave the label consistent with the last known state.

#### Scenario: Saved manga reloaded
- **WHEN** the detail page for a saved manga loads
- **THEN** the button shows "In library" on first paint after membership resolves

### Requirement: Search form

The search page SHALL offer only enabled plugins as sources, a query field, an optional genre row
sourced from the selected plugin's `genres` method, and a paginated result grid with a total count.
State SHALL live in the URL (`?plugin=`, `?q=`, `?genre=`, `?page=`) so a search is linkable and
reloadable. A disabled plugin SHALL NOT be selectable; a direct API call to it SHALL return the
disabled error.

#### Scenario: Disabled plugin is not offered
- **WHEN** a plugin is disabled
- **THEN** the source picker omits it and the page falls back to another enabled source or the empty prompt

#### Scenario: Linkable search
- **WHEN** a search runs with plugin, query, genre and page
- **THEN** reloading that URL reproduces the same request and results

#### Scenario: Inline states
- **WHEN** the query yields no results, or the plugin call fails
- **THEN** the grid shows the empty or error state respectively, without navigating away

### Requirement: Search results are cached and counted

Search SHALL serve from the write-through `api_cache` inside its TTL and SHALL return the plugin's
`total` when the plugin reports it, falling back to the page length when it does not. A cache hit
SHALL NOT call the plugin.

#### Scenario: Repeat query
- **WHEN** the same search repeats inside the TTL
- **THEN** the plugin is not invoked and the same payload is returned

#### Scenario: Missing total
- **WHEN** the plugin omits a total count
- **THEN** the response reports the number of items on the page and the UI shows a next-page control based on a full page
