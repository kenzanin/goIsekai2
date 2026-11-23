# Design: ui-inplace-polish

## D1. Single swap target

`Base` owns `<main id="page-main">`. All internal links carry `hx-select="#page-main"`, `hx-target="#page-main"`, `hx-swap="outerHTML"` and `hx-push-url="true"`, so the server keeps rendering full pages and htmx extracts only the main element.

## D2. Reinitialisation

`htmx:afterSwap` reinitialises Alpine on the swapped target and updates the active header link from `location.pathname`. Mutations call `window.reloadMain()` to re-fetch the current URL into the same target.

## D3. Pager window

A shared `PageWindow` helper always shows the first four and last five pages, with `…` for skipped ranges. The library applies the same shape client-side in its Alpine state.

## D4. Detail controls

The saved toggle becomes a compact `highlight` indicator below the synopsis. Start/Continue uses the stored reader position and links to the exact chapter id, with `CH<n> P<m>` in the label.
