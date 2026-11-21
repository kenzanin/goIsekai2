---
type: Fact
title: "Store DB: pure-Go gorm + glebarez/sqlite verified under CGO_ENABLED=0"
description: gorm v1.31.2 + glebarez/sqlite v1.11.0 (modernc engine) passes the store tests under CGO_ENABLED=0 with single-connection pool and WAL/busy_timeout/foreign_keys pragmas applied in Open.
tags: [internal/store/sql, store, gorm, sqlite]
generated: { by: agent/mcp, at: "2026-11-21T16:36:39Z" }
status: stable
---

## Evidence

```
$ CGO_ENABLED=0 go test ./internal/store/sql
ok  	goisekai2/internal/store/sql	0.008s
$ go run scripts/godoc-check.go   # exit 0
```

`go get gorm.io/gorm@v1.31.2 github.com/glebarez/sqlite@v1.11.0 modernc.org/sqlite@v1.62.0` resolved with no cgo toolchain; `internal/store/sql.Open` uses the glebarez dialector, `SetMaxOpenConns(1)`, and executes `PRAGMA journal_mode=WAL`, `busy_timeout=5000`, `foreign_keys=ON`. The passing test reads all three pragmas back through `db.Raw`.

## Consequence

The store layer is proven buildable and runnable with `CGO_ENABLED=0`; gorm+glebarez is the settled data-access pair for `internal/store`, and all later store code goes through `sql.Open` rather than opening its own handles.

# Related Concepts
- [SQLite pragma readback forms: lowercase wal, numeric foreign_keys](sqlite-pragma-return-values.md): The pragma return-value gotcha was observed in the same Open verification that proved the pure-Go stack.
