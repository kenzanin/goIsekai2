# Proposal: cache-layout-concept

## Why

The cppIsekai baseline keys image cache paths by a stable plugin namespace, a bounded hash-suffixed manga
directory and a collision-safe chapter folder. V2 drifted to raw plugin ids, unbounded slugs and a short
chapter number. Covers must also be cached for library and search thumbnails, and the unused on-disk API
cache should be dropped now that `api_cache` lives in SQLite.

## What Changes

- Image paths use a display-name-derived namespace with plugin-id fallback, a hash-suffixed slug directory, six-digit chapter folders with a last-eight id suffix, and a ratio encoded in cover filenames.
- Library and search thumbnails both render through `/cache/img`.
- Prefetch writes pages into the structured image cache; the on-disk `api/` tree and `APIPath` are removed.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `image-cache` — path naming, cover ratio and removal of the disk API tree.
- `library` — cover thumbnails proxied through the cache.
- `store` — SQLite `api_cache` is the sole response cache.

## Impact

- Code: `internal/cache`, `internal/http/server.go`, `internal/reader/service.go` and components.
- No new dependency; cached files from the old layout are simply re-fetched under the new paths.

## Baseline

Concepts follow cppIsekai `src/cache/`; exact hashes, hash function and sidecar removal are V2 simplifications.
