# Tasks: cache-layout-concept

## 1. Layout

- [x] 1.1 Add namespace, hash-suffixed slug dir, six-digit chapter dir and ratio-in-filename cover naming — verify path tests.
- [x] 1.2 Replace the ratio sidecar write with filename encoding — verify `ReadRatio` parses both forms.

## 2. Cover and prefetch surfaces

- [x] 2.1 Proxy library and search covers through `/cache/img` — verify both HTML pages emit cache URLs.
- [x] 2.2 Store prefetched pages in the image layout and remove `APIPath` — verify no `api/` writes remain.

## 3. Close-out

- [x] 3.1 Validate, merge, archive and update decisions.

Lane gate: `make test lint`, `go vet ./...`, `CGO_ENABLED=0 go build ./...`.
