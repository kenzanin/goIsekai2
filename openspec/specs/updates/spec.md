# updates Specification

## Purpose

Defines the background maintenance loop: auto-update of saved titles, attempt tracking with
exponential backoff, per-tick maintenance (API cache prune, database backup, duplicate recluster), and
the rules that keep a manual refresh and the scheduler from fighting. All facts below are grounded in
the baseline scheduler and store code, not in its prose.

Knobs (config registry, re-read every tick so edits apply without restart): `update_interval_days`,
`update_per_plugin_limit`, `update_request_delay_ms`, `backup_interval_h`, `backup_keep`.

## Requirements

### Requirement: Tick loop

The daemon SHALL run one maintenance goroutine that wakes every 60 seconds, re-reads the knobs, and
SHALL apply a per-process random jitter of up to 15 minutes to the auto-update due check so two
instances do not fire together. The goroutine SHALL stop on shutdown within one tick and SHALL never
propagate a panic into the daemon.

#### Scenario: Not due yet
- **WHEN** the stored `update/last_run` watermark is newer than `interval - jitter`
- **THEN** the tick returns without selecting candidates

#### Scenario: Shutdown
- **WHEN** the daemon shuts down
- **THEN** the maintenance goroutine exits within one tick and no work starts afterwards

### Requirement: Candidate selection

A tick that is due SHALL select candidates from saved entries where `read_ch <= 1`, the owning plugin
is enabled, and `backoff_until` is absent or reached, taking at most
`update_per_plugin_limit` per plugin ordered never-attempted first then oldest attempt.

#### Scenario: Barely-read only
- **WHEN** a saved title has `read_ch = 2`
- **THEN** it is never selected by the scheduler

#### Scenario: Backed-off title skipped
- **WHEN** a title's `backoff_until` is in the future
- **THEN** it is excluded from the batch

#### Scenario: Disabled plugin skipped
- **WHEN** a plugin is disabled
- **THEN** none of its titles are selected, and the manual path still answers `plugin disabled`

#### Scenario: Per-plugin cap with rotation
- **WHEN** a plugin has 30 eligible titles and the cap is 5
- **THEN** exactly 5 are selected, the never-attempted first, then the oldest attempts

### Requirement: Update scope is chapters only

A scheduled update SHALL call only the plugin's chapters method, SHALL write the chapters cache entry
and refresh the stored stats (`total_ch`), and SHALL set bit0 of `status_flags` when the chapter count
grew, preserving bit6 and every other bit. It SHALL NOT fetch detail.

#### Scenario: Growth marks New
- **WHEN** a stored title at 40 chapters reports 42
- **THEN** `total_ch` becomes 42 and bit0 is set, with the dim bit preserved

#### Scenario: No growth, no New
- **WHEN** the chapter count is unchanged
- **THEN** stored stats keep their values and no bit changes

#### Scenario: Detail is not fetched
- **WHEN** a scheduled update runs
- **THEN** the plugin receives exactly one call, the chapters method

### Requirement: Attempt tracking and backoff

Every completed attempt SHALL upsert `update_attempts(plugin, slug)` with `last_attempt_at`,
`outcome` (`ok` | `error`), `error`, `consecutive_errors` and `backoff_until`. Success SHALL reset the
counter and clear the backoff; failure SHALL increment the counter and set
`backoff_until = now + min(2^n, 72) hours`, so 2, 4, 8, 16, 32, 64 and then 72 hours.

#### Scenario: Failure backs off
- **WHEN** a title fails for the third consecutive time
- **THEN** the row records `consecutive_errors = 3` and `backoff_until = now + 8h`

#### Scenario: Success clears backoff
- **WHEN** a backed-off title later succeeds
- **THEN** `consecutive_errors = 0` and `backoff_until = 0`

#### Scenario: Cap
- **WHEN** a title has failed ten times
- **THEN** its backoff is 72 hours, not unbounded

#### Scenario: Rejected pairs are not attempts
- **WHEN** the plugin or slug does not resolve, or the plugin is disabled
- **THEN** no attempt row is written and no upstream call is made

### Requirement: Pacing

The scheduler SHALL sleep `update_request_delay_ms` between titles and SHALL NOT sleep after the last
title of the batch. Batch completion SHALL write the `update/last_run` watermark and log
`checked=<n> ok=<n> error=<n>`.

#### Scenario: Delay between, not after
- **WHEN** a batch of 3 titles runs with a 500 ms delay
- **THEN** two 500 ms pauses occur and the run ends immediately after the third

#### Scenario: Watermark only after the batch
- **WHEN** the batch completes
- **THEN** `update/last_run` is stamped and the summary line is logged

### Requirement: Manual refresh counts as an attempt

A manual `POST /api/refresh/:plugin/:slug` SHALL stamp the same attempt rows (`ok` or `error`), so the
scheduler treats the title as freshly attempted and skips it until its interval elapses and its
backoff clears. A manual path SHALL ignore an existing backoff.

#### Scenario: Manual then scheduler
- **WHEN** the user refreshes a title manually and the scheduler ticks immediately after
- **THEN** the title is not selected again

#### Scenario: Manual bypasses backoff
- **WHEN** a title is backed off and the user refreshes it manually
- **THEN** the plugin is called anyway and the attempt row is updated

### Requirement: Maintenance on the same tick

The same tick SHALL also: prune expired `api_cache` rows, take a VACUUM snapshot into
`<db dir>/backups` when `backup_interval_h` has elapsed keeping `backup_keep` files with its own
`backup/last_run` watermark, and recompute duplicate groups when browsing marked the dedupe cache
dirty. A failure in any of these SHALL be logged and retried on the next tick without aborting the
update batch.

#### Scenario: Backup cadence
- **WHEN** `backup_interval_h = 24` and the last backup is 25 hours old
- **THEN** expired cache rows are pruned, a snapshot is written, and the backup watermark updates

#### Scenario: Backup failure retries
- **WHEN** the snapshot fails
- **THEN** the watermark is unchanged, the error is logged, and the update batch still runs

#### Scenario: One recluster per change
- **WHEN** many detail opens marked the dedupe cache dirty and the tick runs
- **THEN** duplicate groups are recomputed exactly once for that tick

### Requirement: Observability

Every run SHALL log its batch summary and every failure SHALL carry plugin, slug and the upstream
error text, all through the log pipeline that feeds the log stream. The scheduler state SHALL be
visible without a restart.

#### Scenario: Run summary visible
- **WHEN** a run completes
- **THEN** its summary appears in the log stream tagged as auto-update

#### Scenario: No silent failure
- **WHEN** a title fails
- **THEN** the log names the plugin, the slug and the upstream error
