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

func TestApplyCandidateResetAndRelated(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)
	svc := New(st, nil)
	if err := st.UpsertSaved(ctx, &store.SavedManga{Plugin: "p", Slug: "a", Name: "A", Tags: `["Action","Fantasy"]`}); err != nil {
		t.Fatal(err)
	}
	if err := st.UpsertSaved(ctx, &store.SavedManga{Plugin: "p", Slug: "b", Name: "B", Tags: `["Action"]`}); err != nil {
		t.Fatal(err)
	}
	if err := svc.ApplyCandidate(ctx, "p", "a", map[string]any{"name": "Candidate", "cover": "new"}); err != nil {
		t.Fatal(err)
	}
	meta, _ := st.GetTitleMeta(ctx, "p", "a")
	if meta.Fetched == "" || meta.Chosen != "new" {
		t.Fatalf("meta = %+v", meta)
	}
	if err := svc.ResetDetail(ctx, "p", "a"); err != nil || meta.Fetched == "" {
		t.Fatalf("reset detail: %v", err)
	}
	meta, _ = st.GetTitleMeta(ctx, "p", "a")
	if meta.Fetched != "" || meta.Chosen != "new" {
		t.Fatalf("reset clobbered cover: %+v", meta)
	}
	if err := svc.ResetCover(ctx, "p", "a"); err != nil {
		t.Fatal(err)
	}
	related := svc.Related(ctx, "p", "a")
	if len(related) != 1 || related[0]["slug"] != "b" {
		t.Fatalf("related = %v", related)
	}
}
