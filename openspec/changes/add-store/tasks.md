# Tasks: add-store

## 1. Connection and transaction boundary

- [ ] 1.1 Add GORM + glebarez/sqlite dependencies with pinned modernc.org/sqlite — verify `go list -m gorm.io/gorm github.com/glebarez/sqlite modernc.org/sqlite` matches the decision and `CGO_ENABLED=0 go build ./...` exits 0.
- [ ] 1.2 Implement `internal/store/sql/connection.go`: open with `MaxOpenConns(1)`, WAL, busy timeout, foreign keys — verify a test asserts PRAGMA values through the typed connection helper, with no SQL outside the boundary.
- [ ] 1.3 Implement `internal/store/store.go` with `Store.Tx` using `BEGIN IMMEDIATE`, a context marker and a typed nested-Tx error — verify a test opens a nested Tx, asserts the typed error names the outer Tx, and counts commits.

## 2. Migrations and frozen schema

- [ ] 2.1 Put the frozen 11-table DDL in `internal/store/sql/migrations.go` as numbered constants — verify `scripts/sql-boundary-check.go` passes and a fresh database test lists exactly the frozen tables plus `schema_version`.
- [ ] 2.2 Implement the forward-only migration runner with per-migration transactions — verify a failing migration test leaves `schema_version` at the previous number and a newer-database test refuses boot.
- [ ] 2.3 Add a typed `Store.Migrate(ctx)` and call it from daemon boot before `net.Listen` — verify `TestDaemonEndToEnd` still passes and a migration failure test asserts no listener was opened.

## 3. Typed repositories

- [ ] 3.1 Add saved-manga and chapter repositories (`internal/store/saved.go`) with bitfield helpers for status/flags — verify a test marks `read|skip` and folds `ongoin`/`Onhaiatus` while preserving verbatim status.
- [ ] 3.2 Add the same-transaction saved-stats rollup (`total_ch`, `read_ch`, `is_done`) — verify a test marks the last chapter read and asserts both counts and the bit.
- [ ] 3.3 Add reader-position and plugin-state repositories (`internal/store/positions.go`, `plugin_state.go`) — verify composite-key upsert tests preserve unrelated rows.
- [ ] 3.4 Add title-metadata and update-attempt repositories (`internal/store/metadata.go`) with JSON blobs and backoff fields — verify a round-trip test returns the same JSON and `backoff_until`.

## 4. Search, dedup and cache

- [ ] 4.1 Add FTS5 DDL, indexed row replace/delete and BM25 prefix search in `internal/store/sql/fts.go` plus typed methods — verify a search test orders by BM25 and a prefix term matches; a fallback test disables grouping when FTS5 creation fails.
- [ ] 4.2 Implement boot/save dedup recomputation with token-multiset Dice and union-find — verify a three-title test gets one non-zero group and singletons stay 0.
- [ ] 4.3 Add the write-through `api_cache` repository with TTL hit/miss/invalidate — verify a fresh hit avoids the plugin callback and refresh deletes before write-through.

## 5. Migration of sources and backup

- [ ] 5.1 Implement read-boundary source migration (chapter numbers <= highest read, sub-chapters included, existing flags OR-ed) — verify a 1..8 target with source bound 6 marks 1..6 including 6.1 and refuses an already-saved target.
- [ ] 5.2 Implement `VACUUM INTO` backup with timestamped filename in the data directory — verify a backup taken during concurrent writes opens cleanly and contains the last committed row.

## 6. Integration and gates

- [ ] 6.1 Wire store open/close into `cmd/goisekai2/main.go` with shutdown after HTTP — verify `--dump-config` creates no database and the end-to-end daemon test still passes.
- [ ] 6.2 Run the full gate set — verify `make test lint`, `go vet ./...`, `gofmt -l .` empty, `openspec validate add-store --strict` and `CGO_ENABLED=0 go build ./...` all exit 0.

## Workflow follow-up

- Merge `change/add-store` into `main` with `--no-ff`.
- Archive the change with `openspec archive add-store -y` after the gates pass.
