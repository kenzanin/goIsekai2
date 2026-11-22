---
type: Decision
title: Explicit saved state on the saved_manga master row
description: "added_at is the saved marker: >0 means in library, 0 means materialized but unsaved. Viewing a title never saves it."
tags: [library, detail, store]
generated: { by: agent/mcp, at: "2026-11-22T10:37:10Z" }
status: stable
---

## Ruling
Opening a manga from search MUST NOT add it to the library. `saved_manga` remains the master row for all viewed titles, but an explicit save state is required.

- `added_at > 0` means explicitly saved; `added_at = 0` means a materialized master row that is not in the library.
- `library.Save` stamps `added_at` with the current Unix time; `library.Remove` clears it to 0 instead of deleting the master row.
- Library listing, FTS search and scheduler candidates filter on `added_at > 0`.
- The detail API exposes `saved` and the detail page renders Save or Unsave from it.

# Related Concepts
- [saved_manga is the single master for manga data](saved-manga-master-data.md): saved marker on the master row
- [API-first: JSON surface is the contract, SSR is a thin client](../conventions/api-first.md): detail API exposes the saved flag
