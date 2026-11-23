# Design: reader-defaults

## D1. Default source of truth

The server reads the stored position. Empty mode/direction/zoom fields are replaced with `paged`, `ltr` and `width` before rendering. The component also keeps the same fallbacks for direct construction.

## D2. Zoom persistence

`readerBootstrap` now receives and emits `zoom`. The Alpine state initialises `zoom` from the bootstrap value, so a user's saved zoom is restored rather than silently reset.

## D3. Single navigation bar

The bottom overlay panel is removed entirely. The top bar keeps chapter select, page select, adjacent chapter buttons, mode/direction/zoom selectors and the back-to-detail link at the far right. The dim `p N / M` pill stays while the overlay is hidden.
