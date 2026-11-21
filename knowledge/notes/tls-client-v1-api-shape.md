---
type: Fact
title: "tls-client v1.18.1 API: functional options + fhttp types, DefaultClientProfile=Chrome_150"
description: tls-client v1.18.1 builds clients via functional options and requires bogdanfinn/fhttp request types; verified from module source plus passing CGO-free tests.
tags: [internal/plugin/tools, tls-client, dependencies]
generated: { by: agent/mcp, at: "2026-11-21T18:38:36Z" }
status: stable
---

## Evidence

Module source read at `$(go env GOMODCACHE)/github.com/bogdanfinn/tls-client@v1.18.1/client.go` and `client_options.go`, plus the new tests passing against httptest through the forked transport (`CGO_ENABLED=0 go test ./internal/plugin/tools` → ok, 2026-11-22).

- Constructor is `NewHttpClient(logger Logger, options ...HttpClientOption)` — functional options, NOT the older positional `HttpClientOptions` struct: we use `NewHttpClient(nil, WithTimeoutMilliseconds(ms), WithClientProfile(p))`; nil logger is legal (noop logger).
- Requests/responses are `github.com/bogdanfinn/fhttp` types (the fork), not `net/http` — `fhttp.NewRequest` is required for `client.Do`.
- `profiles.DefaultClientProfile = Chrome_150`; `WithTimeoutMilliseconds(int)` exists so no seconds rounding.
- Plain-HTTP (httptest) requests work through the forked transport; utls only engages for TLS.

## Consequence

`internal/plugin/tools/http.go` pins this exact call shape; any upgrade must re-check the option names and the fhttp import before compiling. Profile map exposes chrome_146/133/124/120, firefox_148/133/117 with `chrome_146` default (V1 `hostnet` default), unknown names warn and fall back.

# Related Concepts
- [Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, quickjs, wazero)](../decisions/dependency-stack.md): Pinpoints the exact NewHttpClient/option/fhttp call shape of the tls-client version recorded in the dependency decision.
