# Tasks: add-lua55

## 1. Dependency and decision

- [ ] 1.1 Pin `modernc.org/lua v0.1.0` and its pure-Go transitive versions in go.mod/go.sum — verify `go list -m modernc.org/lua` reports v0.1.0 and `CGO_ENABLED=0 go build ./...` exits 0.
- [ ] 1.2 Record `decisions/lua-runtime` and update `decisions/dependency-stack` + `architecture/stack` — verify `okf validate knowledge --strict --drift` reports 0 errors.

## 2. Runtime port

- [ ] 2.1 Rewrite `internal/plugin/lua.go` on `lua.VM` with fresh-state `Call`, context interrupt and `handle` invocation — verify `go test ./internal/plugin -run TestLua` exits 0.
- [ ] 2.2 Port `internal/plugin/bindings_lua.go` to `NewTable`/`Value.Set`, keeping the frozen host tool names and envelope — verify a Lua test calls `host.call`, `host.text.clean`, `host.http.get_json` and `host.plugin.param`.
- [ ] 2.3 Port Lua-to-Go and Go-to-Lua conversion for `int64`, `float64`, arrays and maps — verify a numeric-index table round-trips as a JSON array and `json.decode` yields a Lua table.
- [ ] 2.4 Port the plugin-local `require` shim with directory confinement and per-VM cache — verify `require("genres")` works and `require("../outside")` fails.
- [ ] 2.5 Add sandbox and budget tests — verify `os`/`io`/`debug` are unavailable, `ffi`/`bit` fail explicitly, and a runaway loop is interrupted.

## 3. Full suite and E2E

- [ ] 3.1 Remove `gopher-lua` from go.mod — verify no source file imports `github.com/yuin/gopher-lua`.
- [ ] 3.2 Run the full gate set — verify `make test lint`, `go vet ./...`, empty `gofmt -l .`, `CGO_ENABLED=0 go build ./...`, `openspec validate add-lua55 --strict` and AAG lint all exit 0.
- [ ] 3.3 Run the live E2E chain on the demo daemon — verify `atsumaru` and `mangafire` complete search, detail, chapters and pages with no empty envelopes.

## Workflow follow-up

- Merge `change/add-lua55` into `main` with `--no-ff`.
- Apply the MODIFIED delta to `openspec/specs/plugin-host/spec.md`.
- Archive the change as `2026-11-22-add-lua55` after the gates pass.
