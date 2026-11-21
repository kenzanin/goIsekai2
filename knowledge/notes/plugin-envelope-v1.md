---
type: Fact
title: Plugin host contract v1 implemented and gated in internal/plugin
description: "The plugin envelope ABI (HostAPI=1, five-method allowlist, Runtime.Call) exists at internal/plugin with green CGO-free test, vet and godoc gates as of 2026-11-22."
tags: [internal/plugin, runtime.go, js.go]
generated: { by: agent/mcp, at: "2026-11-21T17:45:19Z" }
status: stable
---

generated: { by: "kit agent", at: "2026-11-22" }

Verified while implementing `internal/plugin` (no commit, per brief).

Evidence:
- `go get` resolved all four pins: github.com/yuin/gopher-lua v1.1.2, modernc.org/quickjs v0.27.0 (pulls modernc.org/libquickjs v0.14.0), github.com/tetratelabs/wazero v1.12.0, github.com/qmuntal/stateless v1.8.0.
- `CGO_ENABLED=0 go test ./internal/plugin` → `ok goisekai2/internal/plugin 0.003s`.
- `go vet ./internal/plugin/...` clean; `go fmt` reports no changes.
- `go run scripts/godoc-check.go`: zero findings under `internal/plugin`; its only 2 findings are pre-existing in gitignored scratch `tmp/probe.go` (outside this task's allowed files).

Consequence: the plugin ABI seam is now real and gated — `Envelope{API,Method,Params}` with `HostAPI = 1`, the five-method allowlist (`search`, `detail`, `chapters`, `pages`, `genres`) enforced by `ValidateMethod`, and every future engine (Lua/JS/wasm) lands as a `Runtime` behind `Call(ctx, []byte) ([]byte, error)`. Repo-wide godoc gate is red only from `tmp/probe.go`, unrelated to this package.

Update 2026-11-22: the first engine exists — `internal/plugin/js.go` implements `Runtime` on modernc.org/quickjs (per-call VM, ctx-deadline via `SetEvalTimeout`, `handle(<json>)` dispatch, Promise results rejected, no fs/process/network globals); `CGO_ENABLED=0 go test ./internal/plugin`, `go vet` and `go run scripts/godoc-check.go` all green with quickjs promoted to a direct dependency.

# Related Concepts
- [Architecture overview: layers, import rules and frozen external contracts](../architecture/overview.md): Implements the frozen plugin ABI baseline recorded in the architecture overview (envelope contract, method vocabulary).
