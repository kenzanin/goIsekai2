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

`GET /view/manga/:plugin/:slug` SHALL render the detail shell and hydrate metadata and the chapter
list from the JSON APIs. Missing optional fields SHALL render as per-field placeholders instead of
failing the page. Every chapter row SHALL offer a read link and a cbz link, and search result cards
SHALL link into this route.

#### Scenario: Page renders from plugin data
- **WHEN** a user opens a known title
- **THEN** the page shows cover, title, status, tags and the chapter list, each row with read and cbz entries

#### Scenario: Missing fields degrade per field
- **WHEN** the plugin omits optional metadata
- **THEN** those fields show placeholders and the rest of the page still renders

### Requirement: Detail and chapters APIs

`GET /api/detail/:plugin/:slug` for a saved manga SHALL serve the stored `saved_manga` row without
calling the plugin, flagged `cached`, and SHALL still carry the `enrichment` block and
`chosen_cover`. For an unsaved title it SHALL serve the plugin envelope through the stale layer.
`GET /api/chapters/:plugin/:slug` SHALL keep live-or-stale behaviour in both cases.

#### Scenario: Saved row read
- **WHEN** a saved manga's detail is requested while the plugin is unreachable
- **THEN** the stored metadata is returned with `cached` and no plugin call is made

#### Scenario: Cold miss materializes
- **WHEN** a saved manga has no stored detail yet
- **THEN** the envelope is resolved, validated and written to the row before serving

#### Scenario: Disabled plugin still serves the row
- **WHEN** a saved manga's plugin is disabled
- **THEN** detail serves from the row while `POST /api/refresh` answers `plugin disabled`

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
