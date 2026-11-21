---
type: Fact
title: "tools url_encode: QueryEscape semantics, + for space -- recorded divergence from cppIsekai %20"
description: "Dispatch url_encode is net/url QueryEscape (+ for space, lowercase hex), diverging from cppIsekai's %20/uppercase variant; pinned by parity_test.go."
tags: [internal/plugin/tools, parity_test.go, plugin-host]
generated: { by: agent/mcp, at: "2026-11-21T19:33:06Z" }
status: stable
---

Verified by `CGO_ENABLED=0 go test ./internal/plugin/tools` -> `ok goisekai2/internal/plugin/tools 0.080s` (exit 0) and `go run scripts/godoc-check.go` (exit 0), via `internal/plugin/tools/parity_test.go`:

- `Dispatch("url_encode", "a b&c=d")` == `a+b%26c%3Dd`: space becomes `+`, hex is lowercase (`net/url.QueryEscape`).
- cppIsekai's `url_encode` (`~/tmp/cpp/cppIsekai/src/http/pages_api.cpp:17`) instead emits `%20` for space and UPPERCASE hex, passing only unreserved chars through.

Consequence: this is a recorded parity divergence, not a bug -- Go `url_encode` is URL-query-component semantics (matches cppIsekai's call sites, which always encode query values), while raw-path encoding would need `url.PathEscape`. Any plugin output diffed against cppIsekai must normalize `+` vs `%20` before comparing.

# Related Concepts
- [tools text helpers](plugin-tools-text.md): sibling text-tool semantics locked by tests.
- [tools text helpers: Clean keeps newline/tab as whitespace, DateEpoch dual-layout, JSONBlob sorted-key stable](plugin-tools-text.md): url_encode divergence sits next to the locked text-helper semantics in the same dispatch surface.
