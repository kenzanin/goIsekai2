# Design: temp-stage-search-cache

## D1. Two lifetimes

`stage_detail` is a TEMP table on the single SQLite connection: process lifetime, no TTL, empty after restart.
`api_cache` keeps its persistent TTL for search responses and plugin responses.

## D2. Stage contents

One row per `(plugin_id, slug)` holds raw `detail_json`, `chapters_json` and `fetched_at`. `StageSearchResult` writes it; `Detail` and `StagedChapters` read it; `PromoteCached` copies it into `saved_manga`/`chapters` and deletes it.

## D3. Search cache

`GET /api/search` hashes `(plugin, query, page)` and checks `api_cache` before calling the plugin. A hit writes the stored JSON envelope directly; a miss calls the plugin once and stores the raw response with TTL.

## D4. JSON1 view

`api_cache_view` projects `params_json` and `payload` into `query_text`, `ok`, `total`, `bytes` etc. It is created at store open and never duplicates payload storage.

## D5. SQL boundary

All SQL literals live in `internal/store/sql`; store and HTTP layers call typed helpers only.
