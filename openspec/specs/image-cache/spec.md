# image-cache Specification

## Purpose

Defines the on-disk cache: the browsable per-plugin layout, the global transcode format, cover
normalization with a host-computed ratio, the serve-through image route, the shared LRU cap, the
offline/stale-serve behaviour that keeps an opened manga readable without the network, and the reader
prefetch window. Everything is pure Go: no libvips, no cgo, and no system image library.

### Layout

| Target | Path |
|---|---|
| cover | `img/<plugin>/cover/<urlhash>.<ext>` |
| page | `img/<plugin>/manga/<slug>/ch<n>-<id8>/pNNN.<ext>` |
| api payload | `api/<plugin>/<method>/<hash>.json` |

`<slug>` is the normalised display name, `<id8>` is the first 8 characters of the chapter id,
`pNNN` is zero-padded to three digits so lexical order is reading order, and `<ext>` follows the
stored format. A read matches by key without extension, so any stored format serves.

## Requirements

### Requirement: Shared LRU pool

Image bodies and API payloads SHALL live in one shared pool under `data/cache/` capped by
`cache_max_mb` (default 2048): on overflow the least-recently-used files are evicted. Presence and
recency SHALL survive restart with the index rebuilt from disk, and eviction SHALL remove a cover's
sidecar with the image.

#### Scenario: Insert evicts the oldest
- **WHEN** the pool is at its cap and a new image is stored
- **THEN** the least-recently-used cached file is deleted and the new file is present

#### Scenario: Restart keeps the cache
- **WHEN** the host restarts with a populated cache
- **THEN** cached images still serve and still count toward the cap

### Requirement: Placement context is required

A cache write SHALL require placement context (plugin, slug, chapter number, chapter id, page index,
and `cover` for covers). A request without it SHALL be refused with a 400 and SHALL NOT write a file:
there is no hashed fallback path, because a file without placement can never be prefetched, evicted
by chapter, or served offline from the reader. This is a recorded divergence from the baseline spec,
which describes a fallback the baseline code does not implement.

#### Scenario: Page lands in its chapter directory
- **WHEN** page 3 of chapter 12 (id `8f3a2c...`) of `solo-leveling` under `mangadex` is stored
- **THEN** the file is written at `img/mangadex/manga/solo-leveling/ch12-8f3a2c/p003.webp`

#### Scenario: No context is refused
- **WHEN** `/cache/img?url=...` is requested without plugin/slug/chapter/page
- **THEN** the route answers 400 and nothing is written to disk

### Requirement: Atomic writes

A cache write SHALL go to a temporary file in the target directory and be renamed into place only
after a complete, decodable body is available. A killed fetch or a truncated response SHALL leave no
partial file and SHALL not poison a later read.

#### Scenario: Interrupted fetch
- **WHEN** the upstream connection drops mid-body
- **THEN** no file appears at the target path and the next request fetches again

### Requirement: Single flight per key

Concurrent requests for the same uncached key SHALL collapse into one upstream fetch; the others wait
for its result. A failed fetch SHALL be remembered briefly so a burst does not hammer a dead host.

#### Scenario: Two eager readers
- **WHEN** two page requests for the same uncached URL arrive together
- **THEN** the upstream is fetched once and both responses carry the same bytes

### Requirement: Global transcode

Every newly stored image SHALL be transcoded to `[cache] image_format` (webp by default; avif and jxl
supported; `none` = passthrough), resolved through the config registry. The encoders SHALL be pure
Go: `nativewebp` for webp, and the wazero-hosted WASM builds for avif and jxl with the
`purego`/dlopen path disabled so no system library is required. If the input cannot be decoded, the
original bytes SHALL be stored unchanged with a warning.

#### Scenario: Configured format is stored
- **WHEN** `image_format=webp` and a jpeg page is fetched
- **THEN** the stored file has extension `.webp` and the response content type is `image/webp`

#### Scenario: Format switch keeps old files
- **WHEN** `image_format` changes from webp to avif while webp files exist
- **THEN** existing keys keep serving webp and new fetches store avif

#### Scenario: Undecodable input
- **WHEN** the fetched bytes cannot be decoded
- **THEN** the original bytes are stored as-is, a warning is logged, and the response serves normally

#### Scenario: No system library at run time
- **WHEN** the binary runs on a host without libwebp, libavif or libjxl installed
- **THEN** transcode still works through the pure-Go encoders

### Requirement: Cover normalization

A cover stored with `cover=1` SHALL be resized to fit within the configured maximum dimension
(default 800 px) preserving aspect ratio. The ratio SHALL be computed by the host from the decoded
dimensions at store time, persisted in a JSON sidecar beside the cover, and exposed as
`X-Image-Ratio` on the serve response. A client-supplied `ratio` parameter SHALL be ignored; a
plugin-provided ratio MAY be used only when decoding fails, and otherwise the ratio is omitted.

#### Scenario: Oversized cover is resized
- **WHEN** a 2000x3000 cover is stored with `cover=1` and max 800
- **THEN** the stored file is at most 800 px on its longest side, aspect preserved, and the sidecar records the host-computed ratio

#### Scenario: Client ratio is ignored
- **WHEN** a request passes `ratio` in the query
- **THEN** the sidecar value is served instead

### Requirement: Serve-through image route

`GET /cache/img` SHALL serve a hit with its content type, `X-Image-Ratio` when the sidecar exists, and
an immutable cache-control header; on a miss with the network up it SHALL fetch, store, then serve; on
a miss with the network down it SHALL answer 404.

#### Scenario: Second view is local
- **WHEN** the same image URL is requested twice with the network up
- **THEN** the first populates the cache and the second is served from disk without an upstream call

#### Scenario: Offline miss
- **WHEN** the network is down and the page was never cached
- **THEN** the route answers 404

### Requirement: Offline through the proxy setting

The configured `proxy` SHALL apply to every outbound request — plugin `http.fetch` and image fetch —
so pointing it at an unreachable host simulates offline. Cached detail, chapter lists and page images
SHALL keep rendering, with stale-served responses marked `cached: true`; uncached requests SHALL fail
with the normal failure shape.

#### Scenario: Cached manga opens offline
- **WHEN** the proxy is unreachable and a previously viewed manga is opened
- **THEN** detail, chapter list and cached pages render, stale responses carry `cached: true`

#### Scenario: Never-viewed page offline
- **WHEN** the proxy is unreachable and an uncached page is requested
- **THEN** the image route answers 404 and the API route answers its failure shape

### Requirement: Stale-serve on failure

For opted-in read routes the host SHALL persist every successful plugin response keyed by plugin,
method and sorted query, and on a later failure SHALL return the last good payload with
`cached: true` when one exists, logging the underlying failure.

#### Scenario: Upstream down, cache answers
- **WHEN** a plugin call fails and the identical query succeeded before
- **THEN** the route returns the previous payload with `cached: true`

#### Scenario: No history, real error
- **WHEN** a plugin call fails and nothing was ever stored for that query
- **THEN** the route returns the failure shape with no `cached` flag

### Requirement: Sliding prefetch window

`POST /reader/position` with `{chapter, page}` SHALL recompute the next 10 unread pages in ascending
chapter order, excluding skipped chapters, starting after the current page and crossing chapter
boundaries, and SHALL prefetch only the not-yet-cached ones. Toggling a chapter's skip SHALL re-run the
window immediately.

#### Scenario: Boundary spill
- **WHEN** the position is the last page of chapter 1 and chapter 2 is not skipped
- **THEN** the remaining window pages come from the start of chapter 2

#### Scenario: Skip reroutes the window
- **WHEN** chapter 2 is marked skipped at the same position
- **THEN** the window pages come from chapter 3 and no chapter 2 page is fetched

#### Scenario: Cached pages are not refetched
- **WHEN** some window pages are already on disk
- **THEN** only the missing ones are fetched
