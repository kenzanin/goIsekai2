---
type: Process
title: "Capability porting order: one OpenSpec change per capability, config first"
description: "Authoritative dependency-ordered sequence for implementing the 11 V2 capabilities as OpenSpec changes, with per-change rules, gates and a status column that is updated when a change archives."
tags: [process, openspec, porting, sequencing, status]
generated: { by: agent/mcp, at: "2026-11-22T05:35:24Z" }
status: stable
---

Authoritative execution order for the 11 baseline-derived V2 capabilities plus 2 V2-only additions. All specs already exist in
`openspec/specs/`; each capability is then implemented as one OpenSpec change, in this order.
A change is archived before the next one starts. The UI stack (templ + Tailwind + Alpine + htmx)
is a cross-cutting foundation change against the `config` capability, tracked as row 1a.

| # | capability | change id | depends on | state |
|---|---|---|---|---|
| 1 | config | `add-config` | — | archived |
| 1a | UI stack (config page port + foundation) | `add-ui-stack` | config | archived |
| 2 | store | `add-store` | config | archived |
| 3 | plugin-host | `add-plugin-host` | store | archived |
| 3a | Lua 5.5 runtime (modernc.org/lua) | `add-lua55` | plugin-host | archived |
| 4 | host-tools | `add-host-tools` | plugin-host | archived |
| 4a | host-tools baseline name parity | `host-tools-parity` | host-tools | archived |
| 5 | library | `add-library` | host-tools, store | archived |
| 6 | reader | `add-reader` | library | archived |
| 6a | reader fullscreen/controls parity | `add-reader-parity` | reader | archived |
| 6b | reader default paged fit-width and single overlay bar | `reader-defaults` | reader | archived |
| 7 | image-cache | `add-image-cache` | reader | archived |
| 7a | cppIsekai cache-layout concept alignment | `cache-layout-concept` | image-cache, library, store | archived |
| 8 | detail | `add-detail` | library, image-cache | archived |
| 8a | detail pagination and enrichment cache | `detail-pagination` | detail | archived |
| 8b | cache-only detail read path | `cache-only-read` | detail | archived |
| 8c | staged search JSON to master flow | `search-to-master` | detail, library | archived |
| 8d | TEMP staging and search response cache | `temp-stage-search-cache` | detail, library, store | archived |
| 8e | saved toggle highlight and balanced actions | `detail-toggle-highlight` | detail | archived |
| 8f | newest-first chapters and raw cover save | `chapter-sort-cover-fix` | detail, library | archived |
| 9 | updates | `add-updates` | store, plugin-host | archived |
| 10 | browser | `add-browser` | plugin-host | archived |
| 11 | sites | `add-sites` | plugin-host, host-tools | archived |
| 11a | atsumaru newest duplicate release | `atsumaru-newest-release` | sites | archived |
| 12 | observability (V2-only) | `add-behavior-specs` | sites | archived |
| 13 | ui-shell (V2-only) | `add-behavior-specs` | sites | archived |
| 13a | in-place navigation and detail polish | `ui-inplace-polish` | ui-shell, library, detail | archived |
