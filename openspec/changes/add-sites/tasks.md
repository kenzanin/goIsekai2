# Tasks: add-sites

## 1. Validation and reference plugin (parallel lanes 1A/1B)

- [ ] 1A.1 Add folder-id validation to `internal/plugin/manifest.go` — verify a mismatch is rejected with a logged reason.
- [ ] 1A.2 Add a manifest test for missing id, missing manifest and id mismatch — verify each failure path.
- [ ] 1B.1 Port the baseline mangadex plugin into `plugins/lua/mangadex/` with normalized manifest — verify its files load under the Lua runtime.
- [ ] 1B.2 Add a conformance fixture test for canonical search/detail/chapters/pages envelopes — verify integer timestamps and absolute page URLs.

## 2. Logo route and parity (parallel lanes 2A/2B)

- [ ] 2A.1 Add `GET /plugin/{id}/logo` alongside the plural route — verify correct content type and 404 for a missing logo.
- [ ] 2A.2 Add an HTTP test for the logo route — verify bytes and content type.
- [ ] 2B.1 Generate `plugins/PARITY.md` from the baseline `plugins_dev/` inventory — verify every baseline folder is marked ported or omitted with a reason.
- [ ] 2B.2 Install the full ported corpus with `scripts/install-plugins` — verify no plugin folder is lost.

## 3. Gates

- [ ] 3.1 Run the full gate set — verify `make test lint`, `go vet ./...`, `gofmt -l .` empty, `openspec validate add-sites --strict` and `CGO_ENABLED=0 go build ./...` all exit 0.

## Workflow follow-up

- Merge `change/add-sites` into `main` with `--no-ff`.
- Archive the change with `openspec archive add-sites -y` after the gates pass.
