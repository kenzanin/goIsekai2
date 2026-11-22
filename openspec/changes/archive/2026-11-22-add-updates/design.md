# Design: add-updates

## Context

See `proposal.md`. The `updates` capability spec freezes tick, candidates, backoff, pacing and
maintenance. cppIsekai `update/scheduler.cpp` is the source of truth.

## Goals / Non-Goals

- Goals: single owner goroutine, config re-read, exact backoff curve, chapter-only updates,
  independent maintenance steps, clean shutdown.
- Non-Goals: new config knobs, UI, site plugins.

## Decisions

### D1 — Scheduler actor

`internal/updates.Scheduler` owns one goroutine with ticker, context cancellation and panic recovery.
No other goroutine touches its state. Baseline: cppIsekai single maintenance pool.

### D2 — Candidate query

Selection uses typed store methods over `saved_manga` + `plugin` + `update_attempts`: `read_ch <= 1`,
enabled plugin, backoff reached, ordered never-attempted first then oldest attempt, capped per plugin.
Baseline: `Store::update_candidates`.

### D3 — Chapter-only refresh

Call `pages`? No — call only `chapters`. Validate, write chapter cache, refresh `total_ch`, set bit0 on
growth and preserve all other bits. Baseline: cppIsekai scheduled update scope.

### D4 — Backoff

Success clears `consecutive_errors` and `backoff_until`; failure increments and sets
`now + min(2^n, 72)h`. Manual refresh writes the same row but ignores backoff.
Baseline: `upsert_update_attempt`.

### D5 — Pacing and watermark

Sleep `update_request_delay_ms` between titles, never after the last; stamp `update/last_run` and log
`checked=.. ok=.. error=..` with `slog`.

### D6 — Maintenance

Each tick prunes expired `api_cache`, rotates `VACUUM INTO` backups with `backup_keep`, and reclusters
duplicates once when dirty. Failures are logged and isolated.

## Risks / Trade-offs

- [double scheduling] → one actor plus manual stamp freshness.
- [backup disk use] → cap by `backup_keep`.
- [panic in plugin] → recover in tick loop, log and continue.

## Migration Plan

Greenfield. Rollback is reverting the change commit.

## Open Questions

None. Behavior is fixed by the capability spec.
