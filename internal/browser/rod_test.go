package browser

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRodLiveCall(t *testing.T) {
	url := os.Getenv("GOISEKAI_CDP_URL")
	if url == "" {
		t.Skip("set GOISEKAI_CDP_URL to a running CDP endpoint")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	r, err := ConnectRod(ctx, url, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if _, err := r.Call("goto", map[string]string{"url": "data:text/html,<title>rod-ok</title>"}); err != nil {
		t.Fatal(err)
	}
	got, err := r.Call("evaluate", map[string]string{"script": "return document.title"})
	if err != nil || got != "rod-ok" {
		t.Fatalf("evaluate = %q, %v", got, err)
	}
	if _, err := r.Call("nope", nil); err == nil {
		t.Fatal("unknown tool should fail")
	}
}
