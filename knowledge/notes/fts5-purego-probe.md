---
type: Fact
title: "FTS5 works on the pure-Go GORM stack, but V2 specs omit it"
description: "modernc.org/sqlite v1.62.0 exposes SQLite 3.54.1 with FTS5; CREATE VIRTUAL TABLE and MATCH succeed under CGO_ENABLED=0, while the goIsekai2 specs still lack an FTS5 requirement."
tags: [store, fts5, gorm, parity]
generated: { by: agent/cli, at: "2026-11-21T14:30:52Z" }
status: stable
---

## Evidence

Ran a throwaway probe with `gorm.Open(glebarez.Open("file:fts?mode=memory&cache=shared"), &gorm.Config{})`
under `CGO_ENABLED=0` and `modernc.org/sqlite v1.62.0`:

- `CREATE VIRTUAL TABLE docs USING fts5(title, body)` — succeeded.
- `INSERT INTO docs(title, body) VALUES('isekai', 'manga library full text')` — succeeded.
- `SELECT count(*) FROM docs WHERE docs MATCH ?` with `"full"` — returned `1`.
- `SELECT sqlite_version()` — returned `3.54.1`.

So FTS5 and BM25 do not require cgo on this stack; the pure-Go GORM driver is
not the limiting factor.

## Parity gap

The cppIsekai baseline has `knowledge/architecture/fts5-search-dedup.md`:
FTS5 + `bm25()` ranking for title search and dedup candidates, replacing LIKE.
The goIsekai2 capability specs (`openspec/specs/store/spec.md`,
`openspec/specs/library/spec.md`) contain no FTS5 requirement, virtual table or
MATCH scenario. This is a missing requirement, not a driver limitation — it
must be added (as a delta or a spec MODIFIED entry) before `add-store` locks the
frozen schema, so the index tables are part of the migration set from day one.

# Related Concepts
- [Store access: GORM with one raw-SQL boundary](../decisions/store-access.md): proves the chosen pure-Go GORM stack supports the FTS5 feature the baseline relies on
