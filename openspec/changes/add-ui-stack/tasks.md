# Tasks: add-ui-stack

## 1. Toolchain and local assets

- [x] 1.1 Add `github.com/a-h/templ` v0.3.1118 to `go.mod` and pin the CLI (`go install ...@v0.3.1118`) in the Makefile — verify `go list -m github.com/a-h/templ` reports the pinned version and `CGO_ENABLED=0 go build ./...` stays green.
- [x] 1.2 Add `make generate` (templ generate + Tailwind v4.3.5 CLI, npm fallback) and a `make lint` freshness check — verify `make generate && make lint` succeeds twice in a row with no diff.
- [x] 1.3 Vendor minified Alpine v3.17.5 and htmx v2.0.11 under `internal/http/static/vendor/` and embed them with the existing asset FS — verify `go:embed` finds both files and no URL in rendered HTML starts with `http://` or `https://`.

## 2. Typed settings page

- [x] 2.1 Create `internal/http/components/base.templ` and `settings.templ` with the active-page nav, provenance rows, disabled env/flag inputs and restart badges — verify generated `*_templ.go` files exist and `templ fmt` reports no changes.
- [x] 2.2 Render `/settings` through the templ component instead of `html/template` — verify the existing nav/provenance httptest cases pass unchanged.
- [x] 2.3 Replace the hand-written CSS with the compiled Tailwind stylesheet and keep semantic markup — verify `/settings` HTML references only local CSS and retains `<label>` per input.

## 3. Interaction and contract

- [x] 3.1 Move dirty-field tracking and status banners into an Alpine component — verify a test renders `x-data` state and no inline `fetch()` remains in the page source.
- [x] 3.2 Wire the form to `POST /config` with htmx and swap only the status fragment — verify an httptest posts one changed field, the response has `{"ok":true,"written":1}`, and the page keeps `GET/POST /config` JSON shapes byte-identical.
- [x] 3.3 Add a no-CDN browser-level check against the rendered page using the local asset server — verify all asset requests resolve to the daemon and none reach an external host.
- [x] 3.4 Re-run the archived config end-to-end expectations against the ported page — verify `TestDaemonEndToEnd`, `TestGetConfigReflectsFileEdits` and `TestPostConfigMergeSaveAndReject` still pass with no test edits.

## 4. Documentation and gates

- [x] 4.1 Record the stack decision and refresh `knowledge/architecture/stack.md`, `overview.md` and `decisions/http-framework.md` — verify `okf validate knowledge --strict --drift` reports 0 errors/gates/orphans.
- [x] 4.2 Run the full gate set — verify `make generate`, `make test lint`, `go vet ./...`, `gofmt -l .` empty and `openspec validate add-ui-stack --strict` all exit 0.
- [ ] 4.3 Update the porting-order record when this change is archived — verify `knowledge/process/porting-order.md` marks `add-ui-stack` archived and `add-store` active.

## Workflow follow-up

- Merge `change/add-ui-stack` into `main` with `--no-ff`.
- Archive the change with `openspec archive add-ui-stack -y` after the gates pass.
