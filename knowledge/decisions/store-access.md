---
type: Decision
title: "Store access: GORM with one raw-SQL boundary"
description: GORM runs on the pure-Go glebarez/sqlite driver; raw SQL is confined to the internal/store/sql module and enforced by scripts/sql-boundary-check.
tags: [store, gorm, sqlite, raw-sql, gate]
generated: { by: agent/cli, at: "2026-11-21T14:28:09Z" }
status: stable
---

## Context

The no-cgo gate (`CGO_ENABLED=0 go build ./...`) rejects the official
`gorm.io/driver/sqlite`, whose go.mod requires `github.com/mattn/go-sqlite3`
(cgo). The pure-Go escape hatch is `github.com/glebarez/sqlite`, a fork of the
official driver backed by `modernc.org/sqlite`. The user approved GORM on the
condition that raw SQL stays in one module and a gate enforces it.

## Decision

- ORM: `gorm.io/gorm` v1.31.2 with `github.com/glebarez/sqlite` v1.11.0.
- SQLite engine: `modernc.org/sqlite` pinned to v1.62.0 (the fork declares an
  older version; MVS must not silently downgrade the engine).
- `AutoMigrate` stays disabled: the cppIsekai schema is frozen (nine tables,
  composite keys, verbatim status strings, bitfields, JSON blobs,
  `schema_version` migrations). Migrations, bit operations and the saved-stats
  rollup are written as SQL — but only inside the `internal/store/sql` module.
  That module may grow into many files; its files stay one package.
- All other packages call typed store methods. No `Raw`, `Exec`, `Query`,
  `Prepare` or `.sql` file outside the module.
- `scripts/sql-boundary-check.go` (stdlib `go/ast`) scans Go string literals for
  SQL statements, raw-SQL API calls and standalone `.sql` files. `make lint`
  runs it; a violation exits 1 with file and line.

## Alternatives considered

| Option | Verdict |
|---|---|
| Plain `database/sql` + hand-written scans | rejected by the user in favour of GORM |
| Official `gorm.io/driver/sqlite` | rejected: requires `mattn/go-sqlite3`, fails no-cgo |
| `go-jet/jet/v2` (V1's typed builder) | kept for review if typed queries become the bottleneck; reflection-based GORM is not a type-safety upgrade |

## Consequences and review triggers

- Adding a query means adding a named statement/helper inside
  `internal/store/sql` and a typed repository method that calls it — never an
  inline string at the call site. Split the module into files when it grows.
- Review the fork if it lags `modernc.org/sqlite` or if the gate reports false
  positives from legitimate non-SQL literals.
- Review this decision if the store starts duplicating scan logic or if GORM's
  transaction semantics conflict with the anti-nesting `BEGIN IMMEDIATE` rule
  in `openspec/specs/store/spec.md`.

# Related Concepts
- [Dependency stack: pure Go only (modernc sqlite, tls-client, gopher-lua, quickjs, wazero)](dependency-stack.md): extends the dependency stack with GORM on the pure-Go SQLite driver
- [Raw SQL boundary](../conventions/sql-boundary.md): the decision's enforcement mechanism is the raw-SQL boundary rule
