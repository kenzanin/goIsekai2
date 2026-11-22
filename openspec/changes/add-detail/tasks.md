# Tasks: add-detail

## 1. Service and APIs (parallel lanes 1A/1B)

- [ ] 1A.1 Implement `internal/detail/doc.go + service.go`: saved-row read, cold materialization, stale-serve, validate envelopes — verify httptest-free service tests cover saved, cold and failure.
- [ ] 1A.2 Implement forced refresh with all-or-nothing writes and `update_attempts` stamping — verify success overwrites both entries and failure leaves them intact.
- [ ] 1B.1 Add `GET /api/detail/:plugin/:slug`, `GET /view/manga/:plugin/:slug` and `POST /api/refresh/:plugin/:slug` handlers — verify 404 guard, cached flag and 502 naming the failed method.
- [ ] 1B.2 Add enrichment registry loader over `data/enrichment/<id>/` with only the `enrich` method — verify enrichment sources are absent from scraper lists.

## 2. Enrichment UI and merge (parallel lanes 2A/2B)

- [ ] 2A.1 Add templ detail page with per-field placeholders, read/cbz chapter rows and accessible icon controls — verify rendered HTML carries aria labels and chapter links.
- [ ] 2A.2 Add enrichment candidate list and cover wizard actions with normal/force fetch buttons — verify buttons post to the dedicated routes and a candidate is applied only on explicit pick.
- [ ] 2B.1 Implement field-level candidate merge, tag handling and reset detail/cover actions — verify plugin fields are never overwritten automatically.
- [ ] 2B.2 Add related-reading view from stored metadata — verify a fixture produces links and an empty state.

## 3. Gates

- [ ] 3.1 Run the full gate set — verify `make test lint`, `go vet ./...`, `gofmt -l .` empty, `openspec validate add-detail --strict` and `CGO_ENABLED=0 go build ./...` all exit 0.

## Workflow follow-up

- Merge `change/add-detail` into `main` with `--no-ff`.
- Archive the change with `openspec archive add-detail -y` after the gates pass.
