---
type: Fact
title: gorm query chains are legal outside the raw-SQL boundary
description: gorm Where/Order chains outside internal/store/sql pass sql-boundary-check; only Raw/Exec/Query-style APIs are gated.
tags: [internal/reader, store, sql-boundary]
generated: { by: agent/mcp, at: "2026-11-22T00:19:32Z" }
status: stable
---

generated: { by: "kit agent", at: "2026-11-22" }

`scripts/sql-boundary-check.go` only gates raw-SQL APIs (`Raw`, `Exec`, `Query*`, `Prepare*`,
`NamedExec`) and `.sql` files; plain gorm query chains (`.Where("a = ? AND b = ?", ...)`,
`.Order("idx")`, `.Find`) elsewhere are legal and stay unflagged.

Evidence (with the new internal/reader chains in place):
- `go run scripts/sql-boundary-check.go internal` -> `sql-boundary: OK (internal/store/sql is the only raw-SQL module)`, exit 0.
- `CGO_ENABLED=0 go test ./internal/reader/...` -> `ok goisekai2/internal/reader 0.043s`, exit 0.

Consequence: a package needing a query the store does not expose (e.g. reader listing chapters)
may use `store.DB()` gorm chains without a store change; only raw SQL strings must move into
`internal/store/sql`. Keep conditions trivial in callers; anything beyond key lookup belongs in the store.

# Related Concepts
- [Raw SQL boundary](../conventions/sql-boundary.md): The reader relies on gorm chains via store.DB(), which this convention permits outside internal/store/sql.
