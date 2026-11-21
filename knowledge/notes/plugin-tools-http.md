---
type: Fact
title: "tools HTTP surface: non-2xx is data, FetchBody errors on non-200, bare map means headers"
description: "internal/plugin/tools HTTP tool passes the CGO_ENABLED=0 gate with non-2xx-as-data semantics, FetchBody upstream-status errors and bare-header-map options, proven by httptest tests."
tags: [internal/plugin/tools, plugin-host, http]
generated: { by: agent/mcp, at: "2026-11-21T18:38:36Z" }
status: stable
---

## Evidence

`CGO_ENABLED=0 go test ./internal/plugin/tools` → `ok goisekai2/internal/plugin/tools 0.028s` (also `go vet` clean, `go run scripts/godoc-check.go` clean), 2026-11-22.

`internal/plugin/tools/http.go` adds `(*ToolSet).Fetch/FetchBody/GetJSON/PostJSON` on top of bogdanfinn/tls-client v1.18.1:

- Result envelope keys: `status` int, `body` string, `headers` map[string]string with lowercased keys (cppIsekai `fetch::Result` parity).
- Any HTTP status incl. 4xx/5xx is returned as data with nil error; only transport/build failures error. Proven by httptest 410 case.
- `FetchBody` errors `upstream status <code>` for every non-200 (asserted for 410).
- Options map: `method`, `body`, `headers`, `timeout_ms`, `profile`; a map containing none of these keys is treated as a bare header map (cpp get_json/post_json parity), so `GetJSON(url, map[string]any{"Accept": ...})` works — asserted.

## Consequence

Host adapters (`plugin.HostTools.Fetch` still `(ctx,url,headers)→[]byte`) can wrap this surface; plugin runtimes get one options shape matching the cpp bindings.

# Related Concepts
- [tools.ToolSet: storage delegates to plugin_state, missing keys return gorm not-found](plugin-tools-surface.md): HTTP tools extend the same ToolSet surface as storage/log; same package, same gates.
