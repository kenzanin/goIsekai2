# Design: add-host-tools

## Context

See `proposal.md`. The `host-tools` capability spec freezes the tool vocabulary and parity matrix;
`plugin-host` already defines the `HostTools` interface. cppIsekai `Host::call` is the baseline for
argument order and return shapes. V2 implements one Go tool set behind thin runtime bindings.

## Goals / Non-Goals

- Goals: one implementation per tool, identical JSON shapes, no ambient authority, no cgo, per-plugin
  scoping for storage and file reads, deterministic failure surfaces.
- Non-Goals: plugin-side profile/regex tables, new site plugins, library/reader features.

## Decisions

### D1 — One Go tool set

`internal/plugin/tools.ToolSet` owns every tool method. Lua/JS/WASM bindings marshal arguments into
it and marshal results back. Baseline: cppIsekai `Host::call` one host implementation.

Alternative: per-runtime helpers — rejected because parity defects become invisible.

### D2 — HTTP client

`bogdanfinn/tls-client` v1.18.1 provides browser TLS fingerprints without cgo. Per-call `profile`
and `timeout_ms` override config defaults for that call only; unknown profiles warn and fall back.
Non-2xx is data, transport failures are typed errors with URL and plugin id in the log.
Baseline: cppIsekai TLS profile table and V1's tls-client usage.

Alternative: stdlib `net/http` — rejected because sites fingerprint TLS.

### D3 — Cookie jar scoping

An opt-in per-plugin jar lives at `data/plugin-state/<id>.cookies.txt`, is loaded at construction and
saved after each fetch. No default jar, no cross-plugin sharing. Baseline: cppIsekai per-plugin jars.

### D4 — Storage and config

`storage.get/set` use the existing `plugin_state` table through the typed store, keyed by the calling
plugin's folder id. `config.get` reads the immutable config snapshot; secrets are never returned.
Baseline: cppIsekai host storage/config helpers.

### D5 — Confined file reads

`fs.read` joins only inside the plugin's own folder and rejects `..`, absolute paths and paths that
escape after `filepath.Clean`. There is no write tool. Baseline: cppIsekai plugin-folder reads.

### D6 — Text/url/regex/html

Go `regexp` (RE2), `net/url`, `golang.org/x/net/html` and small normalization helpers implement the
frozen `re2.*`, `url_*`, `html.*` and `text.*` names. Promoted helpers absent in WASM stay absent; no
new WASM imports are added. Baseline: V1 plugin usage of these helpers.

### D7 — Crypto and signing

Stdlib `crypto/*` implements hash/hmac/base64/hex and the documented signing helper. No plugin gets
raw key material beyond the params it passes. Baseline: cppIsekai crypto host tools.

### D8 — Bindings

Lua exposes tools as global `host.*` tables with JSON-string marshalling; JS exposes `host` as a
frozen object; WASM receives only the documented flat `env.<group>_<fn>` imports. All three call the
same ToolSet methods. Baseline: V1 Lua/JS plugin call sites.

## Risks / Trade-offs

- [tls-client API drift] → pin v1.18.1 and wrap it behind `http.fetch`; no plugin imports it directly.
- [regex incompatibility] → Go `regexp` is RE2; unsupported syntax fails with a named error.
- [WASM parity gaps] → keep promoted helpers out of the WASM import set; existing modules still link.

## Migration Plan

Greenfield: no real plugins ship yet; the demo Lua plugin exercises the tools. Rollback is reverting
the change commit.

## Open Questions

None. Tool names, argument order and parity are fixed by the spec.
