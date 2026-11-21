package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/gorm"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestTxRejectsNesting(t *testing.T) {
	s := testStore(t)
	var innerErr error
	err := s.Tx(context.Background(), func(inner context.Context, _ *gorm.DB) error {
		innerErr = s.Tx(inner, func(context.Context, *gorm.DB) error { return nil })
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !errors.Is(innerErr, ErrNestedTx) {
		t.Fatalf("inner error = %v, want nested tx", innerErr)
	}
	var nested *NestedTxError
	if !errors.As(innerErr, &nested) || nested.Outer == "" {
		t.Fatalf("nested error lacks outer name: %v", innerErr)
	}
}

func TestTxCommitsAndRollsBack(t *testing.T) {
	s := testStore(t)
	if err := s.Tx(context.Background(), func(_ context.Context, tx *gorm.DB) error {
		return tx.Table("plugin").Create(map[string]any{"name": "a", "plugin_id": "a"}).Error
	}); err != nil {
		t.Fatal(err)
	}
	err := s.Tx(context.Background(), func(_ context.Context, tx *gorm.DB) error {
		if err := tx.Table("plugin").Create(map[string]any{"name": "b", "plugin_id": "b"}).Error; err != nil {
			return err
		}
		return errors.New("abort")
	})
	if err == nil {
		t.Fatal("expected rollback")
	}
	var count int64
	if err := s.DB().Table("plugin").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("rows = %d, want 1", count)
	}
}
