# Proposal: add-updates

## Why

`detail` is archived, so manual refresh already stamps attempts. The update loop is next in the
porting order and closes the gap between manual refresh and automatic chapter maintenance.

## What Changes

- Add one maintenance goroutine with 60-second tick, re-read knobs each tick and jittered due check.
- Select barely-read, enabled, non-backed-off titles with a per-plugin cap and rotation.
- Call only the plugin's `chapters` method, refresh `total_ch`, and set bit0 on growth while preserving bit6.
- Track attempts with exponential backoff capped at 72 hours; manual refresh stays fresh and bypasses backoff.
- Pace requests, write the `update/last_run` watermark and log a run summary.
- Prune expired API cache rows, take rotated VACUUM snapshots and recluster duplicates on the same tick.

## Capabilities

### New Capabilities

- none

### Modified Capabilities

- `updates` — adds daemon wiring (single goroutine with clean shutdown and config re-read) on top of the frozen scheduler contracts.

## Impact

- Code: new `internal/updates` package plus boot/shutdown wiring in `cmd/goisekai2`.
- Dependencies: none beyond the archived stack.
- No cgo: `CGO_ENABLED=0 go build ./...` stays the gate.
- Out of scope: site plugins, UI pages, scheduler config knobs (already declared).

## Baseline

cppIsekai `update/scheduler.cpp`, `Store::update_candidates` and `upsert_update_attempt`.
