package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "meta.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Read(dir, "sub/../meta.txt")
	if err != nil || got != "hello" {
		t.Fatalf("Read(sub/../meta.txt) = %q, %v; want %q, nil", got, err, "hello")
	}

	for _, tc := range []struct{ path, wantErr string }{
		{"/etc/passwd", "must be relative"},
		{"../escape.txt", "escapes the plugin directory"},
	} {
		if _, err := Read(dir, tc.path); err == nil || !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("Read(%q) err=%v; want containing %q", tc.path, err, tc.wantErr)
		}
	}
}
