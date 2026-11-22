---
type: Process
title: "Capability porting order: one OpenSpec change per capability, config first"
description: "Authoritative dependency-ordered sequence for implementing the 11 V2 capabilities as OpenSpec changes, with per-change rules, gates and a status column that is updated when a change archives."
tags: [process, openspec, porting, sequencing, status]
generated: { by: agent/mcp, at: "2026-11-22T01:14:32Z" }
status: stable
---

Authoritative execution order for the 11 V2 capabilities. All 11 specs already exist in
`openspec/specs/`; each capability is then implemented as one OpenSpec change, in this order.
A change is archived before the next one starts. The UI stack (templ + Tailwind + Alpine + htmx)
is a cross-cutting foundation change against the `config` capability, tracked as row 1a.

| # | capability | change id | depends on | state |
|---|---|---|---|---|
| 1 | config | `add-config` | — | archived |
| 1a | UI stack (config page port + foundation) | `add-ui-stack` | config | archived |
| 2 | store | `add-store` | config | archived |
| 3 | plugin-host | `add-plugin-host` | store | archived |
| 4 | host-tools | `add-host-tools` | plugin-host | archived |
| 5 | library | `add-library` | host-tools, store | archived |
| 6 | reader | `add-reader` | library | archived |
| 7 | image-cache | `add-image-cache` | reader | archived |
| 8 | detail | `add-detail` | library, image-cache | active |
| 9 | updates | `add-updates` | store, plugin-host | pending |
| 10 | browser | `add-browser` | plugin-host | pending |
| 11 | sites | `add-sites` | plugin-host, host-tools | pending |
