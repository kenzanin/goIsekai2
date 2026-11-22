# Design: add-behavior-specs

## D1. Specs follow current behavior, not old intent

The archived capability specs predate several deliberate fixes. Where the main spec still describes the old
behavior, this change replaces that requirement instead of appending a competing note. Examples:
`POST /api/library/remove` clears `added_at` and keeps the master row; the detail toggle uses the `saved`
field returned by `/api/detail`.

## D2. Group related unspecced behavior by capability

- `library`: membership semantics and the plugin-backed search API it feeds.
- `detail`: cold materialization, envelope canonicalization and image proxying.
- `observability`: About/Log pages, the WebSocket log stream and debug-only trace logging.
- `ui-shell`: the canonical nav order inherited from the baseline shell.

## D3. Two V2-only capabilities

cppIsekai has no standalone observability or shell specs, but V2 added `/about`, `/log`, `/ws/log` and a pinned
nav order. Recording them as capabilities makes future models treat those routes and ordering as contract,
not incidental template content. `architecture/capability-map.md` and `knowledge/process/porting-order.md`
are updated after archive.

## D4. No runtime edit

This change must produce a clean `make test lint` with zero source diff outside `openspec/` and the two
knowledge index files. If a spec statement cannot be traced to current code, fix the spec wording here,
not the runtime.
