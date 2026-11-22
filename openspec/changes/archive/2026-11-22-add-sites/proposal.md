# Proposal: add-sites

## Why

`browser` is archived, so the runtime can drive JavaScript-only sites. Sites is the last capability:
it fixes the shipped plugin corpus, the mangadex reference plugin, and parity with the baseline.

## What Changes

- Validate the plugin folder contract, including manifest id matching the folder name.
- Ship the mangadex reference plugin with canonical search/detail/chapters/pages/genres behavior and optional auth.
- Serve plugin logos at `/plugin/<id>/logo` with correct content type.
- Port the baseline corpus inventory with an explicit working/omission parity table.
- Add conformance fixtures for canonical envelopes and unix timestamps.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `sites` — adds daemon wiring and the shipped corpus on top of the frozen plugin contracts.

## Impact

- Code: manifest validation, logo route, plugin corpus under `plugins/`, parity document.
- Dependencies: none beyond the archived stack.
- No cgo: `CGO_ENABLED=0 go build ./...` stays the gate.
- Out of scope: runtime engines, host tools, UI redesign.

## Baseline

cppIsekai `plugins_dev/` inventory and the `mangadex-plugin`, `plugin-system` specs.
