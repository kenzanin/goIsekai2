# Spec Delta

## MODIFIED Requirements

### Requirement: Detail and chapters APIs

`GET /view/...`, `GET /api/detail/...` and `GET /api/chapters/...` SHALL read only stored rows and sidecars. They MUST NOT call a plugin or enrichment source. Missing metadata SHALL render as placeholders; an explicit Fetch action or the update scheduler performs all website reads.

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
- **WHEN** a cold title is opened for the first time
- **THEN** no plugin or enrichment source is contacted

#### Scenario: Stale detail still renders
- **WHEN** the stored row is older than its updater
- **THEN** the row still renders and the scheduler, not the reader, marks it stale

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

## ADDED Requirements

### Requirement: Explicit enrichment refresh

`POST /api/enrichment/refresh/:plugin/:slug` SHALL be the user action that fetches enrichment sources for the stored title and writes the result to `api_cache`. The auto-update scheduler MAY call the same service method. A normal GET SHALL NOT.

#### Scenario: Explicit refresh fills candidates
- **WHEN** the refresh route is called for a title with no cached candidates
- **THEN** sources are called once, results are stored, and later reads show them

#### Scenario: Read routes never refresh
- **WHEN** a GET detail or chapter route is called after the cache expires
- **THEN** no enrichment source is contacted
