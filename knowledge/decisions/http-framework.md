---
type: decision
title: chi + coder/websocket for HTTP and WebSocket
description: "Use go-chi v5 as the HTTP router and coder/websocket for the log stream, keeping handlers on net/http so middleware and SSR stay stdlib-compatible."
tags: [http, router, websocket, chi, stream, stack]
generated: { by: agent/cli, at: "2026-11-21T12:35:23Z" }
status: stable
---

# Related Concepts
- [Stack: libraries vs hand-written code](../architecture/stack.md): names the router and websocket libraries on the stack page
