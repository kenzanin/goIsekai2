package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, dir, id, source string) {
	t.Helper()
	folder := filepath.Join(dir, id)
	if err := os.MkdirAll(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := `{"name":"Fixture","version":"1","runtime":"lua","capabilities":["search"],"api":1,"entry":"plugin.lua"}`
	if err := os.WriteFile(filepath.Join(folder, "main.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "plugin.lua"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestHostBootCallDisableAndPurge(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	writeFixture(t, dir, "p", `function handle(req) return '{"ok":true,"data":42}' end`)
	st := lifecycleStore(t)
	host := NewHost(st, nil, dir, nil)

	if err := host.Boot(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := host.Call(ctx, "p", "search", []byte(`{}`)); err == nil {
		t.Fatal("disabled plugin accepted a call")
	}
	if err := host.SetEnabled(ctx, "p", true); err != nil {
		t.Fatal(err)
	}
	out, err := host.Call(ctx, "p", "search", []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"ok":true,"data":42}` {
		t.Fatalf("out = %s", out)
	}
	rows := host.Rows(ctx)
	if len(rows) != 1 || rows[0].Status != "active" || rows[0].TitleCount != 0 {
		t.Fatalf("rows = %+v", rows)
	}
	if err := host.SetEnabled(ctx, "p", false); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(filepath.Join(dir, "p")); err != nil {
		t.Fatal(err)
	}
	if err := host.Boot(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := st.GetPlugin(ctx, "p"); err == nil {
		t.Fatal("purged plugin row still present")
	}
}

func TestHostMissingDirectoryDoesNotPrune(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	writeFixture(t, dir, "p", `function handle(req) return '{}' end`)
	st := lifecycleStore(t)
	host := NewHost(st, nil, dir, nil)
	if err := host.Boot(ctx); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := host.Boot(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := st.GetPlugin(ctx, "p"); err != nil {
		t.Fatal("missing directory pruned the plugin")
	}
}
