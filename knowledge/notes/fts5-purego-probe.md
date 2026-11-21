---
type: Fact
title: "FTS5 works on the pure-Go GORM stack, but V2 specs omit it"
description: "modernc.org/sqlite v1.62.0 exposes SQLite 3.54.1 with FTS5 and bm25(); MATCH and ranking work under CGO_ENABLED=0, and the store/library specs now require FTS5 title search and BM25-ranked dedup."
tags: [store, fts5, gorm, parity]
generated: { by: agent/cli, at: "2026-11-21T14:42:21Z" }
status: stable
---

## Evidence

Throwaway probes with `gorm.Open(glebarez.Open(...), &gorm.Config{})` under
`CGO_ENABLED=0` and `modernc.org/sqlite v1.62.0`:

- `CREATE VIRTUAL TABLE docs USING fts5(title, body)` — succeeded.
- `INSERT` into the virtual table — succeeded.
- `SELECT count(*) FROM docs WHERE docs MATCH ?` with `"full"` — returned `1`.
- `SELECT title, bm25(docs), bm25(docs, 10.0, 1.0) FROM docs WHERE docs MATCH ?
  ORDER BY bm25(docs)` — returned ranked rows with negative scores, e.g.
  `manga = -1.7701149425287355e-06`, `novel = -1.0620689655172414e-06`.
- `SELECT sqlite_version()` — returned `3.54.1`.

The same query against the system `sqlite3` CLI produced **identical** scores,
so the pure-Go engine is not altering BM25 semantics. Column weights and
`bm25()` work; small magnitudes are normal for a three-row corpus.

## Baseline contract ported into V2 specs

The baseline creates (in `src/store/store_core.cpp`):
`CREATE VIRTUAL TABLE IF NOT EXISTS library_fts USING fts5(plugin UNINDEXED,
slug UNINDEXED, name, alt_titles, synopsis)`. `fts_sync` replaces rows by
`(plugin, slug)`; `fts_delete` removes them. Dedup uses
`MATCH '{name alt_titles} : (...)' ORDER BY bm25(library_fts) LIMIT 20`,
token-multiset Dice similarity at the baseline thresholds, and union-find.

V2 specs now carry this as requirements:

- `openspec/specs/store/spec.md`: `library_fts` in the frozen schema, an FTS5
  title-index requirement (create, refresh, ranked MATCH, prefixed terms,
  LIKE fallback), and BM25 candidates in the dedup requirement.
- `openspec/specs/library/spec.md`: local library title search using FTS5
  `MATCH` + `bm25`, distinct from plugin-backed discovery search, with prefix
  matching and an unranked LIKE fallback.

All FTS5 DDL and MATCH statements are raw SQL and therefore live inside the
`internal/store/sql` module under the boundary gate.

# Related Concepts
- [Store access: GORM with one raw-SQL boundary](../decisions/store-access.md): proves FTS5 and BM25 work on the chosen pure-Go GORM stack
- [Concurrency and ownership: one owner per mutable resource, bounded everything](../architecture/concurrency-ownership.md): FTS5 adds no worker-pool requirement; the policy is recorded separately
