---
type: Decision
title: "UI stack: templ + Tailwind + Alpine + htmx"
description: "goIsekai V2 SSR UI uses typed templ components, compiled Tailwind CSS, and locally vendored Alpine/htmx; generated assets are committed and embedded."
tags: [ui, templ, tailwind, alpine, htmx, ssr]
generated: { by: agent/mcp, at: "2026-11-21T16:03:57Z" }
status: stable
---

## Context

The user directed the goIsekai V2 UI to use templ, Tailwind CSS, Alpine.js and htmx. add-config already shipped /settings with stdlib html/template plus a hand-written fetch() script; that page must be ported, so this is a new OpenSpec change (add-ui-stack) with a MODIFIED config delta, not an edit to the archived change.

## Decision

| Piece | Version / form | Role |
|---|---|---|
| templ | github.com/a-h/templ v0.3.1118, CLI via go install | typed Go SSR components; generates *_templ.go committed to git |
| Tailwind CSS | v4.3.5, standalone CLI (no Node at runtime) | utility CSS compiled to one static file |
| Alpine.js | v3.17.5, vendored minified JS | local component state (dirty fields, toggles, banners) |
| htmx | v2.0.11, vendored minified JS | declarative server fragments: form submit, status swap, nav |

All assets are embedded with go:embed and served locally; no CDN and no runtime Node/JS build. make generate runs templ generate plus the Tailwind CLI; make build depends on it, and a gate fails when generated files are stale. CGO_ENABLED=0 remains a hard gate.

## Why

- templ replaces stringly html/template with compiled, type-checked components while keeping SSR and httptest; it does not add a client runtime.
- Tailwind removes bespoke CSS authoring but needs one build step; static output is embedded like any other asset.
- Alpine handles local UI state; htmx handles server round trips. They are deliberately kept separate: no client-side routing, no SPA, no duplicated state between them.
- The alternative (keep html/template + vanilla JS) was rejected by the user; React/Vue/Vite were rejected as SPA build toolchains that fight the local-first daemon model.

## Consequences and review triggers

- /settings must keep its frozen contract: section-grouped rows, doc/value/source, env/flag rows disabled with a hint, restart badge, dirty-field submit, nav from the implemented-pages list only.
- A UI change now carries generated Go/CSS plus the .templ source; review both, and run the freshness gate.
- Revisit if Tailwind standalone lags a platform we ship, or if templ codegen slows the edit loop enough to hurt development.

# Related Concepts
- [chi + coder/websocket for HTTP and WebSocket](http-framework.md): replaces the html/template + htmx fragment note with the full four-piece UI stack
- [Stack: libraries vs hand-written code](../architecture/stack.md): UI stack rows belong in the consolidated dependency review
