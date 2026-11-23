---
type: decision
title: Manga detail is a cache-only read path
description: "Every field rendered on manga detail comes from the store; plugin and enrichment fetches happen only in explicit user actions or the update scheduler, never during GET rendering."
tags: [decision, detail, cache, api-first, scheduler]
generated: { by: agent/mcp, at: "2026-11-23T01:37:16Z" }
status: stable
---

## Decision

`GET /view/manga/:plugin/:slug`, `GET /api/detail/:plugin/:slug` and `GET /api/chapters/:plugin/:slug` SHALL read only stored rows and sidecars. Missing metadata renders as a placeholder with an explicit Fetch action; it never triggers a plugin call or an enrichment lookup.

## Why

A detail render previously called two enrichment plugins and could rematerialize from the source site, taking 8-22 seconds. Cached data was already available in milliseconds. Reads must stay predictable and local-first.

## Allowed fetch paths

- Explicit user action: the existing Fetch/Force action for detail+chapters, and `POST /api/enrichment/refresh/:plugin/:slug` for enrichment candidates.
- Auto-update scheduler: new-chapter fetches and backoff-driven refreshes.
- Never: page render, `GET /api/detail`, `GET /api/chapters`.

## Consequences

- Cold or unsaved titles show placeholders until the user fetches; the page still opens.
- Enrichment candidates are stored in `api_cache` with a TTL; a cache miss returns an empty list.
- `applyCandidate` reads stored candidates only, so it cannot silently fetch the website.

# Related Concepts
- [saved_manga is the single master for manga data](saved-manga-master-data.md): Cache-only reads use saved_manga as the single master row.
- [Explicit saved state on the saved_manga master row](saved-state-marker.md): Unsaved master rows may render placeholders; saved state remains an explicit marker.
