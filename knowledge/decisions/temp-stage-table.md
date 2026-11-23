---
type: decision
title: Staged search results use a SQLite TEMP table
description: Raw detail and chapters envelopes from a search-result click live in a per-connection TEMP table for process lifetime; api_cache remains the persistent TTL cache for search responses.
tags: [decision, sqlite, temp-table, cache, detail]
generated: { by: agent/mcp, at: "2026-11-23T02:50:20Z" }
status: stable
---

## Decision

`StageSearchResult` writes raw detail and chapter envelopes to `stage_detail`, a SQLite TEMP table on the
single daemon connection. `Detail` and `StagedChapters` read it; `PromoteCached` copies it into
`saved_manga`/`chapters` and deletes it. A daemon restart leaves staging empty.

## Why

The staged JSON only needs to survive until Save, not until restart. A TEMP table is queryable with SQL and
JSON1, needs no TTL/prune, and avoids a second in-memory database or a virtual table that cannot hold a blob.

## Complementary caches

- `api_cache` keeps persistent TTL-backed search/plugin responses, keyed by `(plugin, hash(query,page))`.
- `api_cache_view` exposes those payloads through JSON1 `json_extract` for inspection.
- `library_fts` remains the FTS5 index for local library title search.

## Consequences

- All raw SQL for staging and the view lives in `internal/store/sql`; callers use typed helpers.
- The TEMP table is created lazily on the active connection; `MaxOpenConns(1)` makes its lifetime the process.
- This replaces `decisions/staging-json-in-api-cache` for staged search results.

# Related Concepts
- [Manga detail is a cache-only read path](cache-only-read-path.md): Search-result staging is an explicit fetch action; reads stay store-only.
- [saved_manga is the single master for manga data](saved-manga-master-data.md): Save promotes staged JSON into the master row.
