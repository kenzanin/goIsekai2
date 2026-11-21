---
type: reference
title: "Reference source map (cppIsekai baseline, goIsekai V1 corpus)"
description: "Exact read-only paths for porting: cppIsekai knowledge/openspec contracts and goIsekai V1 plugin examples, HAR captures, stack trials."
tags: [reference, porting, cppisekai, goisekai, paths]
generated: { by: agent/mcp, at: "2026-11-21T11:32:17Z" }
status: stable
---

## Purpose

Single entry point for "where do I look for X" while rewriting. Everything below is read-only.

## cppIsekai baseline (`~/tmp/cpp/cppIsekai`) -- behavioral source of truth

| Need | Path |
|---|---|
| Fresh end-to-end survey (counts, request flow) | `knowledge/architecture/verified-layer-map.md` |
| Layer / call-path rules | `knowledge/references/code-layers.md` |
| Capability contracts to port | `openspec/specs/<capability>/spec.md` |
| Plugin ABI + runtimes | `knowledge/architecture/plugin-system.md`, `knowledge/decisions/plugin-runtime*.md` |
| Cache layout | `knowledge/architecture/cache-layout-v3.md` |
| Config knobs | `knowledge/architecture/cli-config.md` |
| Tooling and test gates | `knowledge/references/tooling.md`, `AGENTS.md` |
| Porting backlog | `knowledge/references/plugin-porting-matrix-goIsekai.md` |
| Change artifact format | `knowledge/references/openspec-delta-spec-format.md` |

## goIsekai V1 corpus (`~/tmp/goIsekai`) -- scraping material

| Need | Path |
|---|---|
| Lua plugin examples | `examples/plugins/lua/<site>/main.lua` |
| JS plugin examples | `examples/plugins/js/<site>/` |
| wasm / native / yaegi experiments | `examples/plugins/{wasm,native,yaegi}/` |
| Real HTTP traffic | `tmp/*.har`, `tmp/har/` |
| Go stack trials (chi, jet, tls-client, sled) | `go.mod`, `internal/` |

## Rules

- Never fuse these bundles into this one; they are read-only external references.
- Cite exact paths in findings; do not paraphrase reference behavior from memory.
- Parent decision: [goisekai-v2-bootstrap](../decisions/goisekai-v2-bootstrap.md).
