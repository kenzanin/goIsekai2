package detail

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"goisekai2/internal/store"
)

func testStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestDetailServesSavedRowWithoutHost(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)
	if err := st.UpsertSaved(ctx, &store.SavedManga{
		Plugin: "p", Slug: "s", Name: "Name", Cover: "cover", Synopsis: "syn",
	}); err != nil {
		t.Fatal(err)
	}
	data, err := New(st, nil).Detail(ctx, "p", "s")
	if err != nil {
		t.Fatal(err)
	}
	if data["cached"] != true || data["name"] != "Name" {
		t.Fatalf("detail = %v", data)
	}
}

func TestDetailStaleFallbackAndRefreshStamp(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)
	svc := New(st, nil)
	raw := `{"ok":true,"data":{"name":"Stale"}}`
	if err := st.PutCached(ctx, &store.ApiCache{
		PluginID: "p", QueryHash: cacheHash("p", "detail", "s"),
		Payload: raw, FetchedAt: time.Now().Unix(), TTL: 86400,
	}); err != nil {
		t.Fatal(err)
	}
	data, err := svc.Detail(ctx, "p", "s")
	if err != nil || data["cached"] != true || data["name"] != "Stale" {
		t.Fatalf("stale = %v, %v", data, err)
	}
	if err := svc.Refresh(ctx, "p", "s"); err == nil {
		t.Fatal("refresh without host should fail")
	}
	attempt, err := st.GetUpdateAttempt(ctx, "p", "s")
	if err != nil || attempt.Outcome != "error" {
		t.Fatalf("attempt = %+v, %v", attempt, err)
	}
}
