---
type: Decision
title: First site corpus tranche and parity tracking
description: Port seven baseline plugin folders first and record every other baseline folder as an omission with a reason rather than silently dropping it.
tags: [plugins, sites, parity]
generated: { by: agent/mcp, at: "2026-11-22T05:33:09Z" }
status: stable
---

Decision: the first port tranche contains top-5 manga plugins by cppIsekai saved_manga count (mangafire, atsumaru, 1manga, mangatown, mangahub), the mangadex reference plugin, and two enrichment sources. 1manga is deferred because its Lua modules need a require shim. Every baseline folder is listed in plugins/PARITY.md as ported or omitted with an explicit reason. Corpus shrinkage without a recorded reason is a defect per the sites spec. Verified against data/library.db counts: mangafire 39, atsumaru 15, 1manga 13, mangatown 7, mangahub 7.

# Related Concepts
- [Capability map: 25 baseline specs collapsed into 11 V2 capabilities with a porting order](../architecture/capability-map.md): Site corpus tranche follows the mapped capability order
