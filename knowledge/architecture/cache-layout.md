---
type: architecture
title: "Image cache: v3 layout, namespace by display name, slug rules, LRU and pure-Go transcode"
description: "V2 concept-v3 cache tree: display-name namespace, hash-suffixed slug dirs, six-digit chapter folders with last-eight id suffix, cover ratio in filename (no sidecar), placement-context requirement, byte-based LRU, SQLite api_cache as the sole response cache."
tags: [architecture, cache, images, lru, transcode, slug, parity]
generated: { by: agent/cli, at: "2026-11-21T12:19:11Z" }
status: stable
---

## Root and tree

One cache root (`data/cache`), one directory per plugin **namespace**, one tree per manga:

```text
data/cache/img/
  <plugin-ns>/
    cover/<hash16>-<r4>.<ext>
    manga/<slug-dir>/
      chNNNNNN[-<id8>]/
        p<nnn>.<ext>
```

Everything is derived from immutable inputs (url, slug, chapter, page). No row id, no DB lookup on the hot path,
so a cache path can be computed from a request alone.

## Namespace: display name, not plugin id

`cache_ns(plugin_id, name)` (`src/cache/slug.hpp:107`) slugifies the **display name** and falls back to the plugin id
when the name is blank or slugifies to `untitled`. Consequences, all intentional:

- Changing a plugin's id or folder keeps its cached files.
- Two plugins sharing a display name share one directory.
- The mapping is snapshotted at boot from `store.plugins()` (`src/main.cpp:118-125`); renaming a display name takes
  effect on the next restart, not mid-run.

## Slug rules (frozen)

`slug_base` (`src/cache/slug.hpp:38`):

- keep `[a-z0-9]` (lowercased) and every byte `>= 0x80` verbatim, so CJK titles stay distinct;
- collapse any run of other bytes to a single `-`, trim leading/trailing `-`;
- cap at 64 bytes **on a UTF-8 boundary**; blank result becomes `untitled`.

`slug_dir` (`src/cache/slug.hpp:94`): on-disk directory is `<first 40 chars>-<sha1(slug)[:8]>`, uniformly for every
length. The hash suffix also separates titles that share a long prefix.

Base-slug collisions (`manga_slug`, `src/cache/slug.hpp:65`): suffix `-<last 8 chars of the manga id>` **only** when
another id in the known set produces the same base slug; a lone manga is never suffixed.

## File names

| Item | Rule |
|---|---|
| Cover | `cover/<sha1(url or key)[:16]>.<ext>`; plugin empty means namespace `common` |
| Page | `manga/<slug-dir>/ch<pad(n)>[-<id8>]/p<pad(n)>.<ext>` |
| Padding | six digits (`ch2` -> `ch000002`, `ch1234` -> `ch001234`) so width never changes sorting order |
| Chapter id suffix | last 8 characters of the chapter id, omitted when the id is empty |
| Cover ratio | fixed-point `-<r4>` before the extension, where `r4 = round(ratio * 10000)`; no JSON sidecar for new covers |

The chapter-id suffix exists because multi-language releases share a chapter number; two `ch012` would otherwise
collide.

## Hit, miss, and the extension-free lookup

`Cache::img_ensure` (`src/cache/cache_img.cpp:44`) is the single wrap point:

1. **Hit** = extension-free match. Pages look for a prefix `.../p012.` in the in-memory index and skip `.json`;
   covers go through `cover_serve`. This is what makes a mid-flight `image_format` switch keep serving old files
   instead of missing them.
2. **Stale-hit guard**: every hit is verified with `exists()`; a file deleted under the daemon (manual wipe) drops
   the index entry and falls through to fetch instead of serving a 200 with an empty body.
3. **Recency**: hit bumps a monotonic tick on the entry -- that tick is the LRU order.
4. **Miss**: fetch -> transcode -> store at the v3 path with the ratio encoded in the name, then insert into the index.

## Placement context is mandatory

Correction against the v3 concept text: there is **no sha1 fallback** for images without context. `img_store`
(`src/cache/cache_img.cpp:29`) returns `nullopt` when there is neither a cover context nor a complete page context
(plugin, slug, chapter >= 0, page >= 1), and `img_ensure` refuses to fetch in that case too. Every file lives under
a plugin namespace or it is not stored. V2 keeps that rule: an unnameable image is not cached.

## LRU over bytes

- The index is an in-memory map `path -> {bytes, tick}`, built by a boot scan of the cache root.
- Eviction is by total bytes against `cache_max_mb`; victims are evicted oldest-tick first, one at a time, until
  under the cap.
- `cache.evicted` is published **after** the unlink, so no subscriber can observe a path that still exists.
- Sidecars count as files but not toward image bytes.

## Transcode pipeline

Knobs stay exactly as the baseline has them (`[cache]`):

| Knob | Default | Effect |
|---|---|---|
| `image_format` | `webp` | target ext: `webp`, `avif`, `jxl`, `none` |
| `cover_max_px` | `800` | covers are downscaled to this longest edge; pages are never resized |
| `cache_max_mb` | policy default | LRU byte cap |

Ratio is computed by the host from the decoded image, encoded as fixed-point in the cover filename and served
as `X-Image-Ratio`; client-provided ratio parameters are ignored. Legacy JSON sidecars remain readable.

The Go pipeline is pure Go: decode with `image`/`x/image` (+ `gen2brain/*` for webp/avif/jxl), resize with
`golang.org/x/image/draw` (CatmullRom), encode with the selected format. See `decisions/image-pipeline` for why this
does not use libvips.

## Boundaries and invariants

1. `internal/cache` never opens SQLite. It takes a namespace provider (from the store) and a fetcher (tls-client).
2. Paths are computed, never accepted: no plugin or request may supply a relative path. Slug sanitising plus sha1
   naming is the traversal defence; `..` cannot appear because `.` is not an allowed slug byte.
3. The API response cache is a different thing entirely: SQLite `api_cache` is the sole response cache, not a
   disk `api/` tree. Prefetched page bodies land in the structured image tree with placement context.
4. A cover and a page with the same URL are independent entries (different trees), which is why covers key on the
   URL hash and pages key on placement.

## Tests that make this real

1. Naming table test: a fixture list of (title, id, chapter, page, url) -> expected relative path, including CJK
   titles, a 202-char slug, a collision pair, `ch 2`, `ch 1234`, chapter id shorter than 8, and a blank display name;
V2 asserts six-digit chapter width and the hashed slug directory.
2. LRU test: insert N files over the cap, assert exact eviction order by tick and that every evicted path is gone.
3. Single-flight test: two concurrent `Ensure` calls for one key produce exactly one fetch.
4. Extension-free hit test: store as `.webp`, set `image_format=none`, assert the old file still serves.

# Related Concepts
- [Image transcode without libvips: gen2brain WASM-transpiled webp/avif/jxl, pure-Go resize](../decisions/image-pipeline.md): implements the transcode step with cgo-free codecs
