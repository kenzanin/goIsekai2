# Spec Delta

## MODIFIED Requirements

### Requirement: Plugin-backed search API

`GET /api/search` SHALL hash `(plugin, query, page)` into an `api_cache` key. A live hit SHALL return the stored envelope without calling the plugin; a miss SHALL call the plugin once and store the raw response with TTL. Disabled plugins SHALL still be rejected before any cache or backend access.

#### Scenario: Search page delegates to the API
- **WHEN** the `/search` page runs a query against an enabled plugin
- **THEN** it requests `/api/search` rather than searching the local store and renders `data.items`

#### Scenario: Disabled plugin rejected
- **WHEN** `GET /api/search` targets a disabled plugin
- **THEN** the response reports `ok: false` with the plugin id and the backend is never touched

#### Scenario: Repeated keyword reuses cache
- **WHEN** the same plugin, query and page are requested inside the TTL
- **THEN** the cached envelope is returned and the plugin is not called

#### Scenario: Cache miss calls plugin once
- **WHEN** no live cache entry exists
- **THEN** one plugin call fills the cache and subsequent equal requests hit it
