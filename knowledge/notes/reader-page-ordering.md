---
type: Decision
title: "Reader page order: (number, id) sort with positional neighbours"
description: "Reader chapters are ordered by (number, chapter_id) and prev/next are positional neighbours, so duplicate numbers stay individually reachable."
tags: [internal/reader, service.go, reader-page]
generated: { by: agent/mcp, at: "2026-11-22T00:19:32Z" }
status: stable
---

generated: { by: "kit agent", at: "2026-11-22" }

Ported from cppIsekai `src/http/pages_api.cpp` (design D3, lines 69-93): the chapter list is
stable-sorted by `(number, chapter_id)` and `prev/next` are the entries adjacent by position,
NOT the closest distinct number. Equal numbers therefore become consecutive navigation stops
(mangafreak `_24a`/`_24b`), and an exact `ch_id` match always beats the number fallback.

Evidence:
- cpp source: `std::stable_sort(... a.first != b.first ? a.first < b.first : a.second < b.second)`;
  `target_idx` by id first, else first `number` match; neighbours are `order[target_idx +- 1]`.
- Go port verified: `CGO_ENABLED=0 go test ./internal/reader/...` -> `ok goisekai2/internal/reader 0.043s` (exit 0),
  including `TestPagesDuplicateNumbers` (two chapters numbered 24: number-only request lands on `24a`,
  `next_ch` stays 24 with `next_ch_id=24b`; `ch_id=24b` mirrors back `prev_ch_id=24a`).

Consequence: `internal/reader` owns this ordering; the future http pages route must consume
`Service.Pages` output as-is and must not re-sort or dedupe by number, or duplicates collapse.

# Related Concepts
- [Data flows: five ordered sequences and the invariants that keep them reproducible](../architecture/data-flows.md): Ordering decision realized by internal/reader/service.go Pages/Window.
