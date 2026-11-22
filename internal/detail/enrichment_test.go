package detail

import (
	"os"
	"path/filepath"
	"testing"

	"goisekai2/internal/plugin/tools"
)

func TestLoadEnrichmentSeparatesScrapers(t *testing.T) {
	root := t.TempDir()
	src := "../../plugins/enrichment/mangadex-enrichment"
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "mangadex-enrichment")
	if err := os.MkdirAll(dst, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	sources, err := LoadEnrichment(root, &tools.ToolSet{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 || sources[0].ID != "mangadex-enrichment" {
		t.Fatalf("sources = %+v", sources)
	}
}
