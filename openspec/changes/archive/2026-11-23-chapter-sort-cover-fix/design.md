# Design: chapter-sort-cover-fix

## D1. Chapter order

All display queries use `ORDER BY updated_at DESC, number DESC, chapter_id`. `updated_at` is the upstream release timestamp stored on each chapter row, so the newest release appears first even when numbers collide.

## D2. Cover unwrapping

`originalImageURL` recursively extracts `url` from `/cache/img?` values. The save endpoint applies it to every incoming cover, and the detail toggle now sends `DetailRow.RawCover` instead of the rendered proxy URL.
