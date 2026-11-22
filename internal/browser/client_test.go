package browser

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientCallAndUnknownTool(t *testing.T) {
	tool := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if tool == "bad" {
			fmt.Fprint(w, `{"error":{"code":-32601,"message":"unknown tool"}}`)
			return
		}
		fmt.Fprint(w, `{"result":{"content":[{"text":"OK"}]}}`)
	}))
	defer srv.Close()
	c := &Client{URL: srv.URL, Timeout: time.Second}
	got, err := c.Call("evaluate", map[string]string{"script": "return 1"})
	if err != nil || got != "OK" {
		t.Fatalf("call = %q, %v", got, err)
	}
	tool = "bad"
	if _, err := c.Call("bad", nil); !errors.Is(err, ErrUnknownTool) {
		t.Fatalf("unknown tool error = %v", err)
	}
}

func TestClientSessionExpiryRetriesOnce(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			fmt.Fprint(w, `{"error":{"code":-32000,"message":"session expired"}}`)
			return
		}
		fmt.Fprint(w, `{"result":{"content":[{"text":"RETRY_OK"}]}}`)
	}))
	defer srv.Close()
	c := &Client{URL: srv.URL, Timeout: time.Second}
	got, err := c.Call("evaluate", nil)
	if err != nil || got != "RETRY_OK" || calls != 2 {
		t.Fatalf("retry = %q, %v, calls=%d", got, err, calls)
	}
}

func TestClientLoopbackIgnoresProxy(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"result":{"content":[{"text":"DIRECT"}]}}`)
	}))
	defer srv.Close()
	c := &Client{URL: srv.URL, Timeout: time.Second}
	got, err := c.Call("goto", nil)
	if err != nil || got != "DIRECT" {
		t.Fatalf("loopback = %q, %v", got, err)
	}
}
