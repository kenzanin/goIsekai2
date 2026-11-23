package detail

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"goisekai2/internal/store"
)

func TestStageSearchResultWithNoHostCreatesNoMasterRow(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := &Service{store: st}
	if err := svc.StageSearchResult(context.Background(), "p", "s"); err == nil {
		t.Fatal("StageSearchResult without host should fail")
	}
	if _, err := st.GetSaved(context.Background(), "p", "s"); err == nil {
		t.Fatal("StageSearchResult must not create a master row")
	}
}

func TestPromoteCachedCopiesMetadataWithoutMembership(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	payload := `{"ok":true,"data":{"name":"Staged","cover":"cover","synopsis":"Syn","authors":["A"],"tags":["T"]}}`
	if err := st.PutCached(ctx, &store.ApiCache{
		PluginID: "p", QueryHash: cacheHash("p", "detail", "s"),
		Payload: payload, FetchedAt: time.Now().Unix(), TTL: 86400,
	}); err != nil {
		t.Fatal(err)
	}
	svc := &Service{store: st}
	if err := svc.PromoteCached(ctx, "p", "s"); err != nil {
		t.Fatal(err)
	}
	row, err := st.GetSaved(ctx, "p", "s")
	if err != nil {
		t.Fatal(err)
	}
	if row.Name != "Staged" || row.Synopsis != "Syn" || row.Authors != "A" || row.Tags != "T" {
		t.Fatalf("row = %+v", row)
	}
	if row.AddedAt != 0 {
		t.Fatalf("promotion must not set added_at, got %d", row.AddedAt)
	}
	staged, err := svc.Detail(ctx, "p", "s")
	if err != nil || staged["name"] != "Staged" || staged["saved"] != false {
		t.Fatalf("staged detail = %v, %v", staged, err)
	}
}
