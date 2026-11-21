package plugin

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestJSRuntimeCall(t *testing.T) {
	r, err := NewJSRuntime(`function handle(req){return {echo:req}}`)
	if err != nil {
		t.Fatalf("NewJSRuntime: %v", err)
	}
	got, err := r.Call(context.Background(), []byte(`{"a":1}`))
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if want := `{"echo":{"a":1}}`; string(got) != want {
		t.Fatalf("Call = %s, want %s", got, want)
	}
}

func TestJSRuntimeNewRejectsBrokenSource(t *testing.T) {
	if _, err := NewJSRuntime(`function handle(){`); err == nil {
		t.Fatal("NewJSRuntime with broken source, want error")
	}
}

func TestJSRuntimeCallLoopTimesOut(t *testing.T) {
	r, err := NewJSRuntime(`function handle(){while(true){}}`)
	if err != nil {
		t.Fatalf("NewJSRuntime: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = r.Call(ctx, []byte(`{}`))
	if err == nil {
		t.Fatal("Call on endless loop, want timeout error")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("Call took %s, timeout did not fire", elapsed)
	}
}

func TestJSRuntimeCallPromiseRejected(t *testing.T) {
	r, err := NewJSRuntime(`function handle(){return Promise.resolve(1)}`)
	if err != nil {
		t.Fatalf("NewJSRuntime: %v", err)
	}
	_, err = r.Call(context.Background(), []byte(`{}`))
	if err == nil || !strings.Contains(err.Error(), "promise return unsupported") {
		t.Fatalf("Call with Promise result = %v, want \"promise return unsupported\"", err)
	}
}
