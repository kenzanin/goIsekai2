# Spec Delta

## MODIFIED Requirements

### Requirement: v3 layout from day one, no legacy upgrade path

Fresh directories SHALL use concept-v3 paths: `img/<ns>/manga/<slugdir>/chNNNNNN[-id8]/pNNN.<ext>`, where `ns` is the plugin display-name slug with id fallback, `slugdir` is a bounded hash-suffixed slug, and `id8` is the last eight chapter-id characters. No migration runs; old paths are re-fetched.

#### Scenario: First image on an empty data dir
- **WHEN** the first page image of a chapter is cached on an empty data dir
- **THEN** its path uses the namespace, hashed slug dir and six-digit chapter folder

#### Scenario: Chapter width does not collide
- **WHEN** chapters 999 and 1000 are cached
- **THEN** both folders share the same six-digit width and sort numerically

### Requirement: Cover normalization

Covers SHALL be resized and their host-computed fixed-point ratio encoded in the filename (`-<r4>.<ext>`, r4 = round(ratio × 10000)). New covers SHALL NOT write a JSON ratio sidecar; legacy sidecars remain readable. `X-Image-Ratio` SHALL be derived from the filename.

#### Scenario: Oversized cover is resized
- **WHEN** a 2000x3000 cover is stored with `cover=1` and max 800
- **THEN** the stored file is at most 800 px on its longest side, aspect preserved, and its filename records the host-computed ratio

#### Scenario: Ratio rides in the filename
- **WHEN** an oversized cover is stored
- **THEN** its name carries the fixed-point ratio and the serve response exposes it as `X-Image-Ratio`

#### Scenario: Client ratio is ignored
- **WHEN** a request passes `ratio` in the query
- **THEN** the host-computed filename ratio is served instead

### Requirement: Sliding prefetch window

Prefetched pages SHALL be written to the structured image layout using their chapter and page placement context. No response body SHALL be written to an on-disk API tree.

#### Scenario: Boundary spill
- **WHEN** the position is the last page of chapter 1 and chapter 2 is not skipped
- **THEN** the remaining window pages come from the start of chapter 2

#### Scenario: Skip reroutes the window
- **WHEN** chapter 2 is marked skipped at the same position
- **THEN** the window pages come from chapter 3 and no chapter 2 page is fetched

#### Scenario: Cached pages are not refetched
- **WHEN** some window pages are already on disk
- **THEN** only the missing ones are fetched

#### Scenario: Prefetch lands as page images
- **WHEN** the window fetches an upcoming page
- **THEN** its bytes land under `img/<ns>/manga/<slugdir>/chNNNNNN/pNNN.<ext>`

#### Scenario: No disk API cache
- **WHEN** plugin or search responses are cached
- **THEN** they live only in the SQLite `api_cache` table
