# Proposal: detail-toggle-highlight

## Why

The previous polish used the literal label `highlight` and left action buttons split between the title area and a panel below the synopsis. The user meant the saved toggle should *be highlighted* when active, and all detail actions should live in one balanced row below the synopsis.

## What Changes

- One toggle button below the synopsis: `☆ Save` in the normal state, `★ Saved` highlighted in amber when saved; clicking toggles membership.
- Start/Continue and Fetch/Force/Reset move into the same action row, distributed with `justify-between`.
- No buttons remain clustered under the title.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `detail` — saved toggle appearance and action placement.

## Impact

Presentation only; endpoints and persistence are unchanged.
