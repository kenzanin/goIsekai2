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

// TestMaintenanceIsolatedFailures pins the isolation rule: a failing backup
// step still lets the prune and recluster steps run on the same tick.
func TestMaintenanceIsolatedFailures(t *testing.T) {
	ctx := context.Background()
	st, dir := testStore(t)
	snap, err := testSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	s := New(st, nil, testSnapshot)
	s.DBPath = filepath.Join(dir, "lib.db")

	if err := st.PutCached(ctx, &store.ApiCache{PluginID: "p", QueryHash: "old", Payload: "x", FetchedAt: now.Unix() - 100, TTL: 10}); err != nil {
		t.Fatal(err)
	}
	if err := st.PutCached(ctx, &store.ApiCache{PluginID: "p", QueryHash: "new", Payload: "y", FetchedAt: now.Unix(), TTL: 100}); err != nil {
		t.Fatal(err)
	}
	if err := st.UpsertSaved(ctx, &store.SavedManga{Plugin: "p", Slug: "s", Name: "Alpha", DupGroup: 7}); err != nil {
		t.Fatal(err)
	}
	// A directory where the snapshot file belongs makes the backup step fail.
	if err := os.MkdirAll(filepath.Join(dir, now.Format("20060102-150405")+".db"), 0o755); err != nil {
		t.Fatal(err)
	}

	s.maintenance(ctx, snap, now)

	if _, ok := st.GetCached(ctx, "p", "old", now.Unix()); ok {
		t.Fatal("prune did not run when the backup step failed")
	}
	if _, ok := st.GetCached(ctx, "p", "new", now.Unix()); !ok {
		t.Fatal("prune dropped a live cache row")
	}
	saved, err := st.GetSaved(ctx, "p", "s")
	if err != nil {
		t.Fatal(err)
	}
	if saved.DupGroup != 0 {
		t.Fatalf("recluster did not run when the backup step failed: dup_group = %d", saved.DupGroup)
	}
}
