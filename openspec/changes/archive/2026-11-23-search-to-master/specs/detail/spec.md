# Spec Delta

## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Search result preparation

The host SHALL expose one explicit preparation action for search results. It SHALL write raw detail and chapter JSON to `api_cache` in the main SQLite database with a TTL. A normal library read SHALL NOT invoke it.

#### Scenario: Prepare writes JSON only
- **WHEN** the preparation action succeeds for a new slug
- **THEN** `api_cache` holds both envelopes and `saved_manga` is unchanged

#### Scenario: Prepare failure is non-fatal
- **WHEN** the source is unreachable during preparation
- **THEN** detail renders placeholders and the Save action still works
