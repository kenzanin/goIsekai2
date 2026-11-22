# Design: add-sites

## Context

See `proposal.md`. The `sites` capability spec freezes the folder contract, ABI and corpus parity.
cppIsekai `plugins_dev/` is the read-only corpus, not a dependency.

## Goals / Non-Goals

- Goals: folder validation, reference plugin, logo route, corpus inventory, conformance fixtures.
- Non-Goals: new runtime engines, host tools, UI redesign.

## Decisions

### D1 — Folder id is authoritative

Disk folder name must equal `main.json` `id`; a mismatch is skipped and logged. Loader already keys
instances by folder, so this prevents two identities for one plugin.

### D2 — Mangadex reference

Port the baseline mangadex Lua plugin unchanged where the ABI matches. Canonical vocabulary stays
English; ranking, tag AND semantics, disjoint page windows and optional password-flow auth stay as
in the baseline.

### D3 — Logo route

Serve `GET /plugin/<id>/logo` from the plugin folder using `http.ServeContent`, with a fallback to
the existing plural `/plugins/<id>/logo` used by the management page. Missing logo returns 404.

### D4 — Parity inventory

Generate a table of every baseline plugin folder with status `ported` or `omitted(reason)`. Corpus
shrinkage without a recorded decision is a defect.

### D5 — Conformance fixtures

Embed small recorded response fixtures and assert envelope shape: `ok`, `data.items`, integer
`total`, integer `release_at`, absolute `pages` URLs.

## Risks / Trade-offs

- [ABI drift] → port baseline scripts verbatim; record deviations.
- [auth secrets] → config.json stays in the plugin folder, never logged.
- [corpus size] → inventory documents omissions rather than silently dropping plugins.

## Migration Plan

Greenfield. Rollback is reverting the change commit.

## Open Questions

None. Behavior is fixed by the capability spec.
