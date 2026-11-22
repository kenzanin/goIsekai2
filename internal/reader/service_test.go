package reader

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"goisekai2/internal/plugin"
	"goisekai2/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	// No plugin is booted: every host call fails, exercising the stale path.
	return New(st, plugin.NewHost(st, nil, t.TempDir(), nil))
}

func seedChapters(t *testing.T, st *store.Store) {
	t.Helper()
	err := st.UpsertChapters(context.Background(), "p", "s", []store.Chapter{
		{ChapterID: "24b", Number: 24},
		{ChapterID: "10", Number: 10},
		{ChapterID: "24a", Number: 24},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPagesDuplicateNumbers(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	seedChapters(t, svc.Store)

	out, err := svc.Pages(ctx, "p", "s", 24, "")
	if err != nil {
		t.Fatal(err)
	}
	// Number-only request lands on the first tie by id; the second duplicate
	// is the next stop and keeps number 24.
	if got := out["prev_ch_id"]; got != "10" {
		t.Errorf("prev_ch_id = %v, want 10", got)
	}
	if got := out["next_ch"]; got != float64(24) {
		t.Errorf("next_ch = %v, want 24", got)
	}
	if got := out["next_ch_id"]; got != "24b" {
		t.Errorf("next_ch_id = %v, want 24b", got)
	}

	// Exact id selects the second duplicate; it mirrors back to the first.
	out, err = svc.Pages(ctx, "p", "s", 24, "24b")
	if err != nil {
		t.Fatal(err)
	}
	if got := out["prev_ch_id"]; got != "24a" {
		t.Errorf("prev_ch_id = %v, want 24a", got)
	}
	if _, ok := out["next_ch"]; ok {
		t.Error("next_ch present past the last chapter")
	}

	if _, err := svc.Pages(ctx, "p", "s", 99, ""); !errors.Is(err, ErrUnknownChapter) {
		t.Errorf("err = %v, want ErrUnknownChapter", err)
	}
}

func TestPagesFallsBackToStoredPages(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	seedChapters(t, svc.Store)
	err := svc.Store.ReplacePages(ctx, "p", "s", "24a", []store.ChapterPage{{URL: "u1"}, {URL: "u2"}})
	if err != nil {
		t.Fatal(err)
	}

	// Host call fails (plugin not booted); the stored list is served in idx order.
	out, err := svc.Pages(ctx, "p", "s", 24, "24a")
	if err != nil {
		t.Fatal(err)
	}
	pages, ok := out["pages"].([]string)
	if !ok || len(pages) != 2 || pages[0] != "u1" || pages[1] != "u2" {
		t.Fatalf("pages = %v, want [u1 u2]", out["pages"])
	}

	// Prefetch tolerates the same failure mode and unknown ids alike.
	if err := svc.Prefetch(ctx, "p", "s", "24b", "nope"); !errors.Is(err, ErrUnknownChapter) {
		t.Errorf("Prefetch err = %v, want ErrUnknownChapter for the bogus id", err)
	}
	if err := svc.Prefetch(ctx, "p", "s", "24b", ""); err != nil {
		t.Errorf("Prefetch err = %v, want nil", err)
	}
}

func TestPutPositionTracksProgressAndReadBit(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	seedChapters(t, svc.Store)
	if err := svc.Store.ReplacePages(ctx, "p", "s", "24a", []store.ChapterPage{{URL: "u1"}, {URL: "u2"}}); err != nil {
		t.Fatal(err)
	}
	fields := map[string]any{"chapter_id": "24a", "chapter": float64(24), "page": float64(1)}
	if err := svc.PutPosition(ctx, "p", "s", fields); err != nil {
		t.Fatal(err)
	}
	if got, _ := svc.Store.ReadProgress(ctx, "p", "s", "24a"); got != 1 {
		t.Fatalf("progress = %d, want 1", got)
	}
	fields["page"] = float64(2)
	if err := svc.PutPosition(ctx, "p", "s", fields); err != nil {
		t.Fatal(err)
	}
	var chapter store.Chapter
	if err := svc.Store.DB().Where("plugin = ? AND slug = ? AND chapter_id = ?", "p", "s", "24a").First(&chapter).Error; err != nil {
		t.Fatal(err)
	}
	if chapter.Flags&store.ChapterRead == 0 {
		t.Fatalf("read bit not set: flags = %d", chapter.Flags)
	}
}

func TestNeighbourIDs(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	seedChapters(t, svc.Store)
	next, prev := svc.NeighbourIDs(ctx, "p", "s", "24a")
	if next != "24b" || prev != "10" {
		t.Fatalf("neighbours = (%q, %q), want (24b, 10)", next, prev)
	}
}

func TestPutPositionPreservesAbsentFields(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	err := svc.Store.PutPosition(ctx, &store.ReaderPosition{
		Plugin: "p", Slug: "s", ChapterID: "24a", ChapterNumber: 24, Page: 3, Mode: "paged", Direction: "rtl",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = svc.PutPosition(ctx, "p", "s", map[string]any{"page": float64(9), "zoom": "fit", "bogus": 1})
	if err != nil {
		t.Fatal(err)
	}
	pos, err := svc.Store.GetPosition(ctx, "p", "s")
	if err != nil {
		t.Fatal(err)
	}
	if pos.Page != 9 || pos.Zoom != "fit" {
		t.Errorf("pos = %+v, want page 9 zoom fit", pos)
	}
	if pos.ChapterID != "24a" || pos.ChapterNumber != 24 || pos.Mode != "paged" || pos.Direction != "rtl" {
		t.Errorf("absent fields not preserved: %+v", pos)
	}
}

func TestWindowStopsAtLimitAfterPosition(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	seedChapters(t, svc.Store)
	for _, seed := range []struct {
		id   string
		urls []string
	}{
		{"10", []string{"a1", "a2", "a3"}},
		{"24a", []string{"b1", "b2"}},
	} {
		pages := make([]store.ChapterPage, len(seed.urls))
		for i, u := range seed.urls {
			pages[i] = store.ChapterPage{URL: u}
		}
		if err := svc.Store.ReplacePages(ctx, "p", "s", seed.id, pages); err != nil {
			t.Fatal(err)
		}
	}

	got := svc.Window(ctx, "p", "s", store.ReaderPosition{ChapterID: "10", ChapterNumber: 10, Page: 1}, 3)
	want := []string{"a2", "a3", "b1"}
	if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("Window = %v, want %v", got, want)
	}
}
