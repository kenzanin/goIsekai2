---
type: Decision
title: saved_manga is the single master for manga data
description: "All manga metadata lives in saved_manga; satellite tables are keyed by (plugin, slug) and never own competing title data. Detail API materializes it."
tags: [store, detail, architecture]
generated: { by: agent/mcp, at: "2026-11-22T06:09:08Z" }
status: stable
---

## Ruling
`saved_manga` is the single master record for manga metadata. The manga detail page is its primary presentation; every other manga-scoped table links back by the composite key `(plugin, slug)`.

- All title-facing fields (name, cover, status, synopsis, alt titles, authors, artists, tags, flags, counts) live in `saved_manga`.
- `chapters`, `chapter_pages`, `reader_positions`, `update_attempts`, `title_meta` and API cache rows are satellite data keyed by `(plugin, slug)`; they never own competing title metadata.
- Any plugin detail or chapters response MUST materialize or refresh the `saved_manga` row before serving, including a title the user has not explicitly saved.
- Enrichment candidates stay in `title_meta` until the user applies one; applying writes the chosen fields into `saved_manga`.
- Search/FTS, grid filters and the detail page read from `saved_manga`; satellite tables are joined only for chapter, position and attempt state.
- Detail API remains the contract; SSR is a projection of it (see conventions/api-first).

# Related Concepts
- [Architecture overview: layers, import rules and frozen external contracts](../architecture/overview.md): Master-data rule for the store layer
- [API-first: JSON surface is the contract, SSR is a thin client](../conventions/api-first.md): Detail API projects the saved_manga master row
- [Data flows: five ordered sequences and the invariants that keep them reproducible](../architecture/data-flows.md): Materialization order for detail and chapter flows
