package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConceptV3LayoutPaths(t *testing.T) {
	root := t.TempDir()
	got := PagePath(root, "mangadex", "Solo Leveling", 12, "8f3a2c9d", 3, "webp")
	want := filepath.Join(root, "img", "mangadex", "manga", "solo-leveling-051b627f", "ch000012-8f3a2c9d", "p003.webp")
	if got != want {
		t.Fatalf("page = %s, want %s", got, want)
	}
	if got := CoverPath(root, "p", "abc", 0.6667, "webp"); got != filepath.Join(root, "img", "p", "cover", "abc-6667.webp") {
		t.Fatalf("cover = %s", got)
	}
	if got := chapterDir(12.5, "a1b2c3d4e5f6"); got != "ch000012.5-c3d4e5f6" {
		t.Fatalf("chapter = %s", got)
	}
	if got := Namespace("folder-id", "Atsu.moe Reader"); got != "atsu-moe-reader" {
		t.Fatalf("namespace = %s", got)
	}
	if got := Namespace("folder-id", "!!!"); got != "folder-id" {
		t.Fatalf("fallback namespace = %s", got)
	}
}

func TestPoolRebuildEvictsOldestWithSidecar(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, "old.webp")
	sidecarOld := sidecarPath(old)
	newer := filepath.Join(root, "new.webp")
	for _, path := range []string{old, sidecarOld, newer} {
		if err := os.WriteFile(path, []byte("12345"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	past := time.Now().Add(-time.Hour)
	_ = os.Chtimes(old, past, past)
	_ = os.Chtimes(sidecarOld, past, past)

	pool := NewPool(root, 6)
	if err := pool.Rebuild(); err != nil {
		t.Fatal(err)
	}
	if pool.totalSize != 10 {
		t.Fatalf("total = %d", pool.totalSize)
	}
	if err := pool.Evict(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("old file not evicted")
	}
	if _, err := os.Stat(sidecarOld); !os.IsNotExist(err) {
		t.Fatal("sidecar not evicted")
	}
	if _, err := os.Stat(newer); err != nil {
		t.Fatal("new file evicted")
	}
}
