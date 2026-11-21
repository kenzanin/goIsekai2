# Tasks: add-host-tools

## 1. Tool set core

- [x] 1.1 Add `bogdanfinn/tls-client@v1.18.1` and `golang.org/x/net` — verify `go list -m` shows the pinned versions and `CGO_ENABLED=0 go build ./...` exits 0.
- [x] 1.2 Implement `internal/plugin/tools/tools.go`: `ToolSet{store *store.Store, cfg *config.Snapshot, dir string, pluginID string, log *slog.Logger}` with `Fetch`, `StorageGet/Set`, `ConfigGet`, `Log`, `Now` — verify a unit test covers now/log/storage/config with a fake store.
- [x] 1.3 Implement `internal/plugin/tools/http.go`: `Fetch(url string, options map[string]any) (map[string]any, error)` returning `{status,body,headers}` with tls-client, per-call profile/timeout, non-2xx as data — verify httptest covers defaults, POST headers and a 410 body.
- [x] 1.4 Implement `FetchBody` and `GetJSON`/`PostJSON` helpers — verify 410 yields `upstream status 410`, JSON decodes, and an options map with no known keys is treated as headers.
- [x] 1.5 Implement the opt-in per-plugin cookie jar at `data/plugin-state/<id>.cookies.txt` — verify two plugins do not share cookies and a jar survives reconstruction.

## 2. Text, url, regex and html

- [x] 2.1 Implement `internal/plugin/tools/text.go`: `url_encode`, `url_decode`, `json_blob`, `date_epoch`, `clean` — verify a table test covers unicode and invalid input.
- [x] 2.2 Implement `internal/plugin/tools/regex.go`: RE2 `match`, `find`, `capture`, `find_all`, `gmatch`, `replace`, `quote`, `find_index` — verify table-driven cases and an invalid-pattern error.
- [x] 2.3 Implement `internal/plugin/tools/html.go`: `select`, `attr`, `xpath` subset — verify a fixture HTML test returns expected text/attribute values.

## 3. Crypto, signing and file scope

- [x] 3.1 Implement `internal/plugin/tools/crypto.go`: hash/hmac/base64/hex and the documented signing helper — verify known vectors.
- [x] 3.2 Implement `internal/plugin/tools/fs.go`: read-only `fs.read` confined to the plugin folder, rejecting absolute paths and `..` escapes — verify escape attempts fail and a normal read succeeds.

## 4. Runtime bindings

- [x] 4.1 Bind the tool set into Lua as global `host.*` in `internal/plugin/bindings_lua.go` — verify a Lua fixture calls `host.now`, `host.log.info` and `host.storage.set/get` with the same shapes as Go.
- [x] 4.2 Bind the tool set into JS as `host` in `internal/plugin/bindings_js.go` — verify a JS fixture calls `host.now` and `host.http.fetch` against an httptest server.
- [x] 4.3 Add only the documented WASM imports in `internal/plugin/bindings_wasm.go` — verify an existing WASM fixture still links with no new required imports.
- [x] 4.4 Wire real `ToolSet` into `Host` instead of `noopTools` — verify compile-time test keeps store/bus out of runtime packages and the demo plugin still loads.

## 5. Gates

- [x] 5.1 Run the full gate set — verify `make test lint`, `go vet ./...`, `gofmt -l .` empty, `openspec validate add-host-tools --strict` and `CGO_ENABLED=0 go build ./...` all exit 0.

## Workflow follow-up

- Merge `change/add-host-tools` into `main` with `--no-ff`.
- Archive the change with `openspec archive add-host-tools -y` after the gates pass.
