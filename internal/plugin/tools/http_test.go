package tools

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchNon2xxIsDataAndFetchBodyErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Kind", "gone")
		w.WriteHeader(http.StatusGone)
		_, _ = w.Write([]byte("moved away"))
	}))
	defer srv.Close()

	ts := &ToolSet{PluginID: "probe"}
	res, err := ts.Fetch(srv.URL, map[string]any{"profile": "nope_999"}) // unknown profile must warn, not fail
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if status := res["status"].(int); status != http.StatusGone {
		t.Fatalf("status = %d, want 410", status)
	}
	if body := res["body"].(string); body != "moved away" {
		t.Fatalf("body = %q", body)
	}
	hdrs := res["headers"].(map[string]string)
	if hdrs["x-kind"] != "gone" {
		t.Fatalf("headers = %v, want lowercased x-kind", hdrs)
	}

	if _, err := ts.FetchBody(srv.URL, nil); err == nil || err.Error() != "upstream status 410" {
		t.Fatalf("FetchBody err = %v, want 'upstream status 410'", err)
	}
}

func TestFetchPostHeadersBodyAndTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("X-Token"); got != "abc" {
			t.Errorf("X-Token = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "text/plain" {
			t.Errorf("Content-Type = %q", got)
		}
		buf := make([]byte, 64)
		n, _ := r.Body.Read(buf)
		if string(buf[:n]) != "payload" {
			t.Errorf("body = %q", buf[:n])
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	ts := &ToolSet{PluginID: "probe"}
	res, err := ts.Fetch(srv.URL, map[string]any{
		"method":     "POST",
		"body":       "payload",
		"headers":    map[string]any{"X-Token": "abc", "Content-Type": "text/plain"},
		"timeout_ms": 5000,
	})
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if res["body"].(string) != "ok" {
		t.Fatalf("body = %v", res["body"])
	}
}

func TestGetJSONPostJSONAndBareHeaderMap(t *testing.T) {
	type payload struct {
		OK bool `json:"ok"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			if got := r.Header.Get("Accept"); got != "application/json" {
				t.Errorf("Accept = %q, want bare options map used as headers", got)
			}
			_, _ = w.Write([]byte(`{"ok":true,"n":3}`))
		case http.MethodPost:
			if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
				t.Errorf("Content-Type = %q, want explicit value kept", got)
			}
			_, _ = w.Write([]byte(`{"echo":"sent"}`))
		}
	}))
	defer srv.Close()

	ts := &ToolSet{PluginID: "probe"}
	got, err := ts.GetJSON(srv.URL, map[string]any{"Accept": "application/json"}) // no known key => headers
	if err != nil {
		t.Fatalf("GetJSON: %v", err)
	}
	m, ok := got.(map[string]any)
	if !ok || m["ok"] != true || m["n"] != float64(3) {
		t.Fatalf("GetJSON = %#v", got)
	}

	posted, err := ts.PostJSON(srv.URL, payload{OK: true}, map[string]any{"headers": map[string]string{"content-type": "application/x-www-form-urlencoded"}})
	if err != nil {
		t.Fatalf("PostJSON: %v", err)
	}
	if m := posted.(map[string]any); m["echo"] != "sent" {
		t.Fatalf("PostJSON = %#v", posted)
	}
}
