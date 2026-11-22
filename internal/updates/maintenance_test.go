package updates

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"goisekai2/internal/store"
)

func TestPruneExpiredCache(t *testing.T) {
	ctx := context.Background()
	st, _ := testStore(t)
	now := time.Now().Unix()
	if err := st.PutCached(ctx, &store.ApiCache{PluginID: "p", QueryHash: "old", Payload: "x", FetchedAt: now - 100, TTL: 10}); err != nil {
		t.Fatal(err)
	}
	if err := st.PutCached(ctx, &store.ApiCache{PluginID: "p", QueryHash: "new", Payload: "y", FetchedAt: now, TTL: 100}); err != nil {
		t.Fatal(err)
	}
	if err := st.PruneExpiredCache(ctx, now); err != nil {
		t.Fatal(err)
	}
	if _, ok := st.GetCached(ctx, "p", "old", now); ok {
		t.Fatal("expired cache survived prune")
	}
	if _, ok := st.GetCached(ctx, "p", "new", now); !ok {
		t.Fatal("fresh cache was pruned")
	}
}

func TestBackupRotation(t *testing.T) {
	ctx := context.Background()
	st, dir := testStore(t)
	s := New(st, nil, testSnapshot)
	s.DBPath = filepath.Join(dir, "lib.db")
	if err := s.backup(ctx, filepath.Join(dir, "backups"), 2, time.Now()); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "backups"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("backups = %v, %v", entries, err)
	}
	mark, err := st.BackupWatermark(ctx)
	if err != nil || mark.IsZero() {
		t.Fatalf("watermark = %v, %v", mark, err)
	}
}
