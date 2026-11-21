// Package sql is the single boundary where goIsekai2 keeps raw SQL.
//
// Every statement — queries, migrations and DDL — is declared here as a named
// constant or a typed helper. Other packages call typed store methods and never
// build SQL strings. scripts/sql-boundary-check.go enforces the rule.
package sql
