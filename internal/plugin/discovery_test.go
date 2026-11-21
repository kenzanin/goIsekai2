package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverSkipsBrokenKeepsValidInNameOrder(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "zeta", "")     // valid, later in name order
	writePlugin(t, dir, "alpha", "")    // valid, first in name order
	writePlugin(t, dir, "badjson", `{`) // malformed manifest
	writePlugin(t, dir, "noentry", `{"api":1,"runtime":"lua","entry":"gone.lua"}`)
	plain := filepath.Join(dir, "plainfile")
	if err := os.WriteFile(plain, []byte("x"), 0o644); err != nil { // not a plugin folder
		t.Fatal(err)
	}

	got, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	ids := make([]string, 0, len(got))
	for _, e := range got {
		ids = append(ids, e.ID)
		if e.Manifest == nil || e.Manifest.Name != e.ID {
			t.Fatalf("entry %q: manifest mismatch %+v", e.ID, e.Manifest)
		}
	}
	if len(ids) != 2 || ids[0] != "alpha" || ids[1] != "zeta" {
		t.Fatalf("ids = %v, want [alpha zeta] with broken folders skipped", ids)
	}
}

func TestDiscoverMissingDir(t *testing.T) {
	if _, err := Discover(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("missing dir accepted")
	}
}
