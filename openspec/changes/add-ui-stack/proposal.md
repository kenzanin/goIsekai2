# Proposal: add-ui-stack

## Why

`add-config` shipped `/settings` with `html/template` and a hand-written fetch script. The user
directed a typed, local-first UI stack (templ, Tailwind, Alpine, htmx) for goIsekai V2, so the page
must be ported now and the same stack must be the foundation for later pages (library, plugins,
reader). Changing an archived change is not allowed, so this is a new change with a delta against
the `config` capability.

## What Changes

- Replace the embedded `html/template` pair with typed templ components; generated `*_templ.go`
  files are committed and embedded.
- Compile Tailwind CSS to one static stylesheet; vendor Alpine.js and htmx minified builds under
  the HTTP package and embed them — no CDN, no runtime Node build.
- Use Alpine for local component state (dirty-field tracking, status banners) and htmx for server
  round trips (settings save, fragment/status swap), keeping the existing JSON `/config` contract.
- Add `make generate` (templ + Tailwind) and a freshness gate; `make build` depends on generation.
- **BREAKING**: none in the HTTP API. `/settings` visual markup changes, but routes, JSON payloads,
  provenance rules, restart badges and the no-dead-links rule stay identical.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `config` — settings page behavior gains a local-first asset guarantee and no-reload status
  updates; the registry/API contract is unchanged.

## Impact

- Code: `internal/http` (templates → templ components, static assets, handler rendering).
- Tooling: Go module gains templ as a codegen tool; `Makefile` gains generate + freshness target;
  Tailwind CLI is a build-time tool only.
- Dependencies: no cgo; generated Go and static assets are committed; `CGO_ENABLED=0 go build ./...`
  remains the gate.
- Out of scope: new pages (library/plugins/reader), WebSocket log console, theme system, any SPA
  router, changing `GET/POST /config` payloads, store/schema work.
