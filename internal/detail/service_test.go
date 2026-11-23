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

func TestDetailRendersStagedCacheWithoutFetching(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)
	svc := New(st, nil)
	raw := `{"ok":true,"data":{"name":"Stale"}}`
	if err := st.PutStage(ctx, "p", "s", raw, "", time.Now().Unix()); err != nil {
		t.Fatal(err)
	}
	data, err := svc.Detail(ctx, "p", "s")
	if err != nil || data["cached"] != true || data["name"] != "Stale" || data["saved"] != false {
		t.Fatalf("detail = %v, %v", data, err)
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

func TestMaterializeKeepsMasterRow(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)
	if err := st.UpsertSaved(ctx, &store.SavedManga{
		Plugin: "p", Slug: "s", Name: "Old", Cover: "old", Pinned: 1, StatusFlags: store.StatusDim,
	}); err != nil {
		t.Fatal(err)
	}
	svc := New(st, nil)
	svc.materialize(ctx, "p", "s", map[string]any{
		"title": "New", "cover_url": "new", "synopsis": "desc", "total_ch": float64(12),
	})
	row, err := st.GetSaved(ctx, "p", "s")
	if err != nil {
		t.Fatal(err)
	}
	if row.Name != "New" || row.Cover != "new" || row.Synopsis != "desc" || row.TotalCh != 12 {
		t.Fatalf("master row = %+v", row)
	}
	if row.Pinned != 1 || row.StatusFlags&store.StatusDim == 0 {
		t.Fatal("materialize clobbered user state")
	}
}
