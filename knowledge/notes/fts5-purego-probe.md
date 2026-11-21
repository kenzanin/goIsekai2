---
type: Fact
title: "FTS5 works on the pure-Go GORM stack, but V2 specs omit it"
description: "modernc.org/sqlite v1.62.0 exposes SQLite 3.54.1 with FTS5 and bm25(); MATCH and ranking work under CGO_ENABLED=0, but the goIsekai2 specs still lack an FTS5 requirement."
tags: [store, fts5, gorm, parity]
generated: { by: agent/cli, at: "2026-11-21T14:34:26Z" }
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
`bm25()` work; small magnitudes are normal for a three-row corpus, not a driver
defect.

## Consequence for the store

GORM has no typed FTS5 API: virtual-table DDL, trigger maintenance and
`MATCH ... ORDER BY bm25(...)` are raw SQL, so they belong inside the
`internal/store/sql` module and are covered by the boundary gate. Typed
repository methods expose the results to callers.

## Parity gap

The cppIsekai baseline has `knowledge/architecture/fts5-search-dedup.md`:
FTS5 + `bm25()` ranking for title search and dedup candidates, replacing LIKE.
The goIsekai2 capability specs (`openspec/specs/store/spec.md`,
`openspec/specs/library/spec.md`) contain no FTS5 requirement, virtual table or
MATCH scenario. This is a missing requirement, not a driver limitation — it
must be added (as a delta or a spec MODIFIED entry) before `add-store` locks the
frozen schema, so the index tables are part of the migration set from day one.

# Related Concepts
- [Store access: GORM with one raw-SQL boundary](../decisions/store-access.md): proves FTS5 and BM25 work on the chosen pure-Go GORM stack
