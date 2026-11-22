# Spec Delta

## MODIFIED Requirements

### Requirement: Detail and chapters APIs

`GET /api/detail/:plugin/:slug` serves the `saved_manga` master row without a plugin call when one exists,
flagged `cached`, and carries the `enrichment` block, `chosen_cover` and a `saved` boolean derived from
`added_at > 0`. For an unsaved title the host SHALL call the plugin through the stale layer, materialize the
canonical fields into a new master row with `added_at = 0`, and serve that row; opening a title therefore never
adds it to the library. `GET /api/chapters/:plugin/:slug` SHALL keep live-or-stale behaviour in both cases,
and a cold detail response SHALL also fill chapters and pages when the plugin supplies them.

#### Scenario: Saved row read
- **WHEN** a saved manga's detail is requested while the plugin is unreachable
- **THEN** the stored metadata is returned with `cached: true`, `saved: true`, and no plugin call is made

#### Scenario: Cold miss materializes
- **WHEN** no master row exists for a plugin/slug pair
- **THEN** the validated plugin envelope is written to `saved_manga` with `added_at = 0`, chapters are filled, and `saved: false` is returned

#### Scenario: Disabled plugin still serves the row
- **WHEN** a saved manga's plugin is disabled
- **THEN** detail serves from the row while `POST /api/refresh` answers `plugin disabled`

## ADDED Requirements

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
