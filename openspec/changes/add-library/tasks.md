# Tasks: add-library

## 1. Service and routes (parallel lanes 1A/1B)

- [x] 1A.1 Implement `internal/library/service.go`: Save, Remove, List, SetFlag, ToggleDim, TogglePin over typed store methods — verify tests assert idempotency, placeholder-name protection and bit preservation.
- [x] 1A.2 Publish `library.saved` / `library.removed` after commit only — verify a recorder test asserts no event on rejected saves and correct payload on success.
- [x] 1B.1 Add `GET/POST /api/library` and `POST /api/library/remove` handlers in `internal/http` with the frozen JSON shapes — verify httptest covers add, re-add, remove and error bodies.
- [x] 1B.2 Add category, duplicate and `?plugin=` query handling to the listing handler — verify filtered responses match `status_flags` and `dup_group`.

## 2. Grid and search (parallel lanes 2A/2B)

- [x] 2A.1 Add the templ library grid component with 12-per-page Alpine pagination, data attributes, empty state and `?plugin=` filter — verify httptest asserts card attributes and empty state.
- [x] 2A.2 Wire the grid route at `/` and `/library` using the implemented-pages list — verify nav parity and that saved mode ignores `?q=`/`?page=` server-side.
- [x] 2B.1 Add the search form listing only enabled plugins from `Host.Rows` — verify a disabled plugin is absent while its saved titles still list.
- [x] 2B.2 Add FTS5 local search with LIKE fallback and `api_cache` write-through — verify ranked results, prefix match and a hit avoiding a plugin call.

## 3. Gates

- [x] 3.1 Run the full gate set — verify `make test lint`, `go vet ./...`, `gofmt -l .` empty, `openspec validate add-library --strict` and `CGO_ENABLED=0 go build ./...` all exit 0.

## Workflow follow-up

- Merge `change/add-library` into `main` with `--no-ff`.
- Archive the change with `openspec archive add-library -y` after the gates pass.
