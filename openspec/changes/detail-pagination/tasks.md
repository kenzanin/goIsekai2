# Tasks: detail-pagination

## 1. Detail page slice and pager

- [x] 1.1 Parse and clamp `page` in `renderDetail` (default 1; non-numeric or out-of-range → 1) and render at most 20 chapter rows — verify a 45-chapter fixture renders rows 1–20 on page 1 and rows 21–40 on page 2.
- [x] 1.2 Emit the shared prev/next + `n/total` pager and omit it when the list fits one page — verify page 1 of 45 shows 1/3 with both links and a 12-chapter title renders no pager.
- [x] 1.3 Leave `/api/chapters/:plugin/:slug` and `/chapters/:plugin/:slug` behavior unchanged — verify the existing route tests pass unmodified.

Lane gate: `go vet ./... && gofmt -l internal && CGO_ENABLED=0 go build ./... && go test ./internal/http/...`.

## 2. Enrichment candidate cache

- [x] 2.1 Cache `Service.Candidates` by `(kind, title)` in `api_cache` with a finite TTL — verify a fake enrichment source is called once across two renders.
- [x] 2.2 Expired entries trigger a fresh source call and replace the payload — verify TTL helper behavior via the store cache test.

## 3. Close-out (main session)

- [x] 3.1 `openspec validate detail-pagination --strict` clean; archive the change and sync specs.
