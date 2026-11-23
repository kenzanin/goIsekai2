# Spec Delta

## MODIFIED Requirements

### Requirement: Save and unsave

`POST /api/library` accepts JSON or form data and SHALL first promote a staged cached detail envelope into `saved_manga`, then set `added_at` as the membership marker. Without a cached envelope it still saves the submitted name/cover. `POST /api/library/remove` clears `added_at` and keeps the row. Both are idempotent and publish domain facts after commit.

#### Scenario: Add to library
- **WHEN** a user adds an unsaved manga from the detail page
- **THEN** `added_at` becomes positive and the response reports `saved: true`

#### Scenario: Re-add refreshes
- **WHEN** `POST /api/library` targets an already-saved entry
- **THEN** name and cover refresh in place and no second row is created

#### Scenario: Re-add preserves metadata
- **WHEN** `POST /api/library` targets an already-saved entry with a minimal htmx payload
- **THEN** stored metadata and original `added_at` are preserved

#### Scenario: Remove
- **WHEN** the user removes a saved entry
- **THEN** `added_at` becomes zero, the row remains, and the response reports `saved: false`

#### Scenario: Form-encoded post
- **WHEN** htmx posts form data
- **THEN** the route accepts it and responds with `{"saved": true}`

#### Scenario: Save promotes cached metadata
- **WHEN** a staged `api_cache` detail envelope exists for the slug
- **THEN** Save copies its canonical fields into `saved_manga` before setting `added_at`
