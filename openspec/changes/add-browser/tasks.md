# Tasks: add-browser

## 1. Client and supervisor (parallel lanes 1A/1B)

- [ ] 1A.1 Implement `internal/browser/doc.go + browser.go + client.go`: `Browser` interface, MCP JSON-RPC client, per-session serialization, one expiry retry, typed fail-fast errors — verify with an httptest backend.
- [ ] 1A.2 Add client tests for successful evaluate, unknown tool, repeated expiry and loopback proxy bypass — verify exact error shapes.
- [ ] 1B.1 Implement `internal/browser/supervisor.go`: health probe, reuse foreign instance, spawn lightpanda child, wait for health, one respawn, then dead, graceful shutdown of owned child — verify with a fake binary script.
- [ ] 1B.2 Add supervisor tests for reuse, spawn and single-respawn behavior — verify no foreign process is killed.

## 2. Host wiring (parallel lanes 2A/2B)

- [ ] 2A.1 Add `browser.call` to tool dispatch with capability gating and unknown-tool failure — verify an undeclared plugin never touches the backend.
- [ ] 2A.2 Add `image_referer: "none"` to the manifest and apply it to image cache fetches — verify no Referer for the sentinel and unchanged behavior otherwise.
- [ ] 2B.1 Wire backend construction, start before listen and shutdown stop into `cmd/goisekai2` — verify `CGO_ENABLED=0 go build ./...` and a clean start/stop integration run.
- [ ] 2B.2 Add HTTP/tool tests for the passthrough contract — verify ok and failure envelopes.

## 3. Gates

- [ ] 3.1 Run the full gate set — verify `make test lint`, `go vet ./...`, `gofmt -l .` empty, `openspec validate add-browser --strict` and `CGO_ENABLED=0 go build ./...` all exit 0.

## Workflow follow-up

- Merge `change/add-browser` into `main` with `--no-ff`.
- Archive the change with `openspec archive add-browser -y` after the gates pass.
