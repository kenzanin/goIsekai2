---
type: Fact
title: "tools fs: Read() sandboxes plugin file reads inside the plugin dir"
description: "internal/plugin/tools/fs.go Read() confines plugin reads to the plugin dir: absolute paths and cleaned ..-escapes are rejected, verified by green tests."
tags: [internal/plugin/tools, fs.go, plugin-host]
generated: { by: agent/mcp, at: "2026-11-21T19:06:42Z" }
status: stable
---

## Verified behaviour

`tools.Read(dir, name)` in `internal/plugin/tools/fs.go` resolves `name` strictly inside the plugin dir:

- absolute paths rejected: `plugin file "/etc/passwd": path must be relative`
- `..`-escaping cleaned paths rejected: `plugin file "../escape.txt": path escapes the plugin directory`
- embedded `..` that cleans away (`sub/../meta.txt`) is allowed and resolves inside `dir`

## Evidence

```
$ CGO_ENABLED=0 go test ./internal/plugin/tools
ok  	goisekai2/internal/plugin/tools	0.029s
```

Test `TestRead` (fs_test.go) covers normal read, `/etc/passwd` and `../escape.txt` with message assertions; `go vet`, `gofmt` and `scripts/godoc-check.go` clean, 2026-11-22.

## Consequence

Plugin runtimes get exactly one read-only filesystem gate; anything outside the plugin dir must go through another host tool, never a raw path.

# Related Concepts
- [tools.ToolSet: storage delegates to plugin_state, missing keys return gorm not-found](plugin-tools-surface.md): fs.Read is one host tool on the ToolSet surface exposed to plugin runtimes.
