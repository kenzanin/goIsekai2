# Design: add-detail

## Context

See `proposal.md`. The `detail` capability spec freezes the page, routes, stale behavior and
enrichment flow. cppIsekai `manga-detail`/`detail-enrichment` are the baseline; V2 uses dedicated JSON
routes as a recorded divergence.

## Goals / Non-Goals

- Goals: saved-row read path, validated cold materialization, atomic refresh, user-selected enrichment,
  offline cover/related, accessible controls.
- Non-Goals: site plugins, scheduler internals, reader changes.

## Decisions

### D1 — Detail service

`internal/detail.Service` owns envelope validation, stale-serve and persistence through typed store
methods. No raw SQL outside `internal/store/sql`.

### D2 — Saved-row first

Saved detail reads the row directly, returns `cached: true`, and never calls the plugin. A disabled
plugin still serves the row. Baseline: cppIsekai saved-row read.

### D3 — Refresh transaction

Call detail then chapters; validate both; write both in one store Tx; stamp `update_attempts` `ok` or
`error`. Failure leaves both entries intact. Baseline: cppIsekai refresh semantics.

### D4 — Enrichment registry

Separate loader over `data/enrichment/<id>/` using the same runtimes and host tools, exposing only
`enrich`. Unknown envelope keys are ignored. Baseline: cppIsekai enrichment registry.

### D5 — Merge policy

Plugin fields are the base. Enrichment writes only into a candidate set; stored `fetched` payloads are
used on normal fetch, force always refreshes. Applying a candidate is an explicit user action.

### D6 — Cover and related

Cover wizard selects from candidates and writes `chosen_cover`; reset restores the plugin cover.
Related reading is derived from stored tags/source metadata.

## Risks / Trade-offs

- [envelope drift] → strict validation before any write.
- [partial refresh] → all-or-nothing Tx.
- [merge surprise] → no automatic overwrite of plugin fields.

## Migration Plan

Greenfield. Rollback is reverting the change commit.

## Open Questions

None. Routes and behavior are fixed by the spec and the recorded divergence.
