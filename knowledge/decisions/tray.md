---
type: decision
title: Tray as a separate pure-Go command on fyne.io/systray
description: "fyne.io/systray v1.11.0 in its own tray binary: pure Go on Linux (dbus StatusNotifierItem) and Windows, cgo only on macOS; the daemon itself stays pure Go and is spawned as a child."
tags: [tray, systray, dbus, purego, no-cgo, stack]
generated: { by: agent/cli, at: "2026-11-21T12:43:39Z" }
status: stable
---

## Context

`architecture/stack` listed the tray as the one place where every library seemed to demand cgo. V1
(`~/tmp/goIsekai`) already ships a cgo-free tray, so the slot is not open: it is solved, and the evidence is
a build, not a claim. Verified on 2026-11-21:

```bash
cd ~/tmp/goIsekai && CGO_ENABLED=0 go build -o /tmp/gk-tray ./cmd/goisekai-tray   # exit 0
```

## Decision

| Slot | Choice | Evidence |
|---|---|---|
| Tray | `fyne.io/systray` v1.11.0 | Linux/BSD: pure Go, `godbus/dbus/v5` StatusNotifierItem + DbusMenu; Windows: pure Go syscalls; macOS: cgo + Objective-C (`systray_darwin.m`) |
| Shape | separate command (`cmd/goisekai-tray`) that spawns the daemon as a child | V1 `cmd/goisekai-tray/main.go`: Open Browser / Restart / Quit, child stdout/stderr to a log file |
| Config | `[tray]` keys in the ini (`server_bin`, `client_bin`, `url`, `log_file`, icon path) | a deployment moves paths without a rebuild |

Two consequences matter more than the library choice:

1. **The daemon never links a tray.** The tray is its own binary, so the core stays `CGO_ENABLED=0` on every
   platform, and a platform without a tray still has a working daemon plus web UI.
2. **macOS is the exception.** `systray_darwin.m` is cgo/ARC. A macOS tray build may enable cgo for that one
   command, or be skipped; it never contaminates `internal/*` or `cmd/goisekai2`.

## purego is not what makes the tray cgo-free

V1 does depend on `github.com/ebitengine/purego` v0.10.1, but it is used by
`internal/pluginmanager/goplug_native*.go` to `dlopen` native `.so` plugins while the host keeps
`CGO_ENABLED=0` -- not by the tray. The tray's cgo-freedom comes from Linux tray icons being a dbus protocol,
not a linked library.

That finding is recorded for the future: if goIsekai2 ever adopts native `.so` plugin runtimes, purego is the
proven mechanism (V1 maps a private copy of the library, then resolves symbols). It is not needed for the
tray and it is not needed for the daemon.

## Alternatives rejected

| Option | Why not |
|---|---|
| In-process tray inside the daemon | pulls platform UI into the daemon and breaks the "daemon is a web server" boundary |
| `energye/systray`, `getlantern/systray` | same cgo surface as fyne on macOS, older maintenance, no advantage |
| purego tray shim per OS | hand-written ABI per platform for something fyne already does in pure Go on Linux/Windows |
| No tray at all | V1 users already have one; the split-command shape costs nothing and keeps parity |

## Consequences

1. The tray is a thin supervisor: spawn, supervise, restart, quit. No business logic, no DB access, no
   plugin calls.
2. Single-instance locking belongs to the daemon, not to the tray (two trays must not mean two daemons).
3. Tray tests cover process supervision and ini parsing; the icon itself is a dbus integration and is
   verified by hand.

# Related Concepts
- [Stack: libraries vs hand-written code](../architecture/stack.md): resolves the tray slot listed as open on the stack page
- [Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, quickjs, wazero)](dependency-stack.md): purego is kept there for native plugin loading, not for the tray
