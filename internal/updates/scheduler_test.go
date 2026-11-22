package updates

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"goisekai2/internal/config"
	"goisekai2/internal/store"
)

func testStore(t *testing.T) (*store.Store, string) {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st, dir
}

func testSnapshot() (*config.Snapshot, error) {
	return config.Resolve(config.Knobs(), nil, func(string) string { return "" },
		func(config.Entry) (string, bool) { return "", false })
}

func TestStampAttemptBackoff(t *testing.T) {
	ctx := context.Background()
	st, _ := testStore(t)
	now := time.Now().Unix()
	for i := 1; i <= 3; i++ {
		StampAttempt(ctx, st, "p", "s", false, "boom", now)
		got, err := st.GetUpdateAttempt(ctx, "p", "s")
		if err != nil || got.ConsecutiveErrors != i {
			t.Fatalf("attempt %d = %+v, %v", i, got, err)
		}
	}
	got, _ := st.GetUpdateAttempt(ctx, "p", "s")
	if got.BackoffUntil < now+8*3600 {
		t.Fatalf("third failure backoff = %d", got.BackoffUntil-now)
	}
	for i := 0; i < 8; i++ {
		StampAttempt(ctx, st, "p", "s", false, "boom", now)
	}
	got, _ = st.GetUpdateAttempt(ctx, "p", "s")
	if got.BackoffUntil != now+72*3600 {
		t.Fatalf("capped backoff = %d", got.BackoffUntil-now)
	}
	StampAttempt(ctx, st, "p", "s", true, "", now)
	got, _ = st.GetUpdateAttempt(ctx, "p", "s")
	if got.ConsecutiveErrors != 0 || got.BackoffUntil != 0 {
		t.Fatalf("success did not clear: %+v", got)
	}
}

func TestSchedulerRunStopsOnCancel(t *testing.T) {
	st, dir := testStore(t)
	s := New(st, nil, testSnapshot)
	s.DBPath = filepath.Join(dir, "lib.db")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop")
	}
}

func TestCandidateFilters(t *testing.T) {
	ctx := context.Background()
	st, _ := testStore(t)
	if err := st.UpsertPlugin(ctx, &store.PluginRow{Name: "P", PluginID: "p", Version: "1", Runtime: "lua", Enabled: 1}); err != nil {
		t.Fatal(err)
	}
	if err := st.UpsertPlugin(ctx, &store.PluginRow{Name: "Q", PluginID: "q", Version: "1", Runtime: "lua", Enabled: 0}); err != nil {
		t.Fatal(err)
	}
	rows := []store.SavedManga{
		{Plugin: "p", Slug: "ok", Name: "ok", ReadCh: 1, AddedAt: 1},
		{Plugin: "p", Slug: "read", Name: "read", ReadCh: 2, AddedAt: 1},
		{Plugin: "q", Slug: "off", Name: "off", ReadCh: 0, AddedAt: 1},
	}
	for _, row := range rows {
		if err := st.UpsertSaved(ctx, &row); err != nil {
			t.Fatal(err)
		}
	}
	StampAttempt(ctx, st, "p", "ok", false, "boom", time.Now().Unix())
	got, err := st.UpdateCandidates(ctx, time.Now().Unix(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("backed-off candidate selected: %+v", got)
	}
	StampAttempt(ctx, st, "p", "ok", true, "", time.Now().Unix())
	got, err = st.UpdateCandidates(ctx, time.Now().Unix(), 10)
	if err != nil || len(got) != 1 || got[0].Slug != "ok" {
		t.Fatalf("candidates = %+v, %v", got, err)
	}
}

func TestPaceDelaySkipsAfterLast(t *testing.T) {
	d := 500 * time.Millisecond
	if paceDelay(0, 3, d) != d || paceDelay(1, 3, d) != d || paceDelay(2, 3, d) != 0 {
		t.Fatal("delay must run between titles only")
	}
}
