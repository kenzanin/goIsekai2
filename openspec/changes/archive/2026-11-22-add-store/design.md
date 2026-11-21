# Design: add-store

## Context

See `proposal.md`. The `store` capability spec already freezes the table inventory, vocabularies and
SQL-level behaviors; this design covers only the implementation approach. cppIsekai `src/store/` is
the source of truth for column names, bitfields and query semantics. Any V2 addition (notably
`schema_version`) is recorded here, not silently introduced.

## Goals / Non-Goals

- Goals: one writer, frozen schema, forward-only migrations, typed API for later capabilities,
  raw SQL only behind `internal/store/sql`, all gates green without cgo.
- Non-Goals: library/reader/update domain rules, plugin host calls, new HTTP feature routes.

## Decisions

### D1 — GORM on the pure-Go driver

Use `gorm.io/gorm` v1.31.2 with `github.com/glebarez/sqlite` v1.11.0 over pinned
`modernc.org/sqlite` v1.62.0. Baseline: cppIsekai uses one SQLite connection serialized by its
`db_pool(1)`; V2 expresses the same rule as `sql.DB.SetMaxOpenConns(1)`.

Alternative: `gorm.io/driver/sqlite` (requires cgo via mattn), rejected by the no-cgo gate.

### D2 — Frozen DDL, no AutoMigrate

Migrations are numbered SQL files/constants inside `internal/store/sql/migrations.go`, each applied
in its own `BEGIN IMMEDIATE`; `schema_version` records the last successful number. Baseline: the
C++ schema is frozen (composite keys, verbatim status strings, bitfields, JSON blobs). V2 adds
`schema_version` because the baseline has no version marker.

Alternative: GORM `AutoMigrate` — rejected because reflection would rename/reinterpret baseline
columns.

### D3 — Anti-nesting transaction helper

`Store.Tx(ctx, fn)` owns the only `BEGIN IMMEDIATE`; a context marker rejects a second Tx with a
typed error naming the outer transaction. Reads that feed a write happen before the write Tx.
Baseline: cppIsekai's `OpTrace` + `TxGuard` wedge hazard; Go has no recursive mutex, so nesting must
be rejected explicitly.

Alternative: savepoints — rejected because they hide the same bug behind a different name.

### D4 — Typed repositories over the SQL boundary

`internal/store` exposes typed methods (`SavedManga`, `Chapters`, `ReaderPositions`, `PluginState`,
`ApiCache`, `LibrarySearch`, `Backup`) and holds no SQL strings. All DDL/DML/MATCH statements live in
`internal/store/sql`, many files allowed, one package. Baseline: cppIsekai callers build SQL inline;
V2 diverges internally but keeps every observable result identical.

Alternative: raw `database/sql` at call sites — rejected by the user and by
`knowledge/conventions/sql-boundary`.

### D5 — FTS5 and BM25 dedup in the boundary

`library_fts` is created with the baseline DDL; search uses `MATCH ... ORDER BY bm25(library_fts)`
with prefix terms, candidate limit 20, token-multiset Dice thresholds from the baseline, union-find
grouping in one transaction. If FTS5 creation fails, warn once, fall back to LIKE and keep
`dup_group = 0`. Baseline: cppIsekai `store_dedup.cpp`.

Alternative: application-side scan — rejected as O(n²) and divergent from baseline ranking.

### D6 — Saved-stats rollup and bitfields

Chapter/read writes recompute `total_ch`, `read_ch` and the `is_done` bit inside the same Tx; flags
are changed with bit operations only. Status strings are stored verbatim, including `ongoin`/
`Onhaiatus`. Baseline: cppIsekai `saved_manga` rollup and `ChapterFlag` bit table.

Alternative: trust plugin-reported totals — rejected by the frozen spec.

### D7 — Boot order and shutdown

`run()` opens the store, migrates, validates FTS availability, recomputes dup groups, then opens the
listener. Shutdown reverses: HTTP first, store last. `--dump-config`/`--help` return before any
store call. Baseline: cppIsekai `main.cpp` boot/shutdown ladder.

Alternative: lazy store on first request — rejected because migration failure must block listening.

## Risks / Trade-offs

- [glebarez fork lags modernc] → pin all three versions explicitly; review on upgrade.
- [FTS5 unavailable in a build] → single warning + LIKE fallback + grouping disabled, per spec.
- [GORM single connection contention] → same serialization as baseline; background jobs use short
  transactions only.
- [Large in-memory result sets during dedup] → candidate limit 20 and batched union-find writes.

## Migration Plan

Fresh greenfield database only. No production data exists yet; migrations are forward-only and a
newer database than the binary is a hard boot error. Rollback is reverting the change commit.

## Open Questions

None. Schema, vocabularies and algorithms are fixed by the baseline spec.
