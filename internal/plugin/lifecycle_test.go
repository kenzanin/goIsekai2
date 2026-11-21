package plugin

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"goisekai2/internal/events"
	"goisekai2/internal/store"
)

func lifecycleStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestLifecycleTransitions(t *testing.T) {
	ctx := context.Background()
	st := lifecycleStore(t)
	bus := events.New()
	defer bus.Close()
	rec := bus.Record(events.PluginLifecycle)
	lc := NewLifecycle(st, bus, "p", StateDisabled)

	if err := lc.Fire(ctx, TriggerEnable); err != nil {
		t.Fatal(err)
	}
	state, _ := lc.State(ctx)
	if !strings.HasPrefix(state, "Enabled") {
		t.Fatalf("state = %s", state)
	}
	if err := lc.Fire(ctx, TriggerCallStarted); err != nil {
		t.Fatal(err)
	}
	state, _ = lc.State(ctx)
	if state != StateBusy {
		t.Fatalf("busy state = %s", state)
	}
	if err := lc.Fire(ctx, TriggerCallFinished); err != nil {
		t.Fatal(err)
	}
	if err := lc.Fire(ctx, TriggerDisable); err != nil {
		t.Fatal(err)
	}
	state, _ = lc.State(ctx)
	if state != StateDisabled {
		t.Fatalf("final state = %s", state)
	}
	if len(rec.Items()) != 4 {
		t.Fatalf("events = %d, want 4", len(rec.Items()))
	}
}

func TestLifecycleIllegalTrigger(t *testing.T) {
	ctx := context.Background()
	lc := NewLifecycle(lifecycleStore(t), nil, "p", StateDisabled)
	err := lc.Fire(ctx, TriggerCallStarted)
	if !IsIllegal(err) {
		t.Fatalf("error = %v, want illegal trigger", err)
	}
	if lc.IllegalCount() != 1 {
		t.Fatalf("counter = %d", lc.IllegalCount())
	}
}
