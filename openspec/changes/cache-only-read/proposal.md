# Proposal: cache-only-read

## Why

Manga detail reads still had hidden fetch paths: `Detail` rematerialized from a plugin when metadata was
missing, `Candidates` called enrichment sources on a cache miss, and chapter helpers fell back to a live
plugin call. A GET could therefore spend 8-22 seconds on network work even though the master row was already
cached.

## What Changes

- Make `GET /view/...`, `GET /api/detail/...` and `GET /api/chapters/...` store-only; missing data renders as placeholders.
- Make `Candidates` store-only; a miss returns an empty list.
- Add `POST /api/enrichment/refresh/:plugin/:slug` as the explicit user action that fetches and stores enrichment candidates.
- Add a Refresh candidates button to the detail page and keep Fetch/Force and the auto-update scheduler as the only website fetch paths.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `detail` — read-path purity, explicit enrichment refresh, and placeholder behavior.

## Impact

- Code: `internal/detail/service.go`, `internal/http/server.go`, `internal/http/components/detail.templ`.
- No new dependency; no route removal; the existing Fetch action and scheduler remain unchanged.
- Decision recorded in OKF `decisions/cache-only-read-path`.

## Baseline

 cppIsekai lazily fetched on read; V2 deliberately diverges so local-first reads stay fast and predictable.
