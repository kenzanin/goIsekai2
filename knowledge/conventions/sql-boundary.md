---
type: Convention
title: Raw SQL boundary
description: All raw SQL lives in the internal/store/sql module; scripts/sql-boundary-check enforces it in make lint.
tags: [store, sql, gate, convention]
generated: { by: agent/cli, at: "2026-11-21T14:28:09Z" }
status: stable
---

## Rule

Raw SQL exists in exactly one module (Go package): `internal/store/sql`. The
module may be split into many files as it grows — `queries.go`, `migrations.go`,
and so on — but no raw SQL may appear in any other package. Enforcement is
`scripts/sql-boundary-check.go`, wired into `make lint`
(`lint: godoc sql-boundary`).

## Allowed inside the module

- Named SQL constants and typed helpers for queries, DDL and migrations, in any
  number of files belonging to the package.
- GORM / `database/sql` raw entry points: `Raw`, `Exec`, `ExecContext`, `Query*`,
  `Prepare*`, `NamedExec`, `MustExec`.

## Forbidden outside the module

- String literals containing `SELECT … FROM`, `INSERT INTO`, `UPDATE … SET`,
  `DELETE FROM`, `CREATE/ALTER/DROP TABLE`, `PRAGMA`, `BEGIN IMMEDIATE` or
  `SAVEPOINT`.
- Calls to the raw entry points listed above, in production code and tests.
- Standalone `.sql` files anywhere else in the tree.

## Adding a query

1. Declare the statement (or typed helper) in a file under
   `internal/store/sql/`.
2. Expose it through a typed method on the store repository; keep SQL details
   and scan logic out of callers.
3. Run `make lint`; the gate prints
   `sql-boundary: OK (internal/store/sql is the only raw-SQL module)` when clean,
   otherwise a `file:line` list and exit 1.

Fixtures in tests must call the typed helpers instead of opening their own raw
SQL connection. The gate scans `*_test.go` too.

# Related Concepts
- [Stack: libraries vs hand-written code](../architecture/stack.md): the review page points at the gate that keeps store SQL in one place
