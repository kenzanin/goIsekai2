---
type: Fact
title: Projection and MergeSave verified for internal/config
description: Verified Projection and MergeSave implementation and behavior for internal/config.
tags: [internal/config, projection, mergesave]
generated: { by: agent/mcp, at: "2026-11-21T15:16:21Z" }
status: stable
---

Verified that Projection resolves configuration knobs by re-reading the INI file without writes, falling back to defaults on invalid entries while marking only the offending knob with Invalid=true, Source='default', and Value=Default. Verified that MergeSave validates keys/values, creates missing INI files from registry dump templates, updates existing INI files while preserving comments, and rolls back atomically on validation errors. Tests pass under CGO_ENABLED=0 go test ./internal/config.

# Related Concepts
- [Config resolution gotchas found while building add-config](config-resolution-gotchas.md): Projection and MergeSave build upon configuration resolution rules.
