package library

import (
	"context"
	"path/filepath"
	"testing"

	"goisekai2/internal/events"
	"goisekai2/internal/store"
)

func testService(t *testing.T) (*Service, *store.Store, *events.Recorder) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	bus := events.New()
	t.Cleanup(bus.Close)
	return New(st, bus), st, bus.Record(events.LibrarySaved)
}

func TestSavePlaceholderProtectionAndEventOrder(t *testing.T) {
	ctx := context.Background()
	svc, st, rec := testService(t)

	if err := svc.Save(ctx, "p", "s", "Good Name", "cover1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Save(ctx, "p", "s", "s", "cover2"); err != nil {
		t.Fatal(err)
	}
	row, err := svc.Get(ctx, "p", "s")
	if err != nil || row.Name != "Good Name" || row.Cover != "cover2" {
		t.Fatalf("row = %+v, %v", row, err)
	}
	if len(rec.Items()) != 2 {
		t.Fatalf("events = %d, want 2", len(rec.Items()))
	}
	payload := rec.Items()[1].(events.LibraryPayload)
	if payload.Plugin != "p" || payload.Slug != "s" {
		t.Fatalf("payload = %+v", payload)
	}
	if err := svc.Save(ctx, "", "s", "x", ""); err == nil {
		t.Fatal("missing plugin accepted")
	}
	if len(rec.Items()) != 2 {
		t.Fatal("rejected save published an event")
	}
	_ = st
}

func TestRemoveAndBitToggles(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := testService(t)
	if err := svc.Save(ctx, "p", "s", "Name", ""); err != nil {
		t.Fatal(err)
	}
	if err := svc.SetFlag(ctx, "p", "s", store.StatusOngoing, 0); err != nil {
		t.Fatal(err)
	}
	if err := svc.ToggleDim(ctx, "p", "s"); err != nil {
		t.Fatal(err)
	}
	row, _ := svc.Get(ctx, "p", "s")
	if row.StatusFlags&store.StatusOngoing == 0 || row.StatusFlags&store.StatusDim == 0 {
		t.Fatalf("flags = %d", row.StatusFlags)
	}
	if err := svc.ToggleDim(ctx, "p", "s"); err != nil {
		t.Fatal(err)
	}
	row, _ = svc.Get(ctx, "p", "s")
	if row.StatusFlags&store.StatusDim != 0 || row.StatusFlags&store.StatusOngoing == 0 {
		t.Fatalf("dim toggle clobbered category: %d", row.StatusFlags)
	}
	if err := svc.TogglePin(ctx, "p", "s"); err != nil {
		t.Fatal(err)
	}
	row, _ = svc.Get(ctx, "p", "s")
	if row.Pinned != 1 {
		t.Fatalf("pinned = %d", row.Pinned)
	}
	if err := svc.Remove(ctx, "p", "s"); err != nil {
		t.Fatal(err)
	}
	if row, err := svc.Get(ctx, "p", "s"); err != nil || row.AddedAt != 0 {
		t.Fatalf("unsaved master row = %+v, %v", row, err)
	}
	rows, err := svc.List(ctx)
	if err != nil || len(rows) != 0 {
		t.Fatalf("library = %d rows, %v", len(rows), err)
	}
}
