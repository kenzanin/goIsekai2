# Proposal: reader-defaults

## Why

The reader currently opens in webtoon mode with fit-screen zoom and puts duplicate navigation in both the top and bottom overlay bars. The user wants a cleaner default: paged mode, fit width, navigation only in the top bar, and the back-to-detail control at the top-right.

## What Changes

- Default mode becomes `paged` and default zoom becomes `fit width` when no persisted value exists.
- Zoom is restored from the stored reader position, so a user's explicit zoom survives reopen.
- The bottom overlay bar is removed; chapter/page selects, prev/next chapter and the back-to-detail link live in the top bar.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `reader` — defaults, zoom restore and overlay layout.

## Impact

- `internal/http/components/reader.templ` and `internal/http/server.go` only; no store schema changes.

## Compatibility

Existing positions with non-default mode/direction/zoom are preserved and win over the new defaults.
