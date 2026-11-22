---
type: Fact
title: updates.Maintenance isolates its three steps
description: "Prune, backup and recluster each run and log their own failure, so a failing step never blocks the rest."
tags: [updates, maintenance, slog, isolation]
generated: { by: agent/cli, at: "2026-11-22T03:26:49Z" }
status: stable
---

internal/updates.Scheduler.maintenance runs the three maintenance steps on one tick, each wrapped so an error is logged (s.log.Error) and execution continues; store.PruneExpiredCache, store.BackupWatermark and store.RecomputeDupGroups carry the state. Verified by TestMaintenanceIsolatedFailures (maintenance_test.go): a regular file where the backups directory belongs makes the backup step fail while the prune still removes only the expired api_cache row and the recluster still resets a preseeded dup_group back to 0.

# Related Concepts
- [Architecture overview](../architecture/overview.md): The maintenance steps follow the daemon rule that background work logs failures instead of aborting the tick.
