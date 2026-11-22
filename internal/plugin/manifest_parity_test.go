package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifestRejectsFolderMismatch(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good")
	if err := os.MkdirAll(good, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := `{"id":"good","name":"Good","runtime":"lua","capabilities":[],"api":1,"entry":"main.lua"}`
	if err := os.WriteFile(filepath.Join(good, "main.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(good, "main.lua"), []byte("function handle() end"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(filepath.Join(good, "main.json")); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(dir, "bad")
	if err := os.MkdirAll(bad, 0o700); err != nil {
		t.Fatal(err)
	}
	mismatch := `{"id":"other","name":"Other","runtime":"lua","capabilities":[],"api":1,"entry":"main.lua"}`
	if err := os.WriteFile(filepath.Join(bad, "main.json"), []byte(mismatch), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bad, "main.lua"), []byte("function handle() end"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadManifest(filepath.Join(bad, "main.json")); err == nil {
		t.Fatal("folder/manifest id mismatch was accepted")
	}
}
