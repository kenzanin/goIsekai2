---
type: Fact
title: "SQLite pragma readback forms: lowercase wal, numeric foreign_keys"
description: "SQLite PRAGMA results come back lowercase (\"wal\") and numeric (\"1\"), so equality checks against \"WAL\"/\"ON\" fail."
tags: [internal/store/sql, store, sqlite]
generated: { by: agent/mcp, at: "2026-11-21T16:36:30Z" }
status: stable
---

## Evidence

From `internal/store/sql/connection_test.go` (passing):

```
journal_mode -> "wal"   (lowercase, not "WAL")
busy_timeout -> "5000"
foreign_keys -> "1"     (integer rendered as string, not "ON")
```

Observed by scanning `db.Raw("PRAGMA <name>")` into a string against modernc/glebarez sqlite.

## Consequence

Any code or test comparing pragma results must expect the engine's canonical forms: lowercase `wal` and `0`/`1` integers. Comparing against `WAL` or `ON` would fail despite correct configuration.
