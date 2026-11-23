# Tasks: search-to-master

## 1. Search result preparation

- [x] 1.1 Add `Service.Prepare` to fetch detail and chapters into `api_cache` without materializing — verify no master row is created.
- [x] 1.2 Add `?fetch=1` to search result links and trigger Prepare only for that parameter — verify plain detail reads make no call.

## 2. Save promotion

- [x] 2.1 Add `Service.PromoteCached` to copy cached detail JSON into the master row — verify pinned/dim flags survive.
- [x] 2.2 Call PromoteCached from the library Save handler before setting `added_at` — verify metadata reaches the master row on Save and a cache-less Save still works.

## 3. Close-out

- [x] 3.1 Validate, merge, archive and update the porting order.

Lane gate: `make test lint`, `go vet ./...`, `CGO_ENABLED=0 go build ./...`.
