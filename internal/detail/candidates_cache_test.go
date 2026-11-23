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

	first := svc.Candidates(ctx, "One Piece", "detail")
	if len(first) != 1 || first[0]["id"] != "cand" {
		t.Fatalf("first = %#v", first)
	}
	second := svc.Candidates(ctx, "One Piece", "detail")
	if len(second) != 1 || second[0]["id"] != "cand" {
		t.Fatalf("second = %#v", second)
	}
	if source.calls != 1 {
		t.Fatalf("source calls = %d, want 1", source.calls)
	}
}
