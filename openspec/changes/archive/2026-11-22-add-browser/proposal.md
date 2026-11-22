# Proposal: add-browser

## Why

`updates` is archived, so the maintenance loop is complete. Browser passthrough is next in the
porting order and is the capability plugins need for JavaScript-only sites.

## What Changes

- Add a `Browser` interface with an MCP-backed client exposing `browser.call(tool, args)` to plugins.
- Serialize multi-step sequences per session, recover one session expiry, and fail fast on an unreachable backend.
- Supervise an owned `lightpanda mcp` child: reuse healthy instance, wait for health, one respawn, then dead.
- Bypass the host proxy for loopback backend URLs.
- Support `image_referer: "none"` in plugin manifests so image fetches omit `Referer`.
- Deny `browser.call` to plugins that do not declare browser capability.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `browser` — adds daemon wiring and host-tool registration on top of the frozen browser contracts.

## Impact

- Code: new `internal/browser` package, host-tool dispatch, manifest field, boot/shutdown wiring.
- Dependencies: stdlib `os/exec` + `net/http`; no cgo, no new module.
- Out of scope: CDP adapter, site plugins, UI pages.

## Baseline

cppIsekai `browser-host` spec and the recorded `decisions/browser-host-transport` divergence.
