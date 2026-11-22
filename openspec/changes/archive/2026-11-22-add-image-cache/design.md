# Design: add-image-cache

## Context

See `proposal.md`. The `image-cache` spec freezes layout, LRU, transcode and offline behavior.
cppIsekai `cache-layout` is the baseline; V2 is greenfield so no upgrade path exists.

## Goals / Non-Goals

- Goals: frozen v3 tree, bounded shared pool, atomic single-flight writes, pure-Go transcode, offline
  reads, sliding prefetch.
- Non-Goals: site plugins, new image formats beyond the archived codecs, reader UI changes.

## Decisions

### D1 — Layout and keys

Paths derive from placement context only: `(plugin, slug, chapter number, chapter id, page index)`;
covers use `cover/<urlhash>`. Keys omit extension so a format switch keeps serving old files.
Baseline: cppIsekai v3 cache tree.

### D2 — Shared LRU

One index over image bodies and API payloads in `data/cache/`, capped by `cache_max_mb`. Overflow
evicts least-recently-used files plus sidecars; restart rebuilds order from mtimes. Baseline: cppIsekai
disk LRU.

### D3 — Atomic writes and single flight

Write to a same-dir temp file, decode/transcode fully, then rename. A per-key flight map collapses
concurrent misses; short negative cache prevents hammering a dead host. Baseline: cppIsekai fetch path.

### D4 — Transcode

`image_format` comes from config (`webp` default, `none` passthrough). WebP uses the pure-Go encoder;
AVIF/JXL use the archived wasm codecs; undecodable input is stored unchanged with a warning. No cgo,
no system libraries. Baseline: cppIsekai codec set.

### D5 — Cover ratio

Covers are resized to `cover_max_px` preserving aspect; the host computes the ratio from decoded
dimensions and writes a JSON sidecar. Client `ratio` is ignored. Baseline: cppIsekai cover handling.

### D6 — Serve-through and offline

`/cache/img` serves hits with content type, immutable cache-control and `X-Image-Ratio`; misses fetch
through the configured proxy, store, then serve; offline misses return 404. Baseline: cppIsekai
serve-through route.

### D7 — Prefetch window

`reader` computes the next 10 unread pages after the current position, crossing chapters, excluding
skipped chapters and already cached keys, and calls the cache directly. Baseline: cppIsekai 10-url
window.

## Risks / Trade-offs

- [codec size] → wasm codecs only for avif/jxl; default webp stays small.
- [LRU rebuild cost] → one directory walk at boot; acceptable for local libraries.
- [stale sidecars] → sidecar name derives from the image key and is evicted with the parent.

## Migration Plan

Greenfield; no legacy tree to purge. Rollback is reverting the change commit.

## Open Questions

None. Layout and behavior are fixed by the capability spec.
