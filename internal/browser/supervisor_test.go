package browser

import (
	"context"
	"net/http"
	"net/http/httptest"
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
