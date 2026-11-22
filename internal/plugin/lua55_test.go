package plugin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLua55Syntax(t *testing.T) {
	rt, err := NewLuaRuntime(`function handle(req) local n <const> = 7 return { n = n } end`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := rt.Call(context.Background(), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"n":7`) {
		t.Fatalf("out = %s", out)
	}
}

func TestLua55NumericTableIsJSONArray(t *testing.T) {
	rt, err := NewLuaRuntime(`function handle(req) return { 1, 2, 3 } end`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := rt.Call(context.Background(), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `[1,2,3]` {
		t.Fatalf("out = %s", out)
	}
}

func TestLua55RequireLocalModule(t *testing.T) {
	dir := t.TempDir()
	module := `return { names = { "One Piece" }, map = { op = "one-piece" } }`
	if err := os.WriteFile(filepath.Join(dir, "genres.lua"), []byte(module), 0o600); err != nil {
		t.Fatal(err)
	}
	rt, err := NewLuaRuntime(`local g = require("genres")
function handle(req) return { first = g.names[1], slug = g.map.op } end`)
	if err != nil {
		t.Fatal(err)
	}
	rt.SetDir(dir)
	out, err := rt.Call(context.Background(), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"first":"One Piece"`) || !strings.Contains(string(out), `"slug":"one-piece"`) {
		t.Fatalf("out = %s", out)
	}
}

func TestLua55RequireRejectsOutsidePath(t *testing.T) {
	rt, err := NewLuaRuntime(`function handle(req) return require("../outside") end`)
	if err != nil {
		t.Fatal(err)
	}
	rt.SetDir(t.TempDir())
	_, err = rt.Call(context.Background(), []byte(`{}`))
	if err == nil || !strings.Contains(err.Error(), "outside plugin directory") {
		t.Fatalf("outside require error = %v", err)
	}
}

func TestLua55SandboxRemovesOS(t *testing.T) {
	rt, err := NewLuaRuntime(`function handle(req) return os.execute("true") end`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rt.Call(context.Background(), []byte(`{}`))
	if err == nil || !strings.Contains(err.Error(), "nil value") {
		t.Fatalf("os sandbox error = %v", err)
	}
}
