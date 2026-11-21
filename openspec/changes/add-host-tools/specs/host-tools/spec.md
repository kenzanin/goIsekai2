# Spec Delta

## ADDED Requirements

### Requirement: Host tool injection

The daemon SHALL construct exactly one host tool set and inject it into every plugin runtime. No
runtime SHALL receive the store handle, the event bus, an ambient filesystem root or an ambient HTTP
client that bypasses the tool set. A tool exposed in Lua and JS SHALL resolve to the same Go
implementation with the same JSON shape. WASM SHALL keep its existing import set; promoted helpers
that the matrix marks absent SHALL remain inert rather than becoming required imports.

#### Scenario: Same tool set in every runtime
- **WHEN** Lua and JS plugins call `now`, `log.info`, `storage.set` and `http.fetch`
- **THEN** both resolve the same Go implementation and produce identical return shapes

#### Scenario: No ambient authority
- **WHEN** a plugin tries to reach the SQLite file, the event bus or a path outside its own folder
- **THEN** no such capability is present in its tool environment and the attempt fails inside the plugin

#### Scenario: Existing WASM plugin still links
- **WHEN** a promoted helper is added to Lua and JS
- **THEN** an existing WASM module links and runs with its previous imports unchanged
