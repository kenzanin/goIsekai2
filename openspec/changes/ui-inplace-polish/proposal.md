# Proposal: ui-inplace-polish

## Why

Every menu and internal page link currently performs a full document reload, and several mutations call `location.reload()`. The user wants in-place data via HTTP/htmx, a compact long pagination format with ellipsis on saved-manga pages, a shorter saved indicator under the synopsis, and a visible start/continue reading control.

## What Changes

- Navigation, search, detail, chapter and reader exits swap only `#page-main` via htmx with history push; Alpine is reinitialised after each swap.
- Mutations refresh the current view in place instead of reloading the page.
- Library/detail/chapter pagers use a long format with leading/trailing page runs and ellipsis.
- The saved "highlight" indicator moves below the synopsis; a Start/Continue reading button is added with the stored chapter and page.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `ui-shell` — in-place navigation.
- `library` — long pager format.
- `detail` — highlight placement and resume control.

## Impact

- Presentation and HTTP component wiring only; no store schema or service behavior changes.

## Baseline

Pure V2 UI polish; no cppIsekai contract change.
