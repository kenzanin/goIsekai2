package cache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestGetOrFetchCollapsesConcurrentMisses(t *testing.T) {
	var calls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte("body"))
	}))
	defer upstream.Close()

	store := NewStore(t.TempDir(), "webp", 800, nil)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body, err := store.GetOrFetch(context.Background(), "k", upstream.URL+"/a")
			if err != nil || string(body) != "body" {
				t.Errorf("fetch = %s, %v", body, err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("upstream calls = %d, want 1", calls.Load())
	}
}
