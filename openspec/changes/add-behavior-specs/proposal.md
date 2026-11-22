# Proposal: add-behavior-specs

## Why

Several behavior changes landed directly in code after their capability specs were archived. They are correct
in the running daemon but invisible to a model that only reads `openspec/specs/`. This change writes those
behaviors down as contract deltas so future work cannot silently undo them.

## What Changes

- Document the explicit library-membership marker (`added_at`), metadata-preserving save, remove-keeps-row, and form-encoded htmx posts in the `library` spec.
- Document the plugin-backed `/api/search` endpoint and the search page's use of it.
- Document detail master-row materialization for unsaved titles, envelope field canonicalization, rematerialization, `saved` payload field, and `/cache/img` proxying in the `detail` spec.
- Add `observability` capability for `/about`, `/log`, `/ws/log`, and debug-only instrumentation.
- Add `ui-shell` capability for the canonical nav order inherited from cppIsekai.

## Capabilities

### New Capabilities

- `observability` — log stream, About/Log pages and gated debug instrumentation.
- `ui-shell` — global navigation order shared by every SSR page.

### Modified Capabilities

- `library` — save/remove semantics, toggle hydration and plugin search API.
- `detail` — master-row materialization, envelope canonicalization, image proxying and form save.

## Impact

- Documentation/spec only: no runtime code changes, no new dependencies, no route removal.
- Corrects stale statements: `/api/library/remove` no longer deletes a row, and toggle hydration uses the detail payload's `saved` flag rather than a separate membership fetch.
- The capability map and porting order gain the two new V2-only capabilities after archive.

## Baseline

Behavior already present in `internal/library`, `internal/detail`, `internal/http`, `internal/stream` and the SSR components as of commit `7a0b8ba`. Where V2 diverges from cppIsekai, the divergence is stated in the requirement body.
