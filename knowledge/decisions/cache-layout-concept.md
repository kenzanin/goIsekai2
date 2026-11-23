---
type: decision
title: Cache layout follows cppIsekai concepts with V2 simplifications
description: "V2 image paths adopt display-name namespace, hash-suffixed slug dirs, six-digit chapter folders with last-eight id suffix, URL-hash covers with ratio in the filename, and SQLite api_cache as the sole response cache."
tags: [decision, cache, layout, images, sqlite]
generated: { by: agent/mcp, at: "2026-11-23T03:31:48Z" }
status: stable
---

## Decision

V2 keeps the cppIsekai cache concepts but not byte-for-byte naming:

- Namespace = plugin display-name slug with plugin-id fallback.
- Manga directory = normalized slug truncated to 40 characters plus an eight-character hash of the full slug.
- Chapter folder = six-digit zero padding plus the last eight characters of the chapter id when present. Six digits avoid the ch999/ch1000 sort-width trap present in the baseline.
- Cover = URL-hash plus a fixed-point ratio suffix (`-r4`) in the filename; no JSON sidecar is written for new covers.
- Response cache = SQLite `api_cache` only; the on-disk `api/` tree and `APIPath` are removed.
- Prefetch writes page bodies into the structured image tree with chapter/page context.

## Alternatives considered

Exact baseline parity was rejected because V2 already uses SHA-256 hashes, a different cache root and a SQLite response cache; matching only the concepts keeps parity where it matters (stable names, collision safety, no context-free writes).

## Consequences

Library and search thumbnails both go through `/cache/img`; browser hotlink blocks cannot break them. Old paths are not migrated — a miss simply re-fetches under the new names.

# Related Concepts
- [Image cache: v3 layout, namespace by display name, slug rules, LRU and pure-Go transcode](../architecture/cache-layout.md): Records the V2 concept alignment and simplifications for the image cache tree.
- [Staged search results use a SQLite TEMP table](temp-stage-table.md): SQLite api_cache remains the sole persistent response cache.
