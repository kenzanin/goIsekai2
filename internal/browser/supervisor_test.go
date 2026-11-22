package browser

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSupervisorReusesForeignInstance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	s := &Supervisor{URL: srv.URL, Binary: "/does/not/exist", Port: 1}
	owned, err := s.Start(context.Background())
	if err != nil || owned {
		t.Fatalf("start = %v, %v", owned, err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if srv == nil {
		t.Fatal("foreign server was closed")
	}
}

func TestSupervisorSpawnsAndStopsOwnedChild(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "fake-browser")
	script := "#!/bin/sh\nport=$5\nexec python3 -m http.server \"$port\" --bind 127.0.0.1\n"
	if err := os.WriteFile(bin, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	port := 19422
	s := &Supervisor{
		URL:    fmt.Sprintf("http://127.0.0.1:%d/json/version", port),
		Binary: bin, Port: port,
	}
	owned, err := s.Start(context.Background())
	if err != nil || !owned {
		t.Fatalf("start = %v, %v", owned, err)
	}
	if !s.Probe(context.Background()) {
		t.Fatal("child did not become healthy")
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
