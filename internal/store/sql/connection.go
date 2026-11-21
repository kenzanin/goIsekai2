// Package sql owns the store's single SQLite connection and all raw SQL:
// opening the database, its pragmas, and the queries of the store layer.
package sql

import (
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// OpenError reports a failed database open or pragma step.
type OpenError struct {
	Op  string // failed step, e.g. "open" or "pragma journal_mode"
	Err error  // underlying cause
}

// Error renders the failed step and cause.
func (e *OpenError) Error() string { return "store/sql: " + e.Op + ": " + e.Err.Error() }

// Unwrap exposes the underlying cause for errors.Is/As.
func (e *OpenError) Unwrap() error { return e.Err }

// Open opens the SQLite database at path via the pure-Go glebarez driver,
// caps the pool at one connection, and applies the WAL, busy-timeout and
// foreign-key pragmas before returning.
func Open(path string) (*gorm.DB, error) {
	dsn := path
	if !strings.Contains(dsn, "?") {
		dsn += "?_txlock=immediate"
	}
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, &OpenError{Op: "open", Err: err}
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, &OpenError{Op: "pool handle", Err: err}
	}
	sqlDB.SetMaxOpenConns(1)
	for _, p := range [...][2]string{
		{"journal_mode", "WAL"},
		{"busy_timeout", "5000"},
		{"foreign_keys", "ON"},
	} {
		if err := db.Exec("PRAGMA " + p[0] + "=" + p[1]).Error; err != nil {
			return nil, &OpenError{Op: "pragma " + p[0], Err: err}
		}
	}
	return db, nil
}
