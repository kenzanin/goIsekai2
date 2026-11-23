# Tasks: reader-defaults

## 1. Defaults and persistence

- [x] 1.1 Pass persisted zoom through the reader bootstrap and default empty mode/zoom to paged/width — verify with a stored non-default position and a fresh position.

## 2. Overlay

- [x] 2.1 Remove the bottom overlay bar and move back-to-detail to the top-right — verify in the browser that no bottom bar renders and the link targets the detail page.

## 3. Close-out

- [ ] 3.1 Validate, merge and archive.

Gate: `make test lint`, `CGO_ENABLED=0 go build ./...`.
