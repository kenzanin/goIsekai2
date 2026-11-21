---
type: decision
title: goIsekai2 is a greenfield Go rewrite against the cppIsekai baseline
description: "Decisions: cppIsekai is the read-only behavioral baseline (its knowledge bundle + openspec specs), goIsekai V1 is a read-only scraping corpus, bundles never merge, module goisekai2."
tags: [go, rewrite, bootstrap, reference, goisekai]
generated: { by: agent/mcp, at: "2026-11-21T11:32:17Z" }
status: stable
---

## Context

The first Go generation lives at `~/tmp/goIsekai`: a working harness with ~40 Lua and ~40 JS site
plugins, wasm/native/yaegi runtime trials, HAR captures and a full daemon. It proved the concept but
accumulated stack churn. The finished behavioral baseline is the C++ implementation at
`~/tmp/cpp/cppIsekai`.

## Decision

goIsekai2 starts empty and is rebuilt in Go against the cppIsekai baseline:

1. cppIsekai is the reference, not a dependency. Its `knowledge/` bundle and `openspec/specs/` define
   the target semantics (plugin ABI, cache layout, config knobs, status vocabularies, routes).
2. V1 `~/tmp/goIsekai` is a read-only corpus for porting material: plugin sources
   (`examples/plugins/<runtime>/<site>/`) and real traffic (`tmp/*.har`, `tmp/har/`).
3. Bundles are never merged: this repository keeps its own `knowledge/` and `openspec/`.
4. Module path is `goisekai2`; binary is `bin/goisekai2`. Rename only with a new decision.

## Rationale

- Porting from a finished reference keeps feature parity checkable, while porting from V1 keeps the
  scraping knowledge (selectors, endpoints, auth quirks) without inheriting V1 structure.
- One capability spec per cppIsekai layer gives each rewrite step a testable contract before code lands.

## Consequences

- Every capability must cite the cppIsekai baseline it ports (design decision Dn).
- Divergence (different dependency, different layout, dropped feature) requires its own decision
  concept plus a MODIFIED OpenSpec delta.
- Reference map: [reference-sources](../references/reference-sources.md).
