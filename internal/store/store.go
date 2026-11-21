// Package store exposes typed access to the single SQLite library.
package store

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	sqlb "goisekai2/internal/store/sql"
)

type txMarkerKey struct{}

// NestedTxError reports an attempt to open a transaction inside another one.
type NestedTxError struct {
	Outer string
}

// Error names the enclosing transaction.
func (e *NestedTxError) Error() string {
	return fmt.Sprintf("store: nested transaction inside %s", e.Outer)
}

// ErrNestedTx is the sentinel matched by errors.Is for nested transactions.
var ErrNestedTx = &NestedTxError{}

// Is lets callers match any NestedTxError with errors.Is(err, ErrNestedTx).
func (e *NestedTxError) Is(target error) bool {
	return target == ErrNestedTx
}

// Store owns the GORM handle and all typed repository methods.
type Store struct {
	db    *gorm.DB
	ftsOK bool
}

// New wraps an already-opened GORM handle.
func New(db *gorm.DB) *Store {
	return &Store{db: db, ftsOK: sqlb.EnsureFTS(db)}
}

// DB returns the underlying handle for package-internal repository methods.
func (s *Store) DB() *gorm.DB { return s.db }

// Open opens the library at path and applies the frozen schema.
func Open(path string) (*Store, error) {
	db, err := sqlb.Open(path)
	if err != nil {
		return nil, err
	}
	if err := sqlb.Migrate(db); err != nil {
		return nil, err
	}
	return New(db), nil
}

// FTSAvailable reports whether the FTS5 index is usable.
func (s *Store) FTSAvailable() bool { return s.ftsOK }

// Migrate applies pending migrations on an existing handle.
func (s *Store) Migrate() error { return sqlb.Migrate(s.db) }

// Tx runs fn in one transaction and rejects nesting with a typed error.
func (s *Store) Tx(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	if outer, ok := ctx.Value(txMarkerKey{}).(string); ok {
		return &NestedTxError{Outer: outer}
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		inner := context.WithValue(ctx, txMarkerKey{}, "current")
		return fn(inner, tx.WithContext(inner))
	})
}

// Close releases the underlying connection pool.
func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// IsNestedTx reports whether err is a nested-transaction rejection.
func IsNestedTx(err error) bool {
	var nested *NestedTxError
	return errors.As(err, &nested)
}
