package plugin

import (
	"context"
	"os"
	"testing"
)

func TestWasmRuntimeCall(t *testing.T) {
	wasBinary, err := os.ReadFile("testdata/echo.wasm")
	if err != nil {
		t.Fatal(err)
	}
	rt, err := NewWasmRuntime(wasBinary)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte(`{"ok":true}`)
	got, err := rt.Call(context.Background(), want)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestWasmRuntimeRejectsMissingHandle(t *testing.T) {
	_, err := NewWasmRuntime([]byte(`(module (memory (export "memory") 1))`))
	if err == nil {
		t.Fatal("missing handle was accepted")
	}
}
