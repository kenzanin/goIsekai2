---
type: decision
title: Plugin io shim is read-only and confined to the plugin directory
description: "V2 exposes io.read and io.lines to plugins, but unlike cppIsekai the paths are confined to the calling plugin directory; absolute paths and .. escapes fail with typed errors."
tags: [decision, plugins, sandbox, io, host-tools]
generated: { by: agent/mcp, at: "2026-11-22T17:33:25Z" }
status: stable
---

## Decision

Lua and JS plugin environments expose only `io.read` and `io.lines`, both read-only and resolved through the
same confinement as `host.fs.read`. A relative path resolves inside the plugin's own directory; absolute paths,
`..` escapes and symlink-equivalent traversal fail with a typed path error before any file is opened.

## Why

cppIsekai's `io.lines(path)` accepts arbitrary host paths. V2's no-ambient-authority rule forbids that: a plugin
may read its own assets and bundled modules, but never the daemon config, another plugin's secrets or the host
filesystem. The capability is part of baseline host-tool parity; the confinement is the deliberate divergence.

## Consequences

- Plugins that need a shared library read through `host.fs.read_lib`, which is confined to `<pluginDir>/lib`.
- The host still offers `host.http.fetch` for network data and `host.storage.*` for state; neither becomes a file escape.
- Tests must prove an in-directory read succeeds and that an escape fails typed in both Lua and JS.

# Related Concepts
- [tools fs: Read() sandboxes plugin file reads inside the plugin dir](../notes/plugin-tools-fs-read.md): io shim reuses the same confinement as fs.Read.
- [Architecture overview: layers, import rules and frozen external contracts](../architecture/overview.md): Records the plugin filesystem authority boundary.
