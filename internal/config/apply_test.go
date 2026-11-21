package config

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

var errApplier = errors.New("apply failed")

func loadTestIni(t *testing.T, content string) *Ini {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cfg.ini")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	ini, err := LoadIni(path)
	if err != nil {
		t.Fatal(err)
	}
	return ini
}

func TestApplySnapshotCallsOnlyRegisteredAppliers(t *testing.T) {
	ResetAppliers()
	t.Cleanup(ResetAppliers)

	calls := make(map[string]string)
	RegisterLiveApplier("log.level", func(v string) error {
		calls["log.level"] = v
		return nil
	})

	reg := Knobs()
	snap, err := Resolve(reg, loadTestIni(t, "[log]\nlevel = debug\n"), nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	restart := ApplySnapshot(snap)
	if calls["log.level"] != "debug" {
		t.Fatalf("applier got %q, want debug", calls["log.level"])
	}
	if len(calls) != 1 {
		t.Fatalf("applier called for %d keys, want 1", len(calls))
	}
	for _, key := range restart {
		if key == "log.level" {
			t.Fatal("log.level reported as restart-required")
		}
	}
	if len(restart) == 0 {
		t.Fatal("expected restart-required keys for unapplied knobs")
	}
}

func TestApplySnapshotDuplicateRegistrationReplaces(t *testing.T) {
	ResetAppliers()
	t.Cleanup(ResetAppliers)

	got := ""
	RegisterLiveApplier("log.level", func(string) error { got = "old"; return nil })
	RegisterLiveApplier("log.level", func(v string) error { got = v; return nil })

	if !HasLiveApplier("log.level") {
		t.Fatal("registration should be visible")
	}
	reg := Knobs()
	snap, err := Resolve(reg, loadTestIni(t, "[log]\nlevel = warning\n"), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	ApplySnapshot(snap)
	if got != "warning" {
		t.Fatalf("replacement applier got %q, want warning", got)
	}
}

func TestApplySnapshotFailedApplierNeedsRestart(t *testing.T) {
	ResetAppliers()
	t.Cleanup(ResetAppliers)

	RegisterLiveApplier("log.level", func(string) error { return errApplier })
	reg := Knobs()
	snap, _ := Resolve(reg, loadTestIni(t, "[log]\nlevel = info\n"), nil, nil)

	restart := ApplySnapshot(snap)
	found := false
	for _, key := range restart {
		if key == "log.level" {
			found = true
		}
	}
	if !found {
		t.Fatalf("failed applier not in restart list: %v", restart)
	}
	if !sort.StringsAreSorted(restart) {
		t.Fatalf("restart list not sorted: %v", restart)
	}
}
