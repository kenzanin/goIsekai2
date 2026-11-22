---
type: Fact
title: "Backup cadence watermark is the newest snapshot, not a store key"
description: "Backups are VACUUM snapshots in <db dir>/backups; cadence reads the newest library-*.db timestamp because the frozen schema has no key/value table."
tags: [updates, backup, vacuum, rotation]
generated: { by: agent/cli, at: "2026-11-22T03:18:24Z" }
status: stable
---

internal/updates.Backup reuses store.BackupTo (VACUUM INTO) and writes into <db dir>/backups when dir is empty. The backup/last_run watermark named by the updates spec is derived from the newest library-YYYYMMDD-HHMMSS.db snapshot instead of a store row: the frozen schema has no key/value table and the SQL boundary forbids adding one from internal/updates. Rotation keeps the backup_keep newest snapshots; a failed snapshot writes no file, so the watermark is untouched and the next tick retries.

# Related Concepts
- [Store access: GORM with one raw-SQL boundary](../decisions/store-access.md): VACUUM INTO stays in the store boundary; the watermark is derived from snapshot files instead of new SQL.
