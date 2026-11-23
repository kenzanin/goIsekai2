package detail

import (
	"context"
	"path/filepath"
	"testing"

	"goisekai2/internal/store"
)

type countingSource struct{ calls int }

func (c *countingSource) Call(_ context.Context, _ []byte) ([]byte, error) {
	c.calls++
	return []byte(`{"ok":true,"data":{"id":"cand","name":"Candidate"}}`), nil
}

func TestCandidatesAreCached(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	source := &countingSource{}
	svc := &Service{store: st, enrichment: []*EnrichmentSource{{ID: "fake", rt: source}}}
	ctx := context.Background()

	if got := svc.Candidates(ctx, "One Piece", "detail"); len(got) != 0 {
		t.Fatalf("cold Candidates should be empty, got %#v", got)
	}
	if source.calls != 0 {
		t.Fatalf("Candidates called source %d times", source.calls)
	}
	refreshed := svc.RefreshCandidates(ctx, "One Piece", "detail")
	if len(refreshed) != 1 || refreshed[0]["id"] != "cand" {
		t.Fatalf("refreshed = %#v", refreshed)
	}
	for i := 0; i < 2; i++ {
		if got := svc.Candidates(ctx, "One Piece", "detail"); len(got) != 1 || got[0]["id"] != "cand" {
			t.Fatalf("read %d = %#v", i, got)
		}
	}
	if source.calls != 1 {
		t.Fatalf("source calls = %d, want 1", source.calls)
	}
}
