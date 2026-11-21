# Proposal: add-plugin-host

## Why

`store` is archived, so the daemon can now persist plugin rows. Every later capability (library,
updates, reader, sites) needs the frozen plugin ABI and a safe runtime boundary before it can call a
site plugin. This change implements discovery, the three runtimes, the persisted lifecycle and the
management surface of the `plugin-host` capability.

## What Changes

- Discover one-folder-per-plugin manifests under the configured plugin directory with deterministic
  duplicate handling; skip broken plugins without aborting boot.
- Add one runtime interface (`Call(ctx, envelope) -> response`) with Lua, JS and WASM implementations
  using `gopher-lua`, `modernc.org/quickjs` and `wazero` under the no-cgo gate.
- Enforce the frozen JSON envelope, method vocabulary (`search`, `detail`, `chapters`, `pages`,
  optional `genres`) and per-runtime budgets/sandbox.
- Wrap each plugin instance in the qmuntal/stateless FSM (`Disabled`, `Enabled{Idle,Busy}`, `Failed`)
  with persist→commit→publish ordering and single-call-per-plugin serialization.
- Add boot reconciliation, reload routes, disabled-plugin inertness, health-as-not-state, and the
  templ-rendered `/plugins` management page (reusing the UI stack).
- Purge rows of plugin folders that disappear, but never prune when the plugins directory is missing.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `plugin-host` — adds the daemon-level boot ordering requirement (migrations first, discovery before
  listen) on top of the existing discovery/runtime/lifecycle requirements.

## Impact

- Code: new `internal/plugin` packages (discovery, manifest, runtime, lifecycle, host tools) and the
  `/plugins` HTTP page.
- Dependencies: `yuin/gopher-lua`, `modernc.org/quickjs`, `tetratelabs/wazero`, `qmuntal/stateless`;
  GORM store from `add-store`.
- No cgo: `CGO_ENABLED=0 go build ./...` remains the hard gate.
- Out of scope: host tools network/storage details (separate `host-tools` capability), plugin
  management commands beyond the spec page, site-specific plugins, library/reader business rules.

## Baseline

cppIsekai `src/plugin*` and the V1 plugin corpus define the ABI and capability vocabulary; V2 adds an
explicit FSM and typed lifecycle event while keeping the visible JSON fields byte-compatible.
