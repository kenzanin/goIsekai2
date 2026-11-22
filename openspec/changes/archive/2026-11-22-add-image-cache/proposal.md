# Proposal: add-image-cache

## Why

`reader` is archived, so page URLs are host-owned `/cache/img` links. The image cache is next in the
porting order and is what makes offline reading, LRU bounding and pure-Go transcoding work.

## What Changes

- Add the v3 on-disk layout (`img/<plugin>/cover/...`, `img/<plugin>/manga/<slug>/ch<n>-<id8>/pNNN.<ext>`,
  `api/<plugin>/<method>/<hash>.json`) from day one, with no upgrade path.
- Add a shared LRU pool over image bodies and API payloads under `data/cache/`, capped by
  `cache_max_mb`, with restart index rebuild and sidecar eviction.
- Add `/cache/img` serve-through route with placement context required, atomic writes, single-flight
  per key, configured pure-Go transcode (webp default; avif/jxl via wasm), cover normalization with
  host-computed ratio sidecar, offline 404 and stale-serve metadata.
- Add sliding 10-page prefetch window that skips flagged chapters and avoids cached keys.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `image-cache` — adds daemon wiring (route before listen, config-driven cap/format) on top of the
  frozen cache contracts.

## Impact

- Code: new `internal/cache` package, `/cache/img` route, config knobs already declared.
- Dependencies: pure-Go codecs from the archived stack only; no cgo.
- Out of scope: site plugins, reader UI changes, scheduler.

## Baseline

cppIsekai `cache-layout` and `image-cache` specs define the tree, LRU and transcode behavior; V2 has
no legacy upgrade step because the store is greenfield.
