# store Specification

## Purpose

Defines the single SQLite store: the frozen table set and column semantics ported from the
baseline, the single-writer connection model, the forward-only migration runner with its own
`schema_version` marker, transaction rules that keep reads and writes out of the same deferred
snapshot, and the write-through API cache. The store is the only package that speaks SQL; plugins
reach data exclusively through host tools.

### Frozen schema

The table inventory lives here, outside the requirement text, because it is a verbatim baseline
inventory rather than a behaviour:

| Table | Primary key | Meaning |
|---|---|---|
| `plugin` | `id` (autoincrement) | plugin rows: `name`, `plugin_id` (stable key, empty for seed rows), `version`, `capabilities`, `enabled`, `runtime` |
| `title` | `id` (autoincrement) | legacy title rows keyed to a plugin row |
| `plugin_state` | `(plugin_id, key)` | host bookkeeping values; `plugin_id` is the plugin **folder id**, not the `plugin.id` FK |
| `reader_positions` | `(plugin, slug)` | one row per manga: `chapter_id`, `chapter_number`, `page`, `mode`, `direction`, `zoom`, `updated_at` |
| `saved_manga` | `(plugin, slug)` | one row per saved manga |
| `title_meta` | `(plugin, slug)` | enrichment payloads: `fetched` JSON array, `chosen` JSON object, `fetched_at` watermark |
| `update_attempts` | `(plugin, slug)` | `last_attempt_at`, `outcome` (`ok|error|skipped`), `error`, `consecutive_errors`, `backoff_until` |
| `chapters` | `(plugin, slug, chapter_id)` | `number` (double, display/sort only), `title`, `url`, `flags`, `updated_at` |
| `chapter_pages` | `(plugin, slug, chapter_id, idx)` | `idx` 0-based reading order, `url` of the source image |
| `api_cache` | `(plugin_id, query_hash)` | `params_json`, `fetched_at`, `ttl_s`, payload stored as a JSON blob |
| `library_fts` | none (FTS5 virtual table) | `plugin UNINDEXED`, `slug UNINDEXED`, `name`, `alt_titles` (raw JSON), `synopsis`; one row per indexed title |

`saved_manga` carries: `name`, `cover`, `added_at`, `status` (verbatim from the source), `total_ch`,
`read_ch`, `last_read_at`, `updated_at`, `synopsis`, `alt_titles` (JSON array), `dup_group`,
`status_flags` (bitfield), `authors`, `artists`, `tags` (JSON arrays), `synopsis_display`, `pinned`.

## Requirements

### Requirement: Single writer connection

The store SHALL open SQLite with `MaxOpenConns(1)`, WAL journaling, a busy timeout, and foreign
keys enabled. Every write SHALL go through that one connection. The store SHALL expose a
transaction helper that begins `IMMEDIATE`, rejects nesting, and returns an error when a caller
tries to open a second transaction.

#### Scenario: Concurrent callers serialize
- **WHEN** two goroutines write at the same time
- **THEN** one waits for the other and both commits land, with no `database is locked` surfaced to the caller

#### Scenario: Nested transaction is rejected
- **WHEN** code inside an open transaction calls the transaction helper again
- **THEN** the helper returns a typed error naming the outer transaction and no second `BEGIN` is issued

#### Scenario: No read-upgrade inside a write transaction
- **WHEN** a caller needs to read before writing
- **THEN** the read happens in its own transaction (or before the write transaction starts), so the write never upgrades a deferred snapshot and cannot fail with BusySnapshot

### Requirement: Frozen table set

The store SHALL persist exactly the tables, keys and column meanings listed under *Frozen schema*,
and SHALL NOT rename, drop or repurpose a baseline column. V2 may add columns, but only through a
migration and never by reinterpreting an existing one.

#### Scenario: Chapter identity survives renumbering
- **WHEN** a plugin retcons chapter numbers but keeps chapter ids
- **THEN** the same `chapters` rows are updated in place and read flags are preserved

#### Scenario: Unknown column value is preserved
- **WHEN** a plugin reports a status string the host does not recognise
- **THEN** it is stored verbatim, no error is raised, and the derived `status_flags` keeps the last known category

### Requirement: Status and flag vocabularies

`saved_manga.status_flags` SHALL be a bitfield: bit0 `is_new`, bit1 `is_done`, bit2 `ongoing`,
bit3 `hiatus`, bit4 `canceled/cancelled`, bit5 `completed`, bit6 `is_dim`. `chapters.flags` SHALL
be a bitfield: bit0 `read`, bit1 `skip`, with bits 2 and above reserved. Flags SHALL be updated by
bit operations, never rewritten wholesale.

#### Scenario: Marking read sets one bit
- **WHEN** a chapter is marked read while its skip bit is already set
- **THEN** the stored flags value equals `read|skip`

#### Scenario: Status spelling folds into a category
- **WHEN** the source reports `ongoin` or `Onhaiatus`
- **THEN** it is accepted (case-insensitive, common misspellings tolerated) and folded into the matching category bit while the verbatim string stays in `status`

### Requirement: Saved-stats rollup

After chapter or read-state writes, the store SHALL recompute `total_ch`, `read_ch` and the
`is_done` bit for the affected `(plugin, slug)` from its own rows, in the same transaction as the
write that made them stale. The rollup SHALL NOT trust values reported by the plugin.

#### Scenario: Reading the last page finishes the title
- **WHEN** the highest unread chapter transitions to read and every chapter is read
- **THEN** `read_ch` equals `total_ch` and the `is_done` bit is set on the saved row

#### Scenario: Plugin-reported totals do not override the store
- **WHEN** the plugin claims 500 chapters but the store holds 480
- **THEN** `total_ch` reflects the store's rows and the mismatch is logged once at debug level

### Requirement: Migration runner

The store SHALL keep a `schema_version` table and apply forward-only, numbered migrations at
boot before the HTTP port opens. A migration failure SHALL abort boot with the failing version and
statement logged, and SHALL leave the database at its previous version (each migration in its own
transaction). `--dump-config` and `--help` SHALL NOT touch the database.

#### Scenario: Fresh database
- **WHEN** the store opens a file with no tables
- **THEN** all migrations run in order, `schema_version` ends at the latest number, and the store is usable

#### Scenario: Failed migration
- **WHEN** migration 7 fails
- **THEN** the transaction rolls back, `schema_version` stays 6, and boot exits non-zero naming migration 7

#### Scenario: Newer database than the binary
- **WHEN** `schema_version` is greater than the binary's latest migration
- **THEN** boot refuses to start and says the database is newer than the binary, rather than downgrading

### Requirement: FTS5 title index

The store SHALL create `library_fts` at boot (`CREATE VIRTUAL TABLE IF NOT EXISTS library_fts USING
fts5(plugin UNINDEXED, slug UNINDEXED, name, alt_titles, synopsis)`). Saves, refreshes and source
migrations replace the row for `(plugin, slug)`; removals delete it. Local search SHALL use
`MATCH ... ORDER BY bm25(library_fts)` with prefix terms. DDL and MATCH stay in `internal/store/sql`.
If creation fails, warn once, keep LIKE search and leave `dup_group = 0`.

#### Scenario: Fresh database gets the index
- **WHEN** the store opens a file with no tables
- **THEN** `library_fts` is created before boot completes and an empty search returns no rows

#### Scenario: Refresh replaces the indexed row
- **WHEN** a title's synopsis or alternative titles change and the row is saved again
- **THEN** the index holds exactly one row for that `(plugin, slug)` with `name`, raw-JSON
  `alt_titles` and the new `synopsis`

#### Scenario: Ranked local search
- **WHEN** a query matches both a title name and another title's synopsis
- **THEN** results are ordered by `bm25(library_fts)`, best score first, and a prefix term matches

#### Scenario: FTS5 unavailable
- **WHEN** creating the virtual table fails
- **THEN** boot warns once, local search falls back to LIKE, and duplicate grouping stays disabled

### Requirement: Deduplication group recomputation

`dup_group` SHALL be recomputed at boot and after saves from current data, never trusted from disk.
Candidates SHALL come from `library_fts` via `MATCH ... ORDER BY bm25(library_fts) LIMIT 20`, filtered
by the baseline token-multiset similarity rules; grouping SHALL use union-find, with every group write
in one transaction.

#### Scenario: Groups are rebuilt at boot
- **WHEN** the daemon restarts with three titles that now share a synopsis and author
- **THEN** all three carry the same non-zero `dup_group` after boot even if they were written as `0`

#### Scenario: Baseline similarity and grouping rules
- **WHEN** BM25 candidates are checked with token-multiset Dice similarity
- **THEN** a title score >= 0.75, a name score >= 0.75, or 0.5*title + 0.5*synopsis >= 0.6 joins a group;
  singletons get `0`, multi-row groups get sequential ids from 1, and all writes share one transaction

### Requirement: Write-through API cache

`api_cache` SHALL store plugin responses keyed by `(plugin_id, query_hash)` with `params_json`,
`fetched_at`, `ttl_s` and the payload as a JSON blob. A read SHALL be a hit only while
`fetched_at + ttl_s` is in the future; a miss or expiry SHALL call the plugin and write through.
An explicit refresh SHALL invalidate the affected keys before calling the plugin.

#### Scenario: Fresh hit avoids the network
- **WHEN** the same search query repeats inside its TTL
- **THEN** the plugin is not called and the cached payload is returned

#### Scenario: Refresh invalidates
- **WHEN** `POST /api/refresh/:plugin/:slug` runs
- **THEN** the affected `api_cache` rows are deleted before the plugin call and the new payload is written back

### Requirement: Source migration

Migrating a saved entry to a different source SHALL carry read state **by chapter-number
boundary**: every target chapter whose number is less than or equal to the highest read source
chapter becomes read (sub-chapters such as 6.1 included when the bound is 6). Reader positions and
cached page files SHALL NOT carry. Flags already present on pre-materialised target rows SHALL be
OR-ed, never stomped.

#### Scenario: Read boundary carries
- **WHEN** source chapters up to 6 were read and the target has 1..8
- **THEN** target chapters 1..6 (including 6.1) are read and 7..8 are not

#### Scenario: Target already saved
- **WHEN** the target `(plugin, slug)` already exists in the library
- **THEN** the migration is refused and nothing is written

### Requirement: Backup and maintenance

The store SHALL support a consistent online backup to `library-YYYYmmdd-HHMMSS.db` inside the data
directory, and SHALL run its maintenance operations (vacuum or integrity check, as scheduled)
through the same single connection so they never race a write.

#### Scenario: Backup while running
- **WHEN** a backup is requested while scrapes are in flight
- **THEN** the produced file opens cleanly, contains the last committed state, and no write is lost or blocked beyond the backup duration

### Requirement: Daemon store lifecycle

The daemon SHALL open and migrate the configured SQLite library before it starts listening on the
HTTP port. If a migration fails, the daemon SHALL exit non-zero, name the failing version and leave
the previous `schema_version` intact. The `--dump-config` and `--help` paths SHALL NOT open or create
the database. On graceful shutdown the daemon SHALL close the store after in-flight HTTP handlers
finish.

#### Scenario: Migration failure blocks listening
- **WHEN** a stored migration fails during daemon boot
- **THEN** the daemon exits non-zero naming the version, no HTTP listener is opened, and the database
  stays at the previous version

#### Scenario: Config dump does not touch the database
- **WHEN** `goisekai2 --dump-config <path>` runs against a data directory with no database
- **THEN** it writes the ini and exits without creating any SQLite file

#### Scenario: Store closes after HTTP shutdown
- **WHEN** the daemon receives SIGTERM while a request is in flight
- **THEN** the request completes, the HTTP server shuts down, and the store connection is closed last

#### Scenario: Fresh boot is usable
- **WHEN** the daemon starts against an empty data directory
- **THEN** all migrations run in order, `schema_version` reaches the latest number, and `/settings`
  becomes reachable

### Requirement: JSON1 cache view

The store SHALL expose a read-only `api_cache_view` created at open time. It SHALL project `params_json` and `payload` into queryable columns with SQLite JSON1 functions without duplicating stored payload bytes. The view is a V2 convenience object and SHALL NOT change the frozen table inventory.

#### Scenario: View exposes JSON fields
- **WHEN** an `api_cache` row is written using JSON params and payload
- **THEN** `api_cache_view` reports its `query_text`, `ok`, `total` and `bytes`

#### Scenario: View survives restart
- **WHEN** the daemon restarts with the same database
- **THEN** `api_cache_view` still exists because it is created idempotently at open

### Requirement: SQLite api_cache is the sole response cache

Plugin, search and enrichment response payloads SHALL live in the SQLite `api_cache` table with TTL. The host SHALL NOT write response bodies to an on-disk `api/` tree.

#### Scenario: Search response cached
- **WHEN** `/api/search` stores a plugin response
- **THEN** the payload is queryable through `api_cache_view` and no disk API file is created

#### Scenario: Prefetch uses the image tree
- **WHEN** a reader window prefetches page bodies
- **THEN** they land in the structured image cache rather than `api/prefetch`
