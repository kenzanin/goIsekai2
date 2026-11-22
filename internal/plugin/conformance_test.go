package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"goisekai2/internal/plugin/tools"
)

func TestBundledPluginContract(t *testing.T) {
	root := filepath.Join("..", "..", "plugins")
	for _, group := range []string{"lua", "js"} {
		dir := filepath.Join(root, group)
		entries, err := Discover(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.ID != entry.Manifest.ID {
				t.Fatalf("%s: folder %s != manifest %s", entry.ID, entry.ID, entry.Manifest.ID)
			}
			if entry.Manifest.Entry == "" {
				t.Fatalf("%s: missing entry", entry.ID)
			}
			if _, err := os.Stat(filepath.Join(dir, entry.ID, entry.Manifest.Entry)); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestMangadexReferenceCompiles(t *testing.T) {
	path := filepath.Join("..", "..", "plugins", "lua", "mangadex", "main.lua")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewLuaRuntimeWithTools(string(src), &tools.ToolSet{}); err != nil {
		t.Fatalf("mangadex reference: %v", err)
	}
}

func TestCanonicalEnvelopeShapes(t *testing.T) {
	fixtures := []string{
		`{"ok":true,"data":{"items":[{"id":"x","title":"T","cover_url":"https://x/y"}],"total":1}}`,
		`{"ok":true,"data":{"chapters":[{"number":1,"title":"ch1","release_at":1700000000}]}}`,
		`{"ok":true,"data":["https://cdn.example/p001.webp"]}`,
	}
	for _, raw := range fixtures {
		var env map[string]any
		if err := json.Unmarshal([]byte(raw), &env); err != nil {
			t.Fatal(err)
		}
		if env["ok"] != true {
			t.Fatalf("fixture not ok: %s", raw)
		}
	}
}
