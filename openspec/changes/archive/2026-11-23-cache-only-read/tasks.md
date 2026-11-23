# Tasks: cache-only-read

## 1. Read path

- [x] 1.1 Make `Service.Detail` store-only and return placeholders for missing rows — verify a stale `api_cache` entry is ignored.
- [x] 1.2 Add `StoredChapters` and remove live-plugin fallbacks from `renderDetail` and `listChapters` — verify GET routes make no plugin call.
- [x] 1.3 Make `Candidates` store-only and add `RefreshCandidates` — verify a cold read returns no candidates and does not call a source.

## 2. Explicit refresh

- [x] 2.1 Add `POST /api/enrichment/refresh/:plugin/:slug` — verify it fetches once and stores candidates under the `(kind,title)` key.
- [x] 2.2 Resolve candidate application title from the stored row — verify apply uses only stored candidates.
- [x] 2.3 Add the Refresh candidates button to the detail page — verify the generated HTML posts to the new route.

## 3. Close-out

- [x] 3.1 Record `decisions/cache-only-read-path` in OKF and keep it related to the master-data decisions.
- [x] 3.2 `openspec validate cache-only-read --strict` clean; archive the change and sync specs.

Lane gate: `make test lint`, `go vet ./...`, `CGO_ENABLED=0 go build ./...`.
