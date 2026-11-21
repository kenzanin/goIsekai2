package plugin

import (
	"context"
	"os"
	"testing"
)

// TestWasmBindingsEcho guards the wasm binding after the Lua/JS host-tool
// promotions: echo.wasm must still link and echo with no new imports.
func TestWasmBindingsEcho(t *testing.T) {
	bin, err := os.ReadFile("testdata/echo.wasm")
	if err != nil {
		t.Fatal(err)
	}
	rt, err := NewWasmRuntime(bin)
	if err != nil {
		t.Fatalf("link: %v", err)
	}
	want := []byte(`{"ping":1}`)
	got, err := rt.Call(context.Background(), want)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("got %s, want %s", got, want)
	}
}
