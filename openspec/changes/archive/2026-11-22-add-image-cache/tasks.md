# Tasks: add-image-cache

## 1. Cache core (parallel lanes 1A/1B)

- [x] 1A.1 Implement `internal/cache/layout.go`: v3 path builders for cover, page and api payload keys — verify a table test pins exact paths and zero-padded pNNN.
- [x] 1A.2 Implement `internal/cache/lru.go`: shared index over files under data/cache with cap from config, mtime rebuild, eviction with sidecars — verify cap/eviction and restart tests.
- [x] 1B.1 Implement `internal/cache/store.go`: atomic writes, single-flight per key and short negative cache — verify concurrent misses fetch once and a truncated body leaves no file.
- [x] 1B.2 Implement transcode dispatch (webp default, none passthrough, avif/jxl wasm) with undecodable fallback — verify extension/content-type tests and no cgo.

## 2. Serve and prefetch (parallel lanes 2A/2B)

- [x] 2A.1 Add `GET /cache/img` route with full placement validation, hit headers, ratio sidecar and offline 404 — verify httptest covers 400, hit, miss and offline.
- [x] 2A.2 Wire config `cache_max_mb` / `image_format` / `cover_max_px` into the cache — verify changing format keeps old files and applies to new writes.
- [x] 2B.1 Implement cover resize and JSON ratio sidecar from decoded dimensions — verify a 2000x3000 fixture stores at most 800px with correct ratio, ignoring client ratio.
- [x] 2B.2 Implement sliding 10-page prefetch window crossing chapters, skipping flagged chapters and cached keys — verify boundary spill and skip reroute tests.

## 3. Gates

- [x] 3.1 Run the full gate set — verify `make test lint`, `go vet ./...`, `gofmt -l .` empty, `openspec validate add-image-cache --strict` and `CGO_ENABLED=0 go build ./...` all exit 0.

## Workflow follow-up

- Merge `change/add-image-cache` into `main` with `--no-ff`.
- Archive the change with `openspec archive add-image-cache -y` after the gates pass.
