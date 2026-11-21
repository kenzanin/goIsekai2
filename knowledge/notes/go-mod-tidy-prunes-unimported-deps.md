---
type: Fact
title: go mod tidy prunes not-yet-imported deps; hand-edit for single promotions
description: "go mod tidy prunes declared-but-unimported stack pins (stateless, wazero, gopher-lua); promote a single dependency by hand-editing go.mod instead."
tags: [go.mod, dependencies, tooling]
generated: { by: agent/mcp, at: "2026-11-21T17:45:08Z" }
status: stable
---

generated: { by: "kit agent", at: "2026-11-22" }

Verified while promoting `modernc.org/quickjs` from indirect to direct for `internal/plugin/js.go`.

Evidence:
- `go mod tidy` on goIsekai2 removed three declared-but-not-yet-imported pins: `github.com/qmuntal/stateless v1.8.0`, `github.com/tetratelabs/wazero v1.12.0`, `github.com/yuin/gopher-lua v1.1.2` (plus 6 go.sum lines) — `git diff --stat` showed `go.mod | 5 +----`, `go.sum | 6 ------` — because no package under `internal/` or `cmd/` imports them yet (grep over `--include="*.go"` returns nothing).
- Hand edit instead: moved the single `modernc.org/quickjs v0.27.0` line from the indirect block into the direct `require` block; final diff is exactly `+ modernc.org/quickjs v0.27.0` / `- modernc.org/quickjs v0.27.0 // indirect`, gates stay green.

Consequence: on this repo run `go mod tidy` only when the whole tree's imports are final; to promote one dependency for a single package, hand-edit `go.mod` or the still-unimported stack pins (stateless, wazero, gopher-lua — recorded in decisions/dependency-stack) get pruned and must be re-added later.
