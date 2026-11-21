---
type: decision
title: Adopt gookit/event behind a single-actor bus; separate stream plane
description: "User ruling 2026-11-21: gookit/event v1.2.0 as dispatch engine, owned by one actor goroutine, async API banned, typed payload facade, drop-oldest stream hub delivered over WebSocket GET /ws/log; hand-written hub kept as documented fallback."
tags: [decision, events, gookit, dependency, concurrency]
generated: { by: agent/cli, at: "2026-11-21T11:46:11Z" }
status: stable
---

## User ruling (2026-11-21)

Use `github.com/gookit/event` (v1.2.0, MIT) as the dispatch engine for goIsekai2, wrapped by a
single-goroutine actor bus. Chosen over a hand-written hub; the second candidate the user named
(`mustafaakin/event-bus`) does not exist (GitHub API 404 on 2026-11-21, no such repo in the account).
On 2026-11-21 the user asked to re-evaluate `asaskevich/EventBus`; it was rejected on source review
(see alternatives).

## Decisions

- **D1** -- `gookit/event` v1.2.0 is the only third-party event dependency; it stays behind
  `internal/events` and never appears in a domain package signature.
- **D2** -- One actor goroutine owns the `*event.Manager`. No other goroutine calls `Fire/FireEvent/On/
  RemoveListener`. Reason: the manager is not safe for concurrent fire + registration (`ListenerQueue.Sort`
  sorts in place, `listeners` map is mutated by `On`), and `EnableLock` serializes dispatch with a
  non-reentrant mutex that deadlocks nested publishes.
- **D3** -- The async API (`Async`, `FireC`, `FireAsync`) is banned: one global 100-slot channel shared by
  3 consumers, blocking the publisher when full, `Close()` panics on later sends, and no per-subscriber policy.
- **D4** -- Streaming (log lines, job progress, UI refresh) uses a separate `internal/stream` fan-out with
  per-client buffers and drop-oldest. Domain events never drop.
- **D5** -- Typed facade: payload is one struct under key `payload`; `Publish[T]`/`Listen[T]` assert once.
  Event names are constants, never literals at call sites.
- **D6** -- Commands are not events. Prefetch/sync/warm triggers stay explicit interface calls with errors
  and backpressure; the bus only carries facts that already happened (published after commit).
- **D7** -- Plugins never receive a bus handle. Guest code keeps the host-tool surface only.

## Alternatives and why rejected

| Alternative | Verdict |
|---|---|
| Hand-written typed hub (~100 LOC, `Topic[T]`) | Strictly less code and compile-time safe, but the user chose gookit for its priority/wildcard/once machinery. Kept as the documented fallback (see review trigger). |
| `mustafaakin/event-bus` | Does not exist -- 404 verified via GitHub API and the account's repo list. |
| `asaskevich/EventBus` | Rejected on source review (2026-11-21). Simple (216 LOC, thread-safe registry) but: handlers are reflection varargs (`Publish(topic, args...)` -> `callBack.Call`), so a payload mismatch is a runtime panic and no `recover()` exists anywhere in the file -- a panic inside an async handler kills the process; `Publish` holds the bus lock for the whole handler loop, so nested publish deadlocks and every subscriber serializes; async mode spawns one goroutine per handler per publish (unbounded); `Publish` returns no error, so listener failures are invisible; no priorities or wildcards; last upstream push 2024-06. |
| `ThreeDotsLabs/watermill` | Message router for distributed systems; far too heavy for an in-process daemon. |
| No bus at all | Leaves cppIsekai's mix of direct calls, worker nudges and polling in place; the point of the change is to separate facts, commands and streams. |

## Consequences

- Every capability change that adds an event must extend the catalog in
  [event-driven-core](../architecture/event-driven-core.md), publish after commit, and add a bus test.
- `internal/events` is the only package importing `gookit/event`; `internal/stream` does not import it at all.
- Review trigger: if priority/wildcard/`Once`/`Abort` are still unused after three capability changes,
  replace gookit with the typed hub; only the wrapper changes.
