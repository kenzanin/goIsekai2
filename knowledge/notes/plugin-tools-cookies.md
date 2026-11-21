---
type: Fact
title: "tools cookies: per-plugin TSV jar; SaveJar needs a LoadJar jar (stdlib jar has no enumeration)"
description: "internal/plugin/tools persists per-plugin cookies as host/name/value TSV under plugin-state, and SaveJar only accepts jars from LoadJar because stdlib cookiejar cannot be enumerated."
tags: [internal/plugin/tools, plugin-host, cookies]
generated: { by: agent/mcp, at: "2026-11-21T18:45:41Z" }
status: stable
---

## Verified

`CGO_ENABLED=0 go test ./internal/plugin/tools` → `ok goisekai2/internal/plugin/tools`, plus `go vet`, `gofmt -l`, `go run scripts/godoc-check.go` all clean.

`internal/plugin/tools` cookie persistence (`cookies.go`):

- `ToolSet.JarPath()` = `filepath.Join(DataDir, "plugin-state", PluginID+".cookies.txt")` — per-plugin, no global jar.
- Format is 3-column TSV per line: `host TAB name TAB value`. No path, expiry, domain flag, or secure flag survives a round trip; every restored cookie is host-scoped on path `/`.
- `SaveJar` requires a jar obtained from `LoadJar` (`*cookieJar`), because `net/http/cookiejar.Jar` has no enumeration API (verified: `go doc net/http/cookiejar.Jar` lists only `Cookies`/`SetCookies`); the wrapper records hosts on `SetCookies` and at load.
- Missing state file = empty jar; unreadable file logs a warning and yields an empty jar (cppIsekai parity: proceed without jar).

## Consequence

Any host glue that hands `SaveJar` a foreign `cookiejar.Jar` gets an error, not a silent skip. Cookie values containing tab/newline would corrupt the TSV (ponytail-marked); escape only when a real plugin hits it.

# Related Concepts
- [tools.ToolSet: storage delegates to plugin_state, missing keys return gorm not-found](plugin-tools-surface.md): Cookie persistence is the third leg of the same ToolSet host surface, beside storage and HTTP fetch.
