package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixture = `# banner comment
[sec]
key = value
empty =
other = x ; inline comment
unknown = keep
`

func TestLoadIniMissing(t *testing.T) {
	ini, err := LoadIni(filepath.Join(t.TempDir(), "nope.ini"))
	if err != nil {
		t.Fatal(err)
	}
	if got := ini.Get("sec", "key"); got != "" {
		t.Fatalf("Get on missing file = %q, want %q", got, "")
	}
}

func TestIni(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.ini")
	if err := os.WriteFile(path, []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}
	ini, err := LoadIni(path)
	if err != nil {
		t.Fatal(err)
	}
	ini.Set("sec", "key", "new")
	if err := ini.Save(path); err != nil {
		t.Fatal(err)
	}
	ini2, err := LoadIni(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("changed key", func(t *testing.T) {
		if got := ini2.Get("sec", "key"); got != "new" {
			t.Errorf("Get key = %q, want %q", got, "new")
		}
	})
	t.Run("unknown key kept", func(t *testing.T) {
		if got := ini2.Get("sec", "unknown"); got != "keep" {
			t.Errorf("Get unknown = %q, want %q", got, "keep")
		}
	})
	t.Run("empty value kept", func(t *testing.T) {
		if got := ini2.Get("sec", "empty"); got != "" {
			t.Errorf("Get empty = %q, want %q", got, "")
		}
	})
	t.Run("banner comment kept", func(t *testing.T) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "# banner comment") {
			t.Errorf("banner comment lost, got:\n%s", data)
		}
	})
}

func TestSaveMissingDir(t *testing.T) {
	ini, err := LoadIni(filepath.Join(t.TempDir(), "x.ini"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ini.Save(filepath.Join(t.TempDir(), "no", "such", "a.ini")); err == nil {
		t.Fatal("Save to missing dir = nil error, want error")
	}
}

func TestSaveReadOnly(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "ro")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "a.ini")
	orig := []byte("orig = 1\n")
	if err := os.WriteFile(path, orig, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	ini, err := LoadIni(path)
	if err != nil {
		t.Fatal(err)
	}
	ini.Set("", "orig", "2")
	saveErr := ini.Save(path)
	os.Chmod(dir, 0o700) // restore so t.TempDir cleanup works
	if saveErr == nil {
		t.Fatal("Save to read-only dir = nil error, want error")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(orig) {
		t.Errorf("original bytes changed: %q, want %q", got, orig)
	}
}

func TestSetNewSection(t *testing.T) {
	ini, err := LoadIni(filepath.Join(t.TempDir(), "x.ini"))
	if err != nil {
		t.Fatal(err)
	}
	ini.Set("brandnew", "k", "v")
	if got := ini.Get("brandnew", "k"); got != "v" {
		t.Fatalf("Get on new section = %q, want %q", got, "v")
	}
}
