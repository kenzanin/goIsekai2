# Proposal: chapter-sort-cover-fix

## Why

The user reported two defects on a saved manga detail page: chapter rows were oldest-first while the requested order is newest-first, and unsaving then saving stored the already-proxied `/cache/img` URL as the cover so the next render double-wrapped it and the image disappeared.

## What Changes

- Chapter lists on detail, `/chapters` and `/api/chapters` sort by release time descending, then number descending.
- The save endpoint unwraps any `/cache/img` layers before persisting a cover; the detail toggle sends the raw cover URL.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `detail` — chapter order and cover persistence.
- `library` — cover payload is always the upstream URL.

## Impact

No schema change; existing proxy-shaped cover values are repaired at save time.
