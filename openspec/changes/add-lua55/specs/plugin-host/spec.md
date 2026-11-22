# Spec Delta

## MODIFIED Requirements

### Requirement: Runtime parity and explicit limits

The host SHALL expose one internal runtime interface (`Call(ctx, envelope) -> response`) implemented
by `lua`, `js` and `wasm`, so no runtime is reachable through a code path the others do not share. The
`lua` runtime SHALL be a pure-Go Lua 5.5 implementation (`modernc.org/lua`); a runtime that needs cgo,
or a module the runtime does not provide, SHALL fail with an explicit error naming the plugin rather than
misbehave silently. Lua plugins SHALL keep using the frozen JSON envelope and the same `host.*` tools as
the other runtimes, and a plugin-local `require("<module>")` SHALL resolve only `<module>.lua` inside the
plugin's own folder.

#### Scenario: JS plugin answers a call
- **WHEN** a registered `js` plugin defines `handle` and is called
- **THEN** the response shape matches any Lua plugin's response and the HTTP layer cannot tell the runtimes apart

#### Scenario: WASM plugin answers a call
- **WHEN** a registered `wasm` module exports `handle`
- **THEN** the envelope JSON is passed in and the module's response is returned unchanged in shape

#### Scenario: Lua 5.5 syntax executes
- **WHEN** a registered Lua plugin uses Lua 5.5 syntax such as `<const>` locals in `handle`
- **THEN** the plugin compiles, executes and returns the same envelope shape as a Lua 5.1-era plugin

#### Scenario: LuaJIT-only module
- **WHEN** a Lua plugin calls `require("ffi")`, `require("bit")` or `require("../outside")`
- **THEN** the call fails with an explicit unsupported-module or module-not-found error naming the plugin, and the host process is unaffected

#### Scenario: Broken runtime entry
- **WHEN** a `js` or `wasm` entry fails to parse, compile or instantiate
- **THEN** the plugin is not registered, the failure is logged, and already-registered plugins keep serving
