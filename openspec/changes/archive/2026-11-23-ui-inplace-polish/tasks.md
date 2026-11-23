# Tasks: ui-inplace-polish

## 1. In-place navigation

- [x] 1.1 Add the shared swap target and htmx attributes to nav, search, detail, chapter and reader links — verify by clicking through the live daemon.
- [x] 1.2 Replace every `location.reload()` with an in-place main refresh — verify save/remove, fetch/force/reset, enrichment apply and chapter flags.

## 2. Polish

- [x] 2.1 Long page window with ellipsis on library, detail and chapter pagers — verify a 14-page range renders `1 2 3 4 … 10 11 12 13 14`.
- [x] 2.2 Move the saved indicator below the synopsis as `highlight` and add Start/Continue reading — verify with a saved title that has a stored position.

## 3. Close-out

- [x] 3.1 Validate, merge and archive.

Gate: `make test lint`, `CGO_ENABLED=0 go build ./...`.
