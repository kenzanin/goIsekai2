# detail Specification

## Purpose

Defines the manga detail page and everything that feeds it: the SSR shell plus JSON hydration, the
saved-row read path, stale-serve and forced refresh, the separate enrichment registry that suggests
metadata and covers from other sources, field-level merge that never overwrites the source plugin's
values without an explicit user pick, cover/related wizards, and the accessibility rules for the
page's icon-only controls.

Route naming note (recorded divergence): the baseline funnels several plugin calls through one
`/api/plugins` action-dispatch endpoint; V2 serves dedicated JSON routes with the same payloads
(`/api/detail`, `/api/chapters`, `/api/enrichment*`, `/api/related`). The page routes
(`/`, `/library/grid`, `/search`, `/plugins`, `/logs`, `/view/manga/:plugin/:slug`, `/read/...`) stay
as they are. This is a deliberate divergence and must be recorded as a decision concept.

## Requirements

### Requirement: Detail page rendering

The detail shell SHALL render stored metadata and chapter pages, at most 20 chapter rows per `?page=`, with placeholders for missing fields. It SHALL offer Fetch/Force and a Refresh candidates action; neither metadata nor candidates SHALL be fetched merely by rendering the page.

#### Scenario: Page renders from plugin data
- **WHEN** a user opens a known title with stored data
- **THEN** the page shows the stored cover, title, status, tags and chapter list

#### Scenario: Missing fields degrade per field
- **WHEN** stored metadata omits optional fields
- **THEN** those fields show placeholders and the rest of the page still renders

#### Scenario: Long chapter list paginates
- **WHEN** the chapter list holds 45 entries
- **THEN** page 1 renders 20 rows and the pager reads 1/3

#### Scenario: Short list hides pager
- **WHEN** the chapter list holds 12 entries
- **THEN** all 12 rows render and no pager is emitted

#### Scenario: Refresh enrichment is explicit
- **WHEN** the user activates Refresh candidates
- **THEN** the page posts to the refresh route and renders candidates only afterwards

### Requirement: Detail and chapters APIs

`GET` routes SHALL remain store-only. A search result link MAY carry `?fetch=1`, an explicit user action that fetches detail and chapters into `api_cache` as raw JSON without writing the master row. The route SHALL then render stored data; a cold fetch failure SHALL still render placeholders.

#### Scenario: Saved row read
- **WHEN** a saved manga's detail is requested while the plugin is unreachable
- **THEN** the stored metadata is returned with `cached` and no plugin call is made

#### Scenario: Cold miss materializes
- **WHEN** no master row exists for a plugin/slug pair
- **THEN** a placeholder payload with the slug as name is returned and `saved` is false

#### Scenario: Disabled plugin still serves the row
- **WHEN** a saved manga's plugin is disabled
- **THEN** detail serves from the row while `POST /api/refresh` answers `plugin disabled`

#### Scenario: Cold detail does not fetch
- **WHEN** a cold title is opened without `fetch=1`
- **THEN** no plugin or enrichment source is contacted

#### Scenario: Stale detail still renders
- **WHEN** the stored row is older than its updater
- **THEN** the row still renders and the scheduler, not the reader, marks it stale

#### Scenario: Search result stages JSON
- **WHEN** a search result opens detail with `?fetch=1`
- **THEN** detail and chapter envelopes are stored in `api_cache` and no master row is created

### Requirement: Stale-serve and unknown titles

Both APIs SHALL serve cached data with `cached: true` when the plugin call fails, SHALL answer 404
with `{"ok":false,"error":"unknown title"}` for an unresolvable plugin/slug pair, and SHALL never
propagate a plugin error as a host crash.

#### Scenario: Upstream failure uses cache
- **WHEN** the plugin call fails and a previous response exists
- **THEN** the cached envelope is returned with `cached: true`

#### Scenario: Unknown title guard
- **WHEN** the slug does not resolve
- **THEN** the response is 404 with the unknown-title error body

### Requirement: Forced refresh

`POST /api/refresh/:plugin/:slug` SHALL call the detail and chapters methods and SHALL overwrite both
cache entries only after both envelopes validate, returning `{"ok":true}`. On failure it SHALL answer
502 naming the failed method and leave both entries untouched. In every outcome it SHALL stamp
`update_attempts` with `ok` or `error` so the scheduler treats the manual attempt as fresh.

#### Scenario: Success overwrites
- **WHEN** both envelopes validate
- **THEN** both entries are replaced, `{"ok":true}` is returned, and the attempt is stamped `ok`

#### Scenario: Failure leaves cache intact
- **WHEN** one method fails or an envelope is invalid
- **THEN** the response is 502 naming the method, the cache is unchanged, and the attempt is stamped `error`

#### Scenario: Manual refresh defers the scheduler
- **WHEN** a manual refresh completes with either outcome
- **THEN** the auto-update scheduler skips that title until its interval elapses

### Requirement: Enrichment registry

The host SHALL load enrichment plugins from `data/enrichment/<id>/` as a registry separate from
scrapers, declaring capability `enrichment` and serving only the `enrich` method on the same runtime
and host tools. Envelope keys other than the enrichment contract SHALL be ignored.

#### Scenario: Loaded as an enrichment source
- **WHEN** `data/enrichment/mangadex/main.lua` exists at boot
- **THEN** it is available as an enrichment source and is not listed among scraper plugins

#### Scenario: Chapter keys ignored
- **WHEN** an enrichment envelope carries chapters or pages
- **THEN** only the enrichment fields are consumed

### Requirement: Two fetch paths

For saved manga only, the detail page SHALL offer independent metadata and cover fetch paths, each
with a normal and a force action. Normal SHALL serve the stored `fetched` payload per source when
present, else fetch once; force SHALL always call the plugin and replace that source's `fetched`
entry, stamped with its fetch watermark.

#### Scenario: Normal uses the stored payload
- **WHEN** `fetched` exists and the user taps Fetch detail
- **THEN** no enrichment plugin is called and the stored payload is shown

#### Scenario: Force always fetches
- **WHEN** the user taps the force action
- **THEN** the plugin is called regardless of stored state and `fetched` is replaced

#### Scenario: First fetch
- **WHEN** no stored payload exists
- **THEN** the normal tap fetches once and stores the result

### Requirement: Candidates are user-selected

Fetched candidates from all sources SHALL be listed grouped per source in source order, with
duplicates preserved, and nothing SHALL change until the user picks. Recommendations SHALL be capped
at 10 per source and cover candidates at 5 per source; a source missing a field simply contributes no
candidate for it.

#### Scenario: Cross-source list
- **WHEN** two enrichment plugins return candidates
- **THEN** both groups render, duplicates included, and the displayed detail is unchanged

#### Scenario: Caps apply
- **WHEN** a source returns 25 recommendations and 9 covers
- **THEN** 10 recommendations and 5 covers are offered

### Requirement: Field merge and tag handling

A chosen non-tag field SHALL replace the displayed source value. Tags SHALL render as
`(source ∪ tags_add) − tags_hidden`, where `tags_hidden` hides without deleting candidates. No
chosen value SHALL be applied without an explicit selection.

#### Scenario: Tag select then hide
- **WHEN** the user selects a fetched tag and later hides it
- **THEN** that tag disappears while the other selected tags still render

### Requirement: Cover wizard

Fetch covers SHALL open a wizard of candidate cover URLs; choosing one SHALL persist it as the
displayed cover, served through `/cache/img` so repeat renders hit disk. Tapping fetch twice for an
already cached candidate SHALL not reach the source site.

#### Scenario: Chosen cover is displayed
- **WHEN** the user picks a candidate
- **THEN** the detail page shows it and the image is served from the cache path

#### Scenario: Repeat tap hits disk
- **WHEN** fetch covers runs again with one candidate already cached
- **THEN** that image is served locally with no upstream request

### Requirement: Reset cover and reset detail

**Reset cover** SHALL remove the chosen cover override and restore the `saved_manga` cover. **Reset
detail** SHALL remove the chosen overrides for title, synopsis, artist, author, `tags_add` and
`tags_hidden` and restore the source-plugin state and `saved_manga` name. Neither reset SHALL delete
or modify `fetched`.

#### Scenario: Reset after merge
- **WHEN** the user taps Reset detail after choosing values
- **THEN** the displayed fields revert to source-plugin values while the fetched payloads remain available

### Requirement: Enrichment persistence

Enrichment state SHALL live in `title_meta` keyed by `(plugin, slug)` with `fetched` (JSON array, one
raw payload per source plus its watermark) and `chosen` (JSON object). `fetched` SHALL NOT expire.
`chosen` SHALL be overwritten in full on save. Enrichment writes SHALL NEVER touch `saved_manga` or
`api_cache`; the single exception is the detail resolver materialising its merged, user-visible values
onto the saved row, which is a merge result and never a raw candidate.

#### Scenario: Source cache untouched
- **WHEN** any enrichment fetch or chosen save executes
- **THEN** `saved_manga` and `api_cache` bytes are unchanged

#### Scenario: Resolver materialises the merge
- **WHEN** the resolver runs and merged values differ from the stored ones
- **THEN** the merged display fields are written to the row while `title_meta` keeps candidates and `chosen`

### Requirement: Related reading

Related entries chosen in the detail wizard SHALL persist with the enrichment state and render in a
related section on the detail page, fetched through `GET /api/related/:plugin/:slug` from the chosen
sources. The section SHALL render nothing when no related entry was chosen.

#### Scenario: Related section renders
- **WHEN** the user chooses related titles from a candidate list
- **THEN** the detail page shows them in the related section after reload

#### Scenario: Nothing chosen
- **WHEN** no related entry was chosen
- **THEN** no related section renders and no request is made

### Requirement: Accessible controls

Every icon-only or image-only interactive control on the detail page and the library grid — per-card
plugin control, status filters, dim toggle, reader pager — SHALL expose an accessible name beyond a
`title` attribute, and every focusable control on the dark templates SHALL render a visible focus
indicator.

#### Scenario: Icon-only control is named
- **WHEN** an icon-only control renders
- **THEN** it exposes an accessible name and its glyph and layout are unchanged

#### Scenario: Focus stays visible
- **WHEN** a control receives keyboard focus
- **THEN** a visible focus indicator is drawn

### Requirement: Detail route wiring

The daemon SHALL register detail and enrichment routes before it opens the HTTP port. Saved detail SHALL be served from its stored row without a plugin call, with `cached: true`; a cold saved row SHALL be materialized from a validated envelope first. Forced refresh SHALL replace both detail and chapter entries only when both validate, and SHALL stamp `update_attempts` in every outcome. Enrichment plugins SHALL load from a registry separate from scraper plugins.

#### Scenario: Saved row avoids plugin call
- **WHEN** a saved manga's detail is requested while the plugin is unreachable
- **THEN** the stored metadata is returned with `cached: true` and no plugin call is made

#### Scenario: Refresh is all-or-nothing
- **WHEN** one refresh envelope fails validation
- **THEN** neither cache entry is replaced and the attempt is stamped `error`

#### Scenario: Enrichment stays separate
- **WHEN** `data/enrichment/<id>/main.lua` exists at boot
- **THEN** it is available as an enrichment source and absent from the scraper plugin list

### Requirement: Envelope canonicalization and rematerialization

The host adapter SHALL normalize plugin detail envelopes before writing the master row: `description` maps to
`synopsis`, arrays/tables of authors, artists and tags/genres flatten to strings, and both bare-array and
`{"data": [...]}` shapes are accepted. A materialized unsaved row missing required metadata SHALL be refetched;
pinned, dim and other user-owned flags SHALL survive.

#### Scenario: Plugin uses its own field names
- **WHEN** atsumaru returns `description`, `authors` as an array and `tags` as an array
- **THEN** the master row stores canonical `synopsis`, `authors` and `tags` strings and the detail page renders them

#### Scenario: Stale empty row refreshes
- **WHEN** a previously materialized unsaved row has no synopsis, authors or tags
- **THEN** the host calls the plugin again and replaces only the missing metadata

### Requirement: Image proxying and form save

The detail page SHALL render covers and chapter pages through `/cache/img` instead of raw CDN URLs, parse page
URLs from the reader envelope, and post save/remove as form-encoded data to `/api/library` and
`/api/library/remove`. A save SHALL NOT overwrite already-materialized metadata with the button's minimal
payload.

#### Scenario: CDN blocks the browser
- **WHEN** a cover URL returns 403 directly in the browser
- **THEN** the detail page still renders it through `/cache/img`, which fetches with the host's referer policy

#### Scenario: Save keeps metadata
- **WHEN** the detail save button posts only `{plugin, slug, name, cover}`
- **THEN** the master row's canonical synopsis, authors, tags and chapter links remain intact

### Requirement: Enrichment candidates are cached

On reads, candidates come only from the stored `api_cache` entry. A cache miss returns no candidates and MUST NOT call a source. Only `RefreshCandidates`, invoked by an explicit action or the scheduler, may fetch and store candidates.

#### Scenario: Repeated detail renders do not fetch enrichment
- **WHEN** the same detail page renders twice inside the TTL
- **THEN** the enrichment source is invoked only on the first render

#### Scenario: Expired cache refreshes
- **WHEN** the cached entry has expired
- **THEN** a refresh action calls the sources and replaces the entry

#### Scenario: Read miss shows nothing
- **WHEN** no cached candidate entry exists
- **THEN** the detail page shows the empty enrichment state without a network call

### Requirement: Explicit enrichment refresh

`POST /api/enrichment/refresh/:plugin/:slug` SHALL be the user action that fetches enrichment sources for the stored title and writes the result to `api_cache`. The auto-update scheduler MAY call the same service method. A normal GET SHALL NOT.

#### Scenario: Explicit refresh fills candidates
- **WHEN** the refresh route is called for a title with no cached candidates
- **THEN** sources are called once, results are stored, and later reads show them

#### Scenario: Read routes never refresh
- **WHEN** a GET detail or chapter route is called after the cache expires
- **THEN** no enrichment source is contacted

### Requirement: Search result preparation

The explicit preparation action SHALL write raw detail and chapter envelopes into the process-lifetime TEMP table `stage_detail`, not into persistent `api_cache`. The staged row SHALL be keyed by `(plugin_id, slug)` and SHALL be deleted after promotion.

#### Scenario: Prepare writes JSON only
- **WHEN** the preparation action succeeds for a new slug
- **THEN** `stage_detail` holds both envelopes and `saved_manga` is unchanged

#### Scenario: Prepare failure is non-fatal
- **WHEN** the source is unreachable during preparation
- **THEN** detail renders placeholders and the Save action still works

#### Scenario: Staged row expires with the process
- **WHEN** the daemon restarts
- **THEN** `stage_detail` is empty and a new search-result click stages data again
