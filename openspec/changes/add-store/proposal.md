# Proposal: add-store

## Why

`config` and the UI stack are archived, but the daemon still has no persistence layer. Every later
capability (library, reader, updates, plugins) depends on the frozen cppIsekai schema, the
single-writer rule and the typed repository surface. This change implements the `store` capability
before any feature writes data into it.

## What Changes

- Add the pure-Go GORM store (`gorm.io/gorm` over `github.com/glebarez/sqlite`) with
  `MaxOpenConns(1)`, WAL, foreign keys, busy timeout, and an anti-nesting `IMMEDIATE` transaction
  helper.
- Apply the frozen 11-table schema through forward-only numbered migrations in `schema_version`;
  `AutoMigrate` stays disabled.
- Implement typed repositories for plugin state, reader positions, saved manga, title metadata,
  chapters/pages, update attempts, API cache, FTS5 search, dedup groups, source migration and
  backup.
- Wire store bootstrap into daemon startup before the HTTP port opens; keep `--dump-config` and
  `--help` database-free.
- Keep all raw SQL inside `internal/store/sql`, enforced by `make lint`.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `store` — adds the daemon-level lifecycle requirement (migrations before listen, no DB touch on
  `--dump-config`/`--help`) on top of the existing frozen-schema requirements.

## Impact

- Code: new packages `internal/store` and `internal/store/sql`; `cmd/goisekai2` gains store boot and
  shutdown.
- Dependencies: `gorm.io/gorm`, `github.com/glebarez/sqlite`, pinned `modernc.org/sqlite`.
- No cgo: `CGO_ENABLED=0 go build ./...` remains the hard gate.
- Out of scope: library/reader/update business rules, plugin runtime, HTTP routes beyond store
  bootstrap, tray.

## Baseline

cppIsekai `src/store/` is the behavioral source of truth for table columns, status vocabulary,
bitfields, BM25 dedup and read-boundary migration. V2 adds `schema_version` because the baseline
has no `user_version` marker.
