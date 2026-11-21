---
type: decision
title: Browser host transport and supervision in Go
description: Keep the baseline browser-host tool surface; supervise lightpanda with os/exec and hide the transport behind one Browser interface.
tags: [browser, lightpanda, supervision, plugin, divergence]
generated: { by: agent/cli, at: "2026-11-21T12:29:11Z" }
status: stable
---

# Related Concepts
- [Architecture overview: layers, import rules and frozen external contracts](../architecture/overview.md): browser backend and its supervision are part of the plugin layer described in the overview
