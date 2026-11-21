package sql

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestOpenAppliesPragmas(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for name, want := range map[string]string{
		"journal_mode": "wal",
		"busy_timeout": "5000",
		"foreign_keys": "1",
	} {
		var got string
		if err := db.Raw("PRAGMA " + name).Scan(&got).Error; err != nil {
			t.Fatalf("PRAGMA %s: %v", name, err)
		}
		if got != want {
			t.Errorf("PRAGMA %s = %q, want %q", name, got, want)
		}
	}
}

func TestOpenTypedError(t *testing.T) {
	_, err := Open(filepath.Join(t.TempDir(), "missing-dir", "x.db"))
	var openErr *OpenError
	if !errors.As(err, &openErr) {
		t.Fatalf("want *OpenError, got %T (%v)", err, err)
	}
}
