# Design: add-browser

## Context

See `proposal.md`. The `browser` spec freezes the call contract, session rules and supervision. The
transport decision records `os/exec` + MCP instead of the baseline `reproc`.

## Goals / Non-Goals

- Goals: one plugin tool, per-session serialization, one expiry retry, one respawn, clean shutdown, loopback proxy bypass, referer sentinel.
- Non-Goals: CDP adapter, new config knobs, UI.

## Decisions

### D1 — Interface

`internal/browser.Browser` exposes `Call(tool string, args any) (string, error)` and `Close()`. The MCP client is the default backend; a CDP adapter can replace it without touching plugins.

### D2 — MCP client

JSON-RPC `tools/call` over HTTP to the configured URL, bounded by config timeout. Loopback URLs always use a transport with `Proxy: nil`.

### D3 — Session recovery

The client caches a session id and retries once on a session-expired response. A second expiry returns a typed error. Per-session mutexes serialize complete sequences only.

### D4 — Supervisor

At boot, probe health; reuse a foreign healthy instance. Otherwise spawn `lightpanda mcp --host 127.0.0.1 --port <port>`, wait for health, and record ownership. On child death, respawn once; on further death mark dead. Shutdown stops only owned children, SIGTERM then SIGKILL after a grace period.

### D5 — Capability gate

Host tool dispatch checks the caller manifest's browser capability before the backend is touched. Unknown tool is the failure shape.

### D6 — Referer sentinel

`Manifest.ImageReferer` defaults empty (origin from website); `"none"` means no `Referer` header; any other value is sent verbatim.

## Risks / Trade-offs

- [child lifecycle] → ownership flag prevents killing a foreign instance.
- [session races] → per-session mutex, not global lock.
- [retry storms] → one expiry retry and one respawn only.

## Migration Plan

Greenfield. Rollback is reverting the change commit.

## Open Questions

None. Behavior is fixed by the capability spec and transport decision.
