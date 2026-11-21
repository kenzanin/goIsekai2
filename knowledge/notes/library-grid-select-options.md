---
type: Decision
title: SSR the plugin filter options; Alpine only hydrates the deep-link value
description: "Library grid filter options are server-rendered and only the ?plugin= value is hydrated by Alpine, because x-model desyncs against x-for-generated options."
tags: [internal/http/components, library.templ, alpine, ui]
generated: { by: agent/mcp, at: "2026-11-21T19:52:04Z" }
status: stable
---

## Context

`templ Library` needs a plugin filter seeded from the `?plugin=` deep link, but the component
signature has no request access, so the selected value is only known client-side.

## Decision

Render `<option>` elements server-side from `libraryPlugins(rows)` (sorted distinct `Plugin` values)
and let the Alpine `init()` set `plugin` from `URLSearchParams`, relying on `x-model` to sync the
select. Rejected: `x-for` inside `<select>` to emit options — Alpine `x-model` initializes before
dynamically added options exist, so the visible selection desyncs from state and needs an extra
`x-effect` rebind.

## Evidence

`make generate` exit 0 and `CGO_ENABLED=0 go build ./internal/http/...` exit 0 with the SSR options
in `library_templ.go`; the `?plugin=` value not present among options degrades to the count line
(`filtered.length`), not a broken select.

## Consequence

New grid filters follow the same split: choices SSR'd where the value is derivable from rows, only
the deep-link-seeded state hydrated by Alpine `init()`.

# Related Concepts
- [UI stack: templ + Tailwind + Alpine + htmx](../decisions/ui-stack.md): The SSR-options-vs-x-for decision concrete applies the ui-stack decision's templ+Alpine split to form controls.
