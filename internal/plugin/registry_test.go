package plugin

import (
	"testing"
)

func TestRegistryLookup(t *testing.T) {
	first := &Manifest{Name: "first", Runtime: "lua", API: 1}
	second := &Manifest{Name: "second", Runtime: "js", API: 1}
	r := NewRegistry([]Entry{
		{ID: "asura", Manifest: first},
		{ID: "dup", Manifest: second}, // later duplicate
		{ID: "dup", Manifest: second}, // ...must not displace the first
		{ID: "dup", Manifest: &Manifest{}},
	})

	if m, ok := r.Lookup("asura"); !ok || m != first {
		t.Fatalf("Lookup(asura) = (%+v, %v)", m, ok)
	}
	if m, ok := r.Lookup("dup"); !ok || m != second {
		t.Fatalf("duplicate id: first must win, got %+v", m)
	}
	if _, ok := r.Lookup("absent"); ok {
		t.Fatal("unknown id found")
	}
}

func TestNewRegistryFromDiscover(t *testing.T) {
	// End-to-end: discovery output feeds the registry unchanged.
	dir := t.TempDir()
	writePlugin(t, dir, "asura", "")
	entries, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := NewRegistry(entries).Lookup("asura")
	if !ok || m.Runtime != "lua" {
		t.Fatalf("Lookup after Discover = (%+v, %v)", m, ok)
	}
}
