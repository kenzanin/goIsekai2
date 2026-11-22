# Spec Delta

## ADDED Requirements

### Requirement: Scheduler daemon wiring

The daemon SHALL own exactly one maintenance goroutine started before listen and stopped on shutdown within one tick. Each tick SHALL re-read config knobs, run maintenance steps independently so one failure does not abort the batch, and log a summary through the shared log pipeline. A manual refresh SHALL stamp the same attempt rows as the scheduler.

#### Scenario: Clean shutdown
- **WHEN** the daemon context is cancelled
- **THEN** the goroutine exits within one tick and starts no new work

#### Scenario: Independent maintenance failures
- **WHEN** the backup step fails
- **THEN** the error is logged, its watermark is unchanged, and the update batch still runs
