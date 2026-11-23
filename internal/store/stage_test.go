package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	sqlb "goisekai2/internal/store/sql"
)

func TestStageTableAndCacheView(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()

	if err := s.PutStage(ctx, "p", "s", `{"ok":true}`, `{"ok":true}`, time.Now().Unix()); err != nil {
		t.Fatal(err)
	}
	stage, err := s.GetStage(ctx, "p", "s")
	if err != nil || stage == nil || stage.DetailJSON != `{"ok":true}` {
		t.Fatalf("stage = %+v, %v", stage, err)
	}
	if err := s.DeleteStage(ctx, "p", "s"); err != nil {
		t.Fatal(err)
	}
	if stage, err := s.GetStage(ctx, "p", "s"); err != nil || stage != nil {
		t.Fatalf("deleted stage = %+v, %v", stage, err)
	}

	name, err := sqlb.FindView(s.DB(), "api_cache_view")
	if err != nil || name != "api_cache_view" {
		t.Fatalf("cache view = %q, %v", name, err)
	}
}
