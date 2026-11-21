---
type: Fact
title: "tools regex: Capture falls back to whole match, Replace expands $1, no compile cache"
description: "internal/plugin/tools regex helpers pass the CGO gate with Capture=group-or-whole-match, $1 expansion in Replace, and stdlib limit passthrough in FindAll, proven by table tests."
tags: [internal/plugin/tools, plugin-host, regex]
generated: { by: agent/mcp, at: "2026-11-21T18:53:34Z" }
status: stable
---

## Evidence

```
CGO_ENABLED=0 go test ./internal/plugin/tools   # ok  goisekai2/internal/plugin/tools 0.028s
go vet ./internal/plugin/tools                  # clean
go run scripts/godoc-check.go                   # clean
```

Table tests in `internal/plugin/tools/regex_test.go` cover every entry point incl. invalid-pattern error rows (`"("`).

## Semantics (verified by those tests)

- `Match`/`Find`/`Capture`/`FindAll`/`Replace`/`FindIndex` compile per call over stdlib `regexp` (RE2); a non-compiling pattern is a returned error, distinguishable from a plain no-match (`""`/`nil`/`false`).
- `Capture` returns the first capture group, or the whole match when the pattern has none — the Lua `string.match` shape carried over from goIsekai V1 `internal/pluginutil/regex.go` (`RegexFind`).
- `Replace` expansion is Go syntax (`$1`, `${name}`), same as V1 `RegexReplace`; not literal, not Lua `%1`.
- `FindAll` limit is passed straight to `Regexp.FindAllString` (`<0` all, `0` none, `>0` first n).
- `FindIndex` is the half-open byte pair from `FindStringIndex`; `Quote` is `QuoteMeta`.

## Consequence

Plugin-facing regex helpers live beside the other host tools in `internal/plugin/tools`; `regex.go` carries a `ponytail:` note: add a compiled-pattern cache keyed by pattern only when scrape profiles show recompile cost. V1's coregex dependency is not needed — stdlib RE2 is already linear-time.

# Related Concepts
- [tools.ToolSet: storage delegates to plugin_state, missing keys return gorm not-found](plugin-tools-surface.md): regex helpers extend the plugin host tool surface; same package and gate
