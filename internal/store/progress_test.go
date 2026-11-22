package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestReadProgressRoundTrip(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()

	if got, err := s.ReadProgress(ctx, "p", "s", "c1"); err != nil || got != 0 {
		t.Fatalf("missing progress = %d, %v", got, err)
	}
	if err := s.PutReadProgress(ctx, "p", "s", "c1", 7); err != nil {
		t.Fatal(err)
	}
	if got, err := s.ReadProgress(ctx, "p", "s", "c1"); err != nil || got != 7 {
		t.Fatalf("progress = %d, %v", got, err)
	}
	got, err := s.ReadProgresses(ctx, "p", "s", []string{"c1", "c2"})
	if err != nil {
		t.Fatal(err)
	}
	if got["c1"] != 7 {
		t.Fatalf("progresses = %#v", got)
	}
	if _, ok := got["c2"]; ok {
		t.Fatalf("unexpected progress for c2: %#v", got)
	}
}
