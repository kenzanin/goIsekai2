---
type: decision
title: chi + coder/websocket for HTTP and WebSocket
description: "Use go-chi v5 as the HTTP router and coder/websocket for the log stream, keeping handlers on net/http so middleware and SSR stay stdlib-compatible."
tags: [http, router, websocket, chi, stream, stack]
generated: { by: agent/cli, at: "2026-11-21T12:35:23Z" }
status: stable
---

## Context

The daemon needs more than a router: route groups, middleware, streaming image responses, SSR pages with
htmx fragments, plus a real WebSocket endpoint for the log console. The baseline already streams logs over
WebSocket `GET /ws/log` (baseline `src/http/http_pages.cpp:350`: replays a 200-line ring snapshot on open,
then plain text frames into a `<pre>`), and `architecture/overview` left realtime transport open between SSE
and WS.

User ruling 2026-11-21: use a web framework library rather than `net/http` alone, because WebSocket support
is required.

## Decision

| Slot | Choice | Why |
|---|---|---|
| Router / framework | `github.com/go-chi/chi/v5` v5.5.0 (MIT) | `http.Handler`-native, zero transitive deps, route groups + middleware, `chi/middleware` (Recoverer, RequestID, RealIP, Logger); streaming stays `http.ResponseWriter` |
| WebSocket | `github.com/coder/websocket` v1.8.15 (ISC) | context-aware API on `net/http`, single-writer discipline handled internally, ping/pong built in, zero transitive deps |
| Templates / SSR | stdlib `html/template` + htmx fragments | unchanged |
| Static and image responses | stdlib `http.FileServer` / `http.ServeContent` behind chi routes | unchanged |

Transport stays WebSocket, not SSE: `GET /ws/log` with snapshot replay is part of the frozen route surface,
so parity wins. The stream hub in `internal/stream` stays transport-agnostic; the WS handler is one
subscriber among others.

Handler signature stays `func(http.ResponseWriter, *http.Request)` everywhere. Framework-specific contexts
are forbidden, so the framework can be swapped without touching handlers, and every handler is testable
through `httptest`.

## Alternatives rejected

| Option | Why not |
|---|---|
| stdlib `ServeMux` 1.22 patterns | no middleware chain, no route groups; the user prefers a library, and WS still needs a second dependency |
| `gin` | own `gin.Context` handler signature locks every handler to the framework |
| `echo` | same `echo.Context` lock-in, larger surface than needed |
| `fiber` | `fasthttp`, not `net/http`: loses `http.Handler` middleware, `http.ResponseWriter` streaming semantics and stdlib tooling |
| `gorilla/websocket` | viable and maintained, but one-concurrent-writer discipline is manual and the API is older |
| SSE only | baseline route is WS and its log console opens a `WebSocket`; would also need a second endpoint for bidirectional work later |

## Consequences

1. Two dependencies, both zero-transitive and pure Go; `CGO_ENABLED=0 go build ./...` unaffected.
2. `chi/middleware.Recoverer` guarantees no panic escapes a request, but domain 5xx handling stays in
   handlers -- a recover is a bug report, not error handling.
3. WS frames carry text for the log stream and JSON envelopes for progress and grid refresh.
4. Version pinning: chi and coder/websocket are upgraded deliberately, never floated.

# Related Concepts
- [Stack: libraries vs hand-written code](../architecture/stack.md): names the router and websocket libraries on the stack page
