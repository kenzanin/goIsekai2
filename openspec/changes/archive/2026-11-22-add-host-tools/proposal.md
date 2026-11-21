# Proposal: add-host-tools

## Why

`plugin-host` is archived and runtimes now share one envelope, but their `HostTools` interface is
still a noop. Every site plugin depends on the frozen `host.*` vocabulary; without it, real plugins
cannot fetch pages, read their own files, normalize text or sign requests. This change implements the
single host-side tool set and binds it into all three runtimes.

## What Changes

- Implement one `ToolSet` in Go with the frozen tool names, argument order and JSON return shapes;
  runtime bindings are thin marshals over it.
- Add HTTP helpers (`http.fetch`, `fetch_body`, `get_json`, `post_json`) over the pure-Go
  `bogdanfinn/tls-client`, with per-call profile override, timeout and optional per-plugin cookie jar.
- Add storage/state, confined plugin-folder reads, config lookup, logging/time, text/url/regex/html
  helpers, crypto primitives and request signing.
- Bind the tools into Lua, JS and WASM without giving any runtime the store, event bus or ambient
  filesystem/network authority.
- Keep WASM gaps inert: no new required imports for promoted helpers.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `host-tools` — adds the daemon-level wiring requirement (one tool set injected into every runtime,
  no ambient authority) on top of the frozen tool contracts.

## Impact

- Code: new `internal/plugin/tools` package plus binding files for each runtime.
- Dependencies: `bogdanfinn/tls-client` v1.18.1; `golang.org/x/net/html` for selectors; stdlib for
  regex/url/crypto/encoding.
- No cgo: `CGO_ENABLED=0 go build ./...` remains the hard gate.
- Out of scope: plugin-side scrapers, site plugins, library/reader features, TLS profile table data
  (plugin-side).

## Baseline

cppIsekai `Host::call` and V1's Lua/JS plugins define the vocabulary; V2 keeps names and shapes
identical while adding one typed Go implementation behind all bindings.
