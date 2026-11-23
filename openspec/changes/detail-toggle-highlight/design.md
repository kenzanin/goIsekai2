# Design: detail-toggle-highlight

## D1. Single toggle

`saveToggleRoute`, `saveToggleVals`, `saveToggleLabel`, `saveToggleTitle` and `saveToggleClass` derive one button from `row.Saved`. Saved uses an amber border/background and `aria-pressed="true"`; unsaved uses the neutral border and `aria-pressed="false"`.

## D2. Balanced action row

The title header carries only the name and plugin badge. A single `aria-label="Manga actions"` row below the metadata holds the toggle on the left, Start/Continue in the middle and Fetch/Force/Reset on the right using `justify-between` with wrapping.
