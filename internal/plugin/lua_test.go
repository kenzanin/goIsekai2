package plugin

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestLuaRuntimeCall(t *testing.T) {
	rt, err := NewLuaRuntime(`function handle(req) return '{"ok":true,"data":' .. req .. '}' end`)
	if err != nil {
		t.Fatal(err)
	}
	out, err := rt.Call(context.Background(), []byte(`{"n":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"ok":true,"data":{"n":1}}` {
		t.Fatalf("out = %s", out)
	}
}

func TestLuaRuntimeRejectsLuaJITModules(t *testing.T) {
	rt, err := NewLuaRuntime(`function handle(req) local f = require("ffi"); return "x" end`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rt.Call(context.Background(), []byte(`{}`))
	if err == nil || !strings.Contains(err.Error(), unsupportedModule) {
		t.Fatalf("ffi error = %v", err)
	}
}

func TestLuaRuntimeBudget(t *testing.T) {
	rt, err := NewLuaRuntime(`function handle(req) while true do end end`)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := rt.Call(ctx, []byte(`{}`)); err == nil {
		t.Fatal("runaway loop was not stopped")
	}
}
