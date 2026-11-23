# Design: cache-layout-concept

## D1. Namespace

`Namespace(pluginID, displayName)` slugifies the display name and falls back to the plugin id. The cache key is stable across folder renames while remaining fs-safe.

## D2. Manga directory

`SlugDir` normalizes the slug, keeps at most 40 characters and appends eight hex digits of a hash of the full slug. Long and colliding titles stay distinct.

## D3. Chapter folder

`chapterDir` uses six-digit zero padding so chapter 999 and 1000 sort in the same width. A last-eight `chapter_id` suffix separates duplicate chapter numbers; fractional numbers keep their decimal part.

## D4. Cover ratio

The fixed-point ratio (ratio × 10000, rounded) is encoded in the cover filename. No `.json` sidecar is written for new covers; legacy sidecars remain readable.

## D5. Search and library covers

Both pages render `imageSrc(...)` URLs. The `/cache/img` handler alone decides the on-disk landing path.

## D6. No disk API cache

`api_cache` in SQLite is authoritative for plugin/search/enrichment responses. Prefetch stores page bodies under the image layout, not `api/prefetch`.
