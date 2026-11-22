---
type: Convention
title: "API-first: JSON surface is the contract, SSR is a thin client"
description: Every capability ships a JSON endpoint and contract test first; templates and browser JS only project service data and call the dedicated API routes.
tags: [api, ui, architecture]
generated: { by: agent/mcp, at: "2026-11-22T05:41:03Z" }
status: stable
---

## Rule
goIsekai2 is API-first: the JSON surface is the contract, and the SSR UI is a thin client of it.

- Every user-visible capability SHALL have a named JSON endpoint and a contract test before a page is added.
- Page handlers SHALL render data from the same service methods the JSON handlers call; templates and inline JS SHALL NOT contain business rules, SQL, or direct store access.
- Mutations SHALL post to the dedicated JSON route (for example `POST /api/refresh/:plugin/:slug`); the baseline action-dispatch endpoint is a recorded divergence, not a pattern to copy.
- WebSocket streams are part of the API: `/ws/log` is a first-class route with a typed frame contract, not a UI-only shortcut.
- The About and Log pages render README and stream content but introduce no new data path; they consume the existing README asset and log hub.
- Layout work (including parity with the cppIsekai reference UI) SHALL change presentation only, never move logic from services/API into templates.

# Related Concepts
- [Architecture overview: layers, import rules and frozen external contracts](../architecture/overview.md): API-first extends the layer and import rules
- [UI stack: templ + Tailwind + Alpine + htmx](../decisions/ui-stack.md): SSR components remain projections over the API surface
