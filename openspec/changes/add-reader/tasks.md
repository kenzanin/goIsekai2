# Tasks: add-reader

## 1. Service and routes (parallel lanes 1A/1B)

- [x] 1A.1 Implement `internal/reader/service.go`: Pages resolver with exact ch_id selection, stale-list fallback and neighbour numbers/ids — verify tests cover duplicate numbers and failed refresh.
- [x] 1A.2 Implement window math helpers (mode pairs, direction-aware rtl, 10-url window) — verify table tests for odd/even counts and all four modes.
- [x] 1A.3 Implement position upsert preserving absent fields and direct-call prefetch order (next then previous) — verify recorder test shows no event and stored fields unchanged.
- [x] 1B.1 Add `GET /api/pages/:plugin/:slug/:ch` and `POST /reader/position` handlers with frozen JSON shapes — verify httptest covers 404 unknown chapter and partial position update.
- [x] 1B.2 Add `GET /read/:plugin/:slug` route and resolve mode/direction from stored position — verify httptest restores persisted mode and `ch_id`.

## 2. UI controls (parallel lanes 2A/2B)

- [x] 2A.1 Add templ reader component with four modes, overlay, keyboard hints and page jump — verify rendered HTML carries mode/direction data attributes and nav overlay controls.
- [x] 2A.2 Add in-place chapter switch script using `history.pushState` and `popstate` — verify a fixture asserts no document navigation path in the handler contract.
- [x] 2B.1 Add detail chapter list pagination at 20 with skip/read controls — verify a 45-row fixture renders 20 rows and a 1/3 indicator.
- [x] 2B.2 Add `POST /chapter/flags` handler with bit ops for read/skip — verify toggling one bit preserves the other.

## 3. Gates

- [x] 3.1 Run the full gate set — verify `make test lint`, `go vet ./...`, `gofmt -l .` empty, `openspec validate add-reader --strict` and `CGO_ENABLED=0 go build ./...` all exit 0.

## Workflow follow-up

- Merge `change/add-reader` into `main` with `--no-ff`.
- Archive the change with `openspec archive add-reader -y` after the gates pass.
