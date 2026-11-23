# Proposal: temp-stage-search-cache

## Why

Staged search-result JSON currently lives in `api_cache`, which is persistent and TTL-managed. The user rule is
that staged detail/chapters data only needs process lifetime, while repeated keyword searches must reuse cached
results. A TEMP table and a hash-keyed search cache fit those two lifetimes precisely.

## What Changes

- Move staged detail+chapters envelopes into a SQLite TEMP table `stage_detail`, created lazily on the single connection; it disappears on restart.
- Keep `api_cache` for search responses keyed by `(plugin, hash(query,page))` with TTL, so repeated keyword searches do not call the plugin.
- Add a SQLite JSON1 view `api_cache_view` for querying cached payloads with `json_extract`.
- Promote staged JSON to `saved_manga` + chapter rows on Save, then delete the staged pair.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `detail` — staged search-result source is the TEMP table.
- `library` — plugin search results are cache-backed.
- `store` — JSON1 view over `api_cache`.

## Impact

- Code: `internal/store/sql/stage.go`, `internal/store/stage.go`, `internal/store/sql/views.go`, `internal/detail/service.go`, `internal/http/server.go`.
- No new dependency; the view is a V2 convenience, not a frozen baseline table.

## Baseline

 cppIsekai had neither a process-lifetime stage nor a search response cache; both are V2 local-first optimizations.
