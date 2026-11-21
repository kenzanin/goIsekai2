package tools

import (
	"os"
	"path/filepath"
	"testing"

	"goisekai2/internal/store"
)

func testStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o600)
}

func TestDispatchCoreTools(t *testing.T) {
	ts := &ToolSet{Store: testStore(t), Dir: t.TempDir(), PluginID: "p"}
	if _, err := ts.Dispatch("now"); err != nil {
		t.Fatal(err)
	}
	if err := ts.DispatchAndIgnore("storage.set", "k", "v"); err != nil {
		t.Fatal(err)
	}
	got, err := ts.Dispatch("storage.get", "k")
	if err != nil || got != "v" {
		t.Fatalf("storage = %v, %v", got, err)
	}
	if got, err := ts.Dispatch("url_encode", "a b"); err != nil || got != "a+b" {
		t.Fatalf("encode = %v, %v", got, err)
	}
	if got, err := ts.Dispatch("crypto.hash", "sha256", "abc"); err != nil || got != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("hash = %v, %v", got, err)
	}
	if _, err := ts.Dispatch("unknown.tool"); err == nil {
		t.Fatal("unknown tool accepted")
	}
}

func TestDispatchFSToolIsConfined(t *testing.T) {
	dir := t.TempDir()
	if err := writeTestFile(filepath.Join(dir, "a.txt"), "hello"); err != nil {
		t.Fatal(err)
	}
	ts := &ToolSet{Dir: dir}
	if got, err := ts.Dispatch("fs.read", "a.txt"); err != nil || got != "hello" {
		t.Fatalf("read = %v, %v", got, err)
	}
	if _, err := ts.Dispatch("fs.read", "../a.txt"); err == nil {
		t.Fatal("escape accepted")
	}
}

// DispatchAndIgnore is a test helper that drops the result of a side-effect tool.
func (t *ToolSet) DispatchAndIgnore(name string, args ...any) error {
	_, err := t.Dispatch(name, args...)
	return err
}
