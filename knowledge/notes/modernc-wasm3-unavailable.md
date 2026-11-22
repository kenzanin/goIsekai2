---
type: Fact
title: No published modernc.org/wasm3 module; keep wazero
description: "modernc.org/wasm3 is not published on the Go module proxy (404) and its gitlab.com/cznic/wasm3 source cannot be fetched anonymously, so the pure-Go wasm plugin runtime stays on tetratelabs/wazero v1.12.0."
tags: [wasm, wasm3, wazero, dependencies, no-cgo]
generated: { by: agent/mcp, at: "2026-11-22T14:43:52Z" }
status: stable
---

## Evidence

- `go list -m -versions modernc.org/wasm3` failed: `git ls-remote https://gitlab.com/cznic/wasm3` prompted for credentials (anonymous fetch unavailable).
- `https://proxy.golang.org/modernc.org/wasm3/@v/list` -> 404.
- `https://pkg.go.dev/modernc.org/wasm3` -> 404.
- `https://gitlab.com/cznic/wasm3` -> 429 from this environment.

## Consequence

There is no usable published pure-Go wasm3 module to swap in. Native wasm3 is a C interpreter needing cgo, which violates the project's `CGO_ENABLED=0` gate. Keep tetratelabs/wazero v1.12.0 for WASM plugin modules.

If the user supplies an exact reachable module path or repository, re-evaluate before changing this note.

# Related Concepts
- [Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, quickjs, wazero)](../decisions/dependency-stack.md): Re-evaluates the earlier wasm3 rejection: no published modernc pure-Go module exists, so wazero stays.
