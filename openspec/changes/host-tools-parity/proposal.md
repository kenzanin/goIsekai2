# Proposal: host-tools-parity

## Why

The dispatcher and the runtime bindings drifted while porting: `host.text.normalize_status` and
`host.plugin.param` exist only as binding-local copies, Lua remaps `host.log.debug` onto
`log.info`, and Lua lacks wrappers for the browser, html, codecs and `re2.gmatch` names. cppIsekai
plugins assume the full baseline vocabulary, so every gap surfaces as a missing-name error inside a
real plugin. This change closes the recorded gap set without changing any existing name's behavior.

## What Changes

- MODIFIED `host-tools` — the one-implementation rule is tightened: every exposed host name has
  exactly one dispatcher implementation or a registered alias, and adding names never changes
  plugin-visible behavior in Lua, JS or WASM.
- ADDED baseline name parity: the enumerated cppIsekai host vocabulary resolves, with the recorded
  gaps (`log.debug`, `fs.read_lib`, six `text.*` helpers, `codecs.*`, `re2.gmatch`, crypto
  digest/hmac/vrf/aes names, Lua/JS `browser`/`html` wrappers) closed as canonical dispatcher
  entries or aliases proven equivalent by vector tests.
- ADDED confined `io` shim: read-only `io.read`/`io.lines` locked to the plugin directory — a
  documented divergence from the baseline's arbitrary-path `io.lines`.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `host-tools` — tightens single-implementation to cover aliases, adds baseline name parity and the
  confined io shim.

## Impact

- Code: `internal/plugin/tools` dispatcher entries plus an alias registry; Lua/JS binding wrappers;
  deletion of binding-local duplicate helpers.
- No new dependencies; `CGO_ENABLED=0 go build ./...` gate unchanged.
- Out of scope: tool semantics beyond parity, WASM import-set changes, plugin-side code.

## Baseline

cppIsekai `src/plugin/lua/bindings/*` host tables are the name authority; V2 matches names and
observable behavior, diverging only where recorded (confined io).
