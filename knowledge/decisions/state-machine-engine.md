---
type: decision
title: Use qmuntal/stateless as the FSM engine behind typed domain wrappers
description: "User ruling 2026-11-21: qmuntal/stateless v1.8.0 for plugin and job machines instead of a hand-written engine, with typed enums, single owner goroutine, SQLite external state (persist before publish) and graph golden tests; cache entries and daemon lifecycle excluded."
tags: [decision, fsm, stateless, dependency, lifecycle]
generated: { by: agent/mcp, at: "2026-11-21T12:00:12Z" }
status: stable
---

## User ruling (2026-11-21)

Use `github.com/qmuntal/stateless` v1.8.0 as the FSM engine instead of hand-writing one. Trigger for the
ruling: "hand coded FSM might hurt when we need a feature mid-way". That concern is correct -- what gets
expensive is the engine (guards, hierarchy, entry/exit, persistence hooks, run-to-completion, graph export),
not the domain transition table, which every option requires anyway.

## Decisions

- **D1** -- `qmuntal/stateless` v1.8.0 (BSD-2, zero transitive deps, Go >= 1.24) is the engine for the plugin
  and job machines. It is a port of `.NET Stateless`; the feature list covers the likely future asks.
- **D2** -- Domain stays typed: `PluginState`, `PluginTrigger`, `JobState`, `JobTrigger` enums; the library's
  `any`-typed `State`/`Trigger` never appear outside `internal/fsm` wrappers.
- **D3** -- One owner goroutine per manager fires every trigger, ordered. Library thread-safety is not relied
  on for ordering.
- **D4** -- External state storage maps to SQLite: the mutator writes `state` + `state_version` inside the
  transition; events are published from `OnTransitioned` afterwards (persist before publish).
- **D5** -- `FiringQueued` only; callbacks are cheap (enqueue/persist/publish), no HTTP or scraping.
- **D6** -- The cache entry is NOT an engine machine: typed enum + singleflight reservation, because there can
  be thousands of entries and two transitions. Daemon boot/shutdown is an ordered slice.
- **D7** -- Timeouts are owner-generated triggers (`TrigTimeout`), not library timers.

## Alternatives and why rejected

| Alternative | Verdict |
|---|---|
| Hand-rolled typed table + engine (~100 LOC) | Cheapest today and fully typed, but guards, hierarchical states, entry/exit ordering, external persistence, queueing and DOT export would all be rebuilt on demand -- exactly the risk that motivated this decision. Kept as fallback only if stateless proves a poor fit. |
| `looplab/fsm` | 3.4k stars, active, Apache-2.0, but stringly typed: `NewFSM(initial string, ...)`, `Event(ctx, "trigger")`, callbacks keyed by strings. Typos surface at runtime; no hierarchy. |
| `mustafaakin/event-bus` style micro-machine | No maintained option exists; writing one is the hand-rolled path. |
| Informal flags + switches | Leaves illegal transitions unexpressed and untestable; boot reconciliation ad hoc. |

## Consequences

- One more dependency (no transitive deps). It stays behind `internal/fsm`.
- `State`/`Trigger` are `any`: the wrapper and the graph golden test are the compile/runtime guard rails.
- New persisted columns/table (`plugin.state*`, `jobs`) are divergences from cppIsekai and need an OpenSpec
  MODIFIED delta; the public API contract (boolean `enabled`) is unchanged.
- Transition model and rules: [state-machines](../architecture/state-machines.md).
- Review trigger: if the DOT graph of any machine exceeds ~8 states, split the entity instead of growing
  the graph.
