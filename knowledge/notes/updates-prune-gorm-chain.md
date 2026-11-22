---
type: Fact
title: "Cache prune is a GORM chain, not a store method"
description: PruneCache deletes expired api_cache rows via Where/Delete on Store.DB() because no typed prune method exists.
tags: [updates, cache, prune, gorm, sql-boundary]
generated: { by: agent/cli, at: "2026-11-22T03:18:24Z" }
status: stable
---

internal/store has no typed prune method, so internal/updates.PruneCache deletes rows with st.DB().Where("fetched_at + ttl_s <= ?", now).Delete(&store.ApiCache{}). That is a GORM model chain, not raw SQL: scripts/sql-boundary-check.go still reports OK because only Raw/Exec/Query-style APIs and SQL statement literals are gated. Moving it into a typed store method is the cleaner end state once the store lane owns it.

# Related Concepts
- [Raw SQL boundary](../conventions/sql-boundary.md): The prune uses a gorm chain via store.DB(), which this convention permits outside internal/store/sql.
