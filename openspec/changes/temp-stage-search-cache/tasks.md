# Tasks: temp-stage-search-cache

## 1. TEMP staging

- [x] 1.1 Add `stage_detail` TEMP table helpers under `internal/store/sql` and store wrappers — verify put/get/delete and process-lifetime behavior.
- [x] 1.2 Route `StageSearchResult`, staged detail reads and `PromoteCached` through the TEMP table — verify promotion deletes the staged row.

## 2. Search cache and JSON1 view

- [x] 2.1 Cache `/api/search` responses by `(plugin, query, page)` in `api_cache` with TTL — verify a repeated keyword search makes no plugin call.
- [x] 2.2 Add `api_cache_view` using JSON1 — verify the view exists and exposes `json_extract` columns.

## 3. Close-out

- [x] 3.1 Validate, gate, merge, archive and update decisions.

Lane gate: `make test lint`, `go vet ./...`, `CGO_ENABLED=0 go build ./...`.
