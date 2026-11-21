---
type: Fact
title: "Library grid: client-only filter/pagination over data attributes"
description: "The library grid is fully client-filterable: SSR emits every card with data-plugin/data-dup/data-status-flags/data-status and one Alpine component paginates 12/page, so grid filters never hit the server."
tags: [internal/http/components, library.templ, alpine, ui]
generated: { by: agent/mcp, at: "2026-11-21T19:52:04Z" }
status: stable
---

## Verified

Implemented `internal/http/components/library.templ` (session 2026-11-22); gates:

- `templ fmt internal/http/components` → exit 0
- `make generate` → exit 0 (regenerates all `*_templ.go` + `app.css`)
- `CGO_ENABLED=0 go build ./internal/http/...` → exit 0
- `go run scripts/godoc-check.go` → exit 0

`grep -o 'data-status-flags' internal/http/components/library_templ.go` matches: the generated
markup emits every card with `data-plugin`, `data-dup`, `data-status-flags`, `data-status`, and a
single Alpine component (`libraryGridState`) owns page/plugin state client-side. No `?page=`/
server pagination exists for the grid.

## Consequence

- Grid filtering stays client-only: the future Complete chip gates on `status_flags & 32` and the
  duplicate chip on `dup_group != 0` read from the attributes (cppIsekai `library-save` bit layout:
  bit0 new, bit1 done, bits2-5 category, bit6 dim) — never by string-matching `data-status`.
- Any new grid filter or sort is an Alpine getter on the same data attributes; do not add server
  query params for the saved view (baseline: saved mode ignores `?q=`/`?page=`).

# Related Concepts
- [UI stack: templ + Tailwind + Alpine + htmx](../decisions/ui-stack.md): The grid contract is the consumer side of the frozen UI stack: templ SSR plus vendored Alpine, no new runtime dependency.
