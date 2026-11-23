# Proposal: search-to-master

## Why

The intended flow is: search result click fetches all metadata and chapters from the site into a JSON cache,
Save copies that cached JSON into the master `saved_manga` row, and every later library read is store-only.
Current code fetched on a separate manual action and wrote the master row before Save.

## What Changes

- Search result links open detail with `?fetch=1`, an explicit user action that prepares detail + chapters JSON in `api_cache` without touching the master row.
- `POST /api/library` promotes the cached JSON into `saved_manga` before setting `added_at`.
- Normal detail and library reads remain store-only, so opening a library title never fetches.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `detail` — search-result preparation and staged JSON cache.
- `library` — Save promotes cached JSON into the master row.

## Impact

- Code: `internal/detail/service.go`, `internal/http/server.go`, `internal/http/components/search.templ`.
- No new dependency, no route removal, no change to the cache tree or plugin ABI.

## Baseline

 cppIsekai fetched lazily; V2 stages fetch on an explicit search-result action and keeps the master row authoritative for reads.
