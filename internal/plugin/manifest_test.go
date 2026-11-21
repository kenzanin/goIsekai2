package plugin

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// writePlugin lays down a minimal valid plugin folder and returns the
// manifest path.
func writePlugin(t *testing.T, dir, id string, manifest string) string {
	t.Helper()
	root := filepath.Join(dir, id)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if manifest == "" {
		manifest = `{"name":"` + id + `","version":"1.2.3","runtime":"lua","capabilities":["pages"],"api":1,"entry":"main.lua"}`
	}
	if err := os.WriteFile(filepath.Join(root, "main.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.lua"), []byte("--"), 0o644); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, "main.json")
}

func TestLoadManifestValid(t *testing.T) {
	path := writePlugin(t, t.TempDir(), "asura", "")
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}
	want := Manifest{
		Name:         "asura",
		Version:      "1.2.3",
		Runtime:      "lua",
		Capabilities: []string{"pages"},
		API:          1,
		Entry:        "main.lua",
	}
	if !reflect.DeepEqual(m, &want) {
		t.Fatalf("got %+v, want %+v", m, want)
	}
}

func TestLoadManifestRejections(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		wantErr  string
	}{
		{"api", `{"api":2,"runtime":"lua","entry":"main.lua"}`, "unsupported manifest api 2"},
		{"runtime", `{"api":1,"runtime":"exe","entry":"main.lua"}`, `unknown runtime "exe"`},
		{"missing entry file", `{"api":1,"runtime":"js","entry":"nope.js"}`, "entry \"nope.js\""},
		{"empty entry", `{"api":1,"runtime":"lua"}`, "entry \"\""},
		{"malformed json", `{`, ":"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writePlugin(t, t.TempDir(), "broken", tt.manifest)
			_, err := LoadManifest(path)
			if err == nil {
				t.Fatalf("manifest accepted, want error %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) || !strings.Contains(err.Error(), path) {
				t.Fatalf("error %q: want %q and path %q", err, tt.wantErr, path)
			}
		})
	}
}

func TestLoadManifestMissingFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadManifest(filepath.Join(dir, "absent", "main.json")); err == nil {
		t.Fatal("missing manifest accepted")
	}
}
