---
type: Fact
title: "tools text helpers: Clean keeps newline/tab as whitespace, DateEpoch dual-layout, JSONBlob sorted-key stable"
description: "internal/plugin/tools text helpers pass the CGO gate with Clean treating \\n\\t as whitespace but stripping other control runes, DateEpoch limited to RFC3339 and 2006-01-02, and JSONBlob stable via encoding/json key sorting."
tags: [internal/plugin/tools, plugin-host, text]
generated: { by: agent/mcp, at: "2026-11-21T18:47:53Z" }
status: stable
---

Verified by `CGO_ENABLED=0 go test ./internal/plugin/tools` → `ok goisekai2/internal/plugin/tools 0.028s` plus `go run scripts/godoc-check.go` (clean) on internal/plugin/tools/text.go.

Text helper semantics locked in by text_test.go:

- `Clean` strips non-space control runes (e.g. `\x00\x01`) but lets `\n`/`\t` through the control filter and collapses them as whitespace: `Clean("  a \t\t b \n c  ") == "a b c"`, `Clean("a\x00\x01b") == "ab"`.
- `DateEpoch` accepts exactly `time.RFC3339` then `2006-01-02`; `2026-11-22T12:34:56+07:00` → `1795325696`, plain date parses as UTC midnight `1795305600`; anything else is a `*time.ParseError`.
- `JSONBlob` is stable via `encoding/json` (map keys sorted, nested maps too): `map[b:2 a:1]` → `{"a":1,"b":2}`; unmarshalable input (chan) → `""`.

Consequence: plugin authors get deterministic serialization and one lenient date parser; callers of `JSONBlob` must treat `""` as marshal-failure, not empty JSON — but `json.Marshal` of `nil` correctly yields `"null"`, checked by test.

# Related Concepts
- [tools.ToolSet: storage delegates to plugin_state, missing keys return gorm not-found](plugin-tools-surface.md): text helpers extend the host tool surface next to the ToolSet core.
