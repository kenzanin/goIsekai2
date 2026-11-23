---
type: decision
title: "Staged plugin JSON lives in api_cache, not a separate in-memory DB"
description: The raw detail/chapters envelopes fetched from a search-result action are stored in the main SQLite api_cache table with a TTL; Save promotes them into saved_manga.
tags: [decision, cache, sqlite, detail, save-flow]
generated: { by: agent/mcp, at: "2026-11-23T02:00:21Z" }
status: stable
---

## Decision

A search-result click is an explicit fetch action. The raw detail and chapters envelopes are written to
`api_cache` in the main `library.db` under the existing `(plugin_id, query_hash)` key with a TTL. The later
Save action promotes that cached JSON into `saved_manga` and sets `added_at`.

## Why not a separate in-memory SQLite

A second in-memory DB adds lifecycle and synchronization, loses staged data on restart, and conflicts with the
single-connection `MaxOpenConns(1)` serialization. The main DB already has the table, TTL, pruning and backup.

## Why not a virtual table

FTS5 virtual tables index text for search; they are not a blob store. Raw JSON stays in the ordinary `api_cache`
table and can be queried with `json_extract(payload, '$.name')` when needed.

## Consequences

- One DB to back up and inspect; staged JSON survives daemon restarts until TTL expiry.
- Reads from the library remain store-only; only `?fetch=1` from search stages new JSON.
- If no staged envelope exists, Save keeps its minimal-payload fallback.

# Related Concepts
- [Manga detail is a cache-only read path](cache-only-read-path.md): Search-result fetch is an allowed explicit action; reads stay cache-only.
- [saved_manga is the single master for manga data](saved-manga-master-data.md): Save promotes staged JSON into the master row.
