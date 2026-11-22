---
type: decision
title: Browser host transport and supervision in Go
description: Keep the baseline browser-host tool surface; supervise lightpanda with os/exec and hide the transport behind one Browser interface.
tags: [browser, lightpanda, supervision, plugin, divergence]
generated: { by: agent/mcp, at: "2026-11-22T04:36:55Z" }
status: stable
---

# Related Concepts
- [Architecture overview: layers, import rules and frozen external contracts](../architecture/overview.md): browser backend and its supervision are part of the plugin layer described in the overview

## Ruling 2026-11-22: use go-rod/rod

The user replaced the hand-written MCP JSON-RPC client with github.com/go-rod/rod v0.116.2.
Verified: MIT, pure Go (no cgo), and rod.New().ControlURL(ws://127.0.0.1:<port>/).Connect()
succeeds against lightpanda serve (CDP /json/version reports webSocketDebuggerUrl).
The supervisor now spawns lightpanda serve instead of lightpanda mcp. The Browser interface
and typed errors stay ours; CDP transport, page mechanics and waiting helpers come from rod.
The hand-written MCP client was deleted. Reason: prefer a maintained library over bespoke
protocol code. The rod Eval helper expects a function expression, so plugin evaluate scripts
are wrapped into function(){ return (...) }.
