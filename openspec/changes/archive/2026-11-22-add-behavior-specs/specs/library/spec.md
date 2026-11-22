# Spec Delta

## MODIFIED Requirements

### Requirement: Save and unsave

`POST /api/library` SHALL accept both JSON and `application/x-www-form-urlencoded` bodies (htmx posts form
data). It SHALL set the master row's `added_at` to the current unix time, which is the library-membership
marker; an already-materialized row SHALL keep its canonical metadata and its original `added_at` when
re-saved. `POST /api/library/remove` SHALL clear `added_at` to zero and SHALL NOT delete the `saved_manga`
row, because that row is the single master record and may be unsaved. Both routes SHALL respond with
`{"saved": true|false}`, SHALL be idempotent, and SHALL publish `library.saved` / `library.removed` after the
commit.

#### Scenario: Add to library
- **WHEN** a user adds an unsaved manga from the detail page
- **THEN** `added_at` becomes positive, the response reports `saved: true`, and the button reads "Remove from library"

#### Scenario: Re-add refreshes
- **WHEN** `POST /api/library` targets an already-saved entry
- **THEN** name and cover refresh in place and no second row is created

#### Scenario: Re-add preserves metadata
- **WHEN** `POST /api/library` targets an already-saved entry with a minimal htmx payload
- **THEN** the stored name, cover, stats and original `added_at` are preserved and no second row is created

#### Scenario: Remove
- **WHEN** the user clicks "Remove from library" on a saved entry
- **THEN** `added_at` becomes zero, the row remains as the master record, the response reports `saved: false`, and the button reads "Save to library"

#### Scenario: Form-encoded post
- **WHEN** htmx posts `plugin`, `slug`, `name` and `cover` as form data
- **THEN** the route accepts it and responds with the same `{"saved": true}` envelope as a JSON post

### Requirement: The upsert never stores a placeholder name

`POST /api/library` SHALL skip writing `name` when the submitted value is empty or equal to the slug, and
SHALL NOT overwrite a non-empty stored name with such a value. A re-add while the plugin is unreachable SHALL
only refresh `cover` and the membership marker; it cannot degrade a good title into a slug.

#### Scenario: Plugin unreachable on re-add
- **WHEN** a saved manga without stored detail is re-added while its plugin is unreachable and the page submits the slug as the name
- **THEN** the stored name is left as-is and only cover and `added_at` are refreshed

### Requirement: Detail toggle hydration

The detail page SHALL initialise its save/remove control from the `saved` boolean in the `/api/detail`
payload, not from a separate `/api/library` membership fetch. A failed toggle SHALL surface on the page's
existing error line and leave the label consistent with the last known state.

#### Scenario: Saved manga reloaded
- **WHEN** the detail page for a saved manga loads
- **THEN** the button renders "Remove from library" on first paint because the detail payload reports `saved: true`

## ADDED Requirements

### Requirement: Plugin-backed search API

The daemon SHALL expose `GET /api/search?q=<text>&plugin=<id>&genre=<optional>&page=<n>` as the API-first
search surface used by the search page. It SHALL call the selected enabled plugin's `search` method through
the host, cache the response in `api_cache` inside its TTL, and return the plugin envelope wrapped as
`{"ok": true, "data": {...}}`. A disabled or unknown plugin SHALL answer `{"ok": false, "error": "...",
"plugin": "..."}` with a non-2xx status, and no plugin session SHALL be created.

#### Scenario: Search page delegates to the API
- **WHEN** the `/search` page runs a query against an enabled plugin
- **THEN** it requests `/api/search` rather than searching the local store and renders `data.items` from the response

#### Scenario: Disabled plugin rejected
- **WHEN** `GET /api/search` targets a disabled plugin
- **THEN** the response reports `ok: false` with the plugin id and the backend is never touched
