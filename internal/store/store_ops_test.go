package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStatusFoldPreservesVerbatimString(t *testing.T) {
	if got := FoldStatus("ongoin", 0); got&StatusOngoing == 0 {
		t.Fatalf("ongoin flags = %d", got)
	}
	if got := FoldStatus("Onhaiatus", 0); got&StatusHiatus == 0 {
		t.Fatalf("Onhaiatus flags = %d", got)
	}
	if got := FoldStatus("???", StatusOngoing); got != StatusOngoing {
		t.Fatalf("unknown status changed category: %d", got)
	}
}

func TestChapterFlagsAndRollup(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	if err := s.UpsertSaved(ctx, &SavedManga{Plugin: "p", Slug: "s", Name: "T"}); err != nil {
		t.Fatal(err)
	}
	chapters := []Chapter{
		{ChapterID: "1", Number: 1},
		{ChapterID: "2", Number: 2},
	}
	if err := s.UpsertChapters(ctx, "p", "s", chapters); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkChapterRead(ctx, "p", "s", "1"); err != nil {
		t.Fatal(err)
	}
	m, err := s.GetSaved(ctx, "p", "s")
	if err != nil {
		t.Fatal(err)
	}
	if m.TotalCh != 2 || m.ReadCh != 1 {
		t.Fatalf("rollup = %d/%d, want 2/1", m.ReadCh, m.TotalCh)
	}
	if err := s.MarkChapterRead(ctx, "p", "s", "2"); err != nil {
		t.Fatal(err)
	}
	m, _ = s.GetSaved(ctx, "p", "s")
	if m.ReadCh != m.TotalCh || m.StatusFlags&StatusDone == 0 {
		t.Fatalf("done not set: %+v", m)
	}
}

func TestCacheTTLAndInvalidate(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	now := time.Now().Unix()
	if err := s.PutCached(ctx, &ApiCache{PluginID: "p", QueryHash: "h", Payload: "body", FetchedAt: now, TTL: 60}); err != nil {
		t.Fatal(err)
	}
	if payload, hit := s.GetCached(ctx, "p", "h", now+59); !hit || payload != "body" {
		t.Fatalf("fresh hit = %q, %v", payload, hit)
	}
	if _, hit := s.GetCached(ctx, "p", "h", now+61); hit {
		t.Fatal("expired entry was a hit")
	}
	if err := s.InvalidateCached(ctx, "p", "h", false); err != nil {
		t.Fatal(err)
	}
	if _, hit := s.GetCached(ctx, "p", "h", now); hit {
		t.Fatal("invalidated entry was a hit")
	}
}

func TestSourceMigrationBoundaryIncludesSubchapters(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	if err := s.UpsertChapters(ctx, "target", "s", []Chapter{
		{ChapterID: "1", Number: 1}, {ChapterID: "6", Number: 6},
		{ChapterID: "6.1", Number: 6.1}, {ChapterID: "7", Number: 7},
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CarryReadToTarget(ctx, "target", "s", 6); err != nil {
		t.Fatal(err)
	}
	var chapters []Chapter
	if err := s.DB().Where("plugin = ? AND slug = ?", "target", "s").Find(&chapters).Error; err != nil {
		t.Fatal(err)
	}
	read := make(map[string]bool)
	for _, c := range chapters {
		read[c.ChapterID] = c.Flags&ChapterRead != 0
	}
	for _, id := range []string{"1", "6", "6.1"} {
		if !read[id] {
			t.Fatalf("chapter %s not carried", id)
		}
	}
	if read["7"] {
		t.Fatal("chapter 7 should stay unread")
	}
}

func TestDedupGroupsAndBackup(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	for _, slug := range []string{"a", "b", "c"} {
		m := &SavedManga{Plugin: "p", Slug: slug, Name: "Same Name", Synopsis: "shared plot words", Authors: `["author"]`}
		if err := s.UpsertSaved(ctx, m); err != nil {
			t.Fatal(err)
		}
		if err := s.IndexSavedTitle(ctx, m); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.RecomputeDupGroups(ctx); err != nil {
		t.Fatal(err)
	}
	var groups []int
	if err := s.DB().Model(&SavedManga{}).Where("plugin = ?", "p").Order("slug").Pluck("dup_group", &groups).Error; err != nil {
		t.Fatal(err)
	}
	if len(groups) != 3 || groups[0] == 0 || groups[0] != groups[1] || groups[1] != groups[2] {
		t.Fatalf("dup groups = %v, want one shared non-zero id", groups)
	}

	dest, err := s.BackupTo(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dest)
	if err != nil || info.Size() == 0 {
		t.Fatalf("backup %s: %v", dest, err)
	}
	_ = filepath.Base(dest)
}

func TestPositionsAndMetadataRoundTrip(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	if err := s.PutPosition(ctx, &ReaderPosition{Plugin: "p", Slug: "s", ChapterID: "2", ChapterNumber: 2, Page: 3, Mode: "webtoon"}); err != nil {
		t.Fatal(err)
	}
	pos, err := s.GetPosition(ctx, "p", "s")
	if err != nil || pos.Page != 3 || pos.ChapterNumber != 2 {
		t.Fatalf("position = %+v, %v", pos, err)
	}
	if err := s.PutPluginState(ctx, "p", "k", "v"); err != nil {
		t.Fatal(err)
	}
	value, err := s.GetPluginState(ctx, "p", "k")
	if err != nil || value != "v" {
		t.Fatalf("plugin state = %q, %v", value, err)
	}
	if err := s.PutTitleMeta(ctx, &TitleMeta{Plugin: "p", Slug: "s", Fetched: `[1]`, Chosen: `{}`, FetchedAt: 42}); err != nil {
		t.Fatal(err)
	}
	meta, err := s.GetTitleMeta(ctx, "p", "s")
	if err != nil || meta.Fetched != `[1]` || meta.FetchedAt != 42 {
		t.Fatalf("meta = %+v, %v", meta, err)
	}
	if err := s.PutUpdateAttempt(ctx, &UpdateAttempt{Plugin: "p", Slug: "s", Outcome: "error", ConsecutiveErrors: 2, BackoffUntil: 99}); err != nil {
		t.Fatal(err)
	}
	attempt, err := s.GetUpdateAttempt(ctx, "p", "s")
	if err != nil || attempt.BackoffUntil != 99 || attempt.ConsecutiveErrors != 2 {
		t.Fatalf("attempt = %+v, %v", attempt, err)
	}
}

func TestLikeFallbackWhenFTSUnavailable(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	if err := s.UpsertSaved(ctx, &SavedManga{Plugin: "p", Slug: "s", Name: "Fallback Title", Synopsis: "needle"}); err != nil {
		t.Fatal(err)
	}
	fallback := &Store{db: s.DB(), ftsOK: false}
	hits, err := fallback.SearchTitles(ctx, "needle")
	if err != nil || len(hits) != 1 || hits[0].Slug != "s" {
		t.Fatalf("fallback hits = %+v, %v", hits, err)
	}
}
