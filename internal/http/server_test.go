package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"goisekai2/internal/config"
	"goisekai2/internal/events"
)

func testIni(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.ini")
	if err := os.WriteFile(path, []byte("[paths]\ndata_dir = "+dir+"\nlibrary = "+filepath.Join(dir, "lib.db")+"\n\n[server]\n; listen port\nport = 8090\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRoutesOnlyIncludeImplementedPages(t *testing.T) {
	s := New(config.Knobs(), testIni(t), nil)
	if got := s.Pages(); len(got) != 1 || got[0].Path != "/settings" {
		t.Fatalf("pages = %v, want only /settings", got)
	}

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/settings", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `href="/settings" aria-current="page"`) {
		t.Fatalf("active Settings link missing: %s", body)
	}
	for _, dead := range []string{`href="/library"`, `href="/plugins"`} {
		if strings.Contains(body, dead) {
			t.Fatalf("dead nav link %q rendered", dead)
		}
	}
}

func TestGetConfigReflectsFileEdits(t *testing.T) {
	path := testIni(t)
	s := New(config.Knobs(), path, nil)
	get := func() []config.Projection {
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/config", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d", rec.Code)
		}
		var out []config.Projection
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	first := get()
	if len(first) == 0 || first[0].Section != "server" || first[0].Value != "8090" || first[0].Source != "ini" {
		t.Fatalf("projection = %+v", first)
	}
	if err := os.WriteFile(path, []byte("[server]\nport = 7777\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	second := get()
	if second[0].Value != "7777" {
		t.Fatalf("stale projection: %+v", second[0])
	}
}

func TestPostConfigMergeSaveAndReject(t *testing.T) {
	path := testIni(t)
	s := New(config.Knobs(), path, nil)
	post := func(body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/config", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		s.Handler().ServeHTTP(rec, req)
		return rec
	}

	rec := post(`{"server.port":"9000"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"written":1`) {
		t.Fatalf("success = %d %s", rec.Code, rec.Body.String())
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "; listen port") || !strings.Contains(string(after), "port = 9000") {
		t.Fatalf("comment or value lost: %q", after)
	}

	before, _ := os.ReadFile(path)
	rec = post(`{"server.port":"http"}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "server.port") {
		t.Fatalf("reject = %d %s", rec.Code, rec.Body.String())
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(before, got) {
		t.Fatalf("rejected save changed the file: %q -> %q", before, got)
	}
}

func TestServeAndShutdown(t *testing.T) {
	s := New(config.Knobs(), testIni(t), nil)
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- s.Serve(l) }()

	resp, err := http.Get("http://" + l.Addr().String() + "/settings")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil && err != http.ErrServerClosed {
		t.Fatalf("serve: %v", err)
	}
}

func TestPostConfigPublishesAfterWriteAndAppliesLive(t *testing.T) {
	config.ResetAppliers()
	t.Cleanup(config.ResetAppliers)
	levelVar := new(slog.LevelVar)
	config.RegisterLiveApplier("log.level", func(v string) error {
		var level slog.Level
		if err := level.UnmarshalText([]byte(v)); err != nil {
			return err
		}
		levelVar.Set(level)
		return nil
	})

	path := filepath.Join(t.TempDir(), "cfg.ini")
	if err := os.WriteFile(path, []byte("[log]\nlevel = info\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	bus := events.New()
	t.Cleanup(bus.Close)
	rec := bus.Record(events.ConfigChanged)
	s := New(config.Knobs(), path, bus)

	post := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/config", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, req)
		return rec
	}

	recHTTP := post(`{"log.level":"debug"}`)
	if recHTTP.Code != http.StatusOK {
		t.Fatalf("status = %d %s", recHTTP.Code, recHTTP.Body.String())
	}
	onDisk, _ := os.ReadFile(path)
	if !strings.Contains(string(onDisk), "level = debug") {
		t.Fatalf("file not written before event: %q", onDisk)
	}
	items := rec.Items()
	if len(items) != 1 {
		t.Fatalf("events = %d, want 1", len(items))
	}
	payload, ok := items[0].(events.ConfigChangedPayload)
	if !ok || payload.Key != "log.level" || payload.Old != "info" || payload.New != "debug" {
		t.Fatalf("payload = %+v", items[0])
	}
	if levelVar.Level() != slog.LevelDebug {
		t.Fatalf("level = %v, want debug", levelVar.Level())
	}
	if strings.Contains(recHTTP.Body.String(), `"log.level"`) {
		t.Fatalf("applied key listed as restart-required: %s", recHTTP.Body.String())
	}

	post(`{"log.level":"loud"}`)
	if n := len(rec.Items()); n != 1 {
		t.Fatalf("rejected save published %d events, want 0 more", n-1)
	}
}

func TestAssetsAreServedLocally(t *testing.T) {
	s := New(config.Knobs(), testIni(t), nil)

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/settings", nil))
	for _, attr := range []string{`src="http`, `href="http`, `src="//`, `href="//`} {
		if strings.Contains(rec.Body.String(), attr) {
			t.Fatalf("page references an external asset (%s): %s", attr, rec.Body.String())
		}
	}
	for _, path := range []string{"/assets/vendor/htmx.min.js", "/assets/vendor/alpine.min.js"} {
		assetRec := httptest.NewRecorder()
		s.Handler().ServeHTTP(assetRec, httptest.NewRequest(http.MethodGet, path, nil))
		if assetRec.Code != http.StatusOK || assetRec.Body.Len() == 0 {
			t.Fatalf("asset %s = %d (%d bytes)", path, assetRec.Code, assetRec.Body.Len())
		}
	}
}

func TestPostConfigHTMXFragment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.ini")
	if err := os.WriteFile(path, []byte("[log]\nlevel = info\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := New(config.Knobs(), path, nil)

	req := httptest.NewRequest(http.MethodPost, "/config", strings.NewReader("log.level=debug"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Saved 1 change") {
		t.Fatalf("status fragment missing: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"written"`) {
		t.Fatalf("htmx response should be HTML, got JSON: %s", rec.Body.String())
	}
	onDisk, _ := os.ReadFile(path)
	if !strings.Contains(string(onDisk), "level = debug") {
		t.Fatalf("form values not saved: %q", onDisk)
	}
}
