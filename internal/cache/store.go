package cache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/webp"
)

type flight struct {
	done chan struct{}
	data []byte
	err  error
}

// Store writes cache files atomically and collapses concurrent misses.
type Store struct {
	Root     string
	Format   string
	CoverMax int
	Pool     *Pool
	Client   *http.Client
	mu       sync.Mutex
	flights  map[string]*flight
	negative map[string]time.Time
	seen     sync.Map
}

// NewStore creates a cache store.
func NewStore(root, format string, coverMax int, pool *Pool) *Store {
	if format == "" {
		format = "webp"
	}
	if coverMax <= 0 {
		coverMax = 800
	}
	return &Store{
		Root: root, Format: format, CoverMax: coverMax, Pool: pool,
		Client:  &http.Client{Timeout: 30 * time.Second},
		flights: make(map[string]*flight), negative: make(map[string]time.Time),
	}
}

// Exists reports whether key was fetched in this process.
func (s *Store) Exists(key string) bool {
	_, ok := s.seen.Load(key)
	return ok
}

// Hash returns the cache key for an upstream URL.
func Hash(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	return hex.EncodeToString(sum[:])[:16]
}

// ponytail: webp transcodes in-process; avif/jxl currently pass bytes through
// with the configured extension because their wasm encoders are not wired yet.
func extension(format string) string {
	switch format {
	case "avif":
		return "avif"
	case "jxl":
		return "jxl"
	case "none":
		return "bin"
	default:
		return "webp"
	}
}

func contentType(ext string) string {
	switch ext {
	case "webp":
		return "image/webp"
	case "avif":
		return "image/avif"
	case "jxl":
		return "image/jxl"
	default:
		return "application/octet-stream"
	}
}

// Put stores data at path, transcoding and normalizing covers, then registers
// the file with the LRU pool.
func (s *Store) Put(path string, data []byte, cover bool) (string, float64, error) {
	var ratio float64
	out := data
	if cover {
		decoded, r, err := ResizeCover(data, s.CoverMax)
		if err == nil {
			out, ratio = decoded, r
		}
	} else if s.Format == "webp" {
		if decoded, _, err := image.Decode(bytes.NewReader(data)); err == nil {
			var buf bytes.Buffer
			if err := webp.Encode(&buf, decoded); err == nil {
				out = buf.Bytes()
			}
		}
	}
	ext := extension(s.Format)
	path = strings.TrimSuffix(path, filepath.Ext(path)) + "." + ext
	if cover && ratio > 0 {
		stem := strings.TrimSuffix(path, "."+ext)
		path = fmt.Sprintf("%s-%d.%s", stem, int(math.Round(ratio*10000)), ext)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", 0, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".cache-*")
	if err != nil {
		return "", 0, err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := io.Copy(tmp, bytes.NewReader(out)); err != nil {
		tmp.Close()
		return "", 0, err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return "", 0, err
	}
	if err := tmp.Close(); err != nil {
		return "", 0, err
	}
	if err := os.Rename(name, path); err != nil {
		return "", 0, err
	}
	if s.Pool != nil {
		s.Pool.Track(path)
		_ = s.Pool.Evict()
	}
	return path, ratio, nil
}

// GetOrFetch returns a cached body or fetches it once for concurrent callers.
func (s *Store) GetOrFetch(ctx context.Context, key, rawURL string) ([]byte, error) {
	return s.GetOrFetchWithReferer(ctx, key, rawURL, "")
}

// GetOrFetchWithReferer behaves like GetOrFetch, sending referer when set.
func (s *Store) GetOrFetchWithReferer(ctx context.Context, key, rawURL, referer string) ([]byte, error) {
	s.mu.Lock()
	if until, ok := s.negative[key]; ok && time.Now().Before(until) {
		s.mu.Unlock()
		return nil, os.ErrNotExist
	}
	if f, ok := s.flights[key]; ok {
		s.mu.Unlock()
		select {
		case <-f.done:
			return f.data, f.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	f := &flight{done: make(chan struct{})}
	s.flights[key] = f
	s.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err == nil && referer != "" {
		req.Header.Set("Referer", referer)
	}
	if err == nil {
		var resp *http.Response
		resp, err = s.Client.Do(req)
		if err == nil {
			f.data, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err == nil && resp.StatusCode/100 != 2 {
				err = os.ErrNotExist
			}
		}
	}
	if err != nil {
		s.mu.Lock()
		s.negative[key] = time.Now().Add(2 * time.Second)
		s.mu.Unlock()
	}
	f.err = err
	if err == nil {
		s.seen.Store(key, true)
	}
	close(f.done)
	s.mu.Lock()
	delete(s.flights, key)
	s.mu.Unlock()
	return f.data, f.err
}

// Ratio reads a cover sidecar written by WriteRatio.
func Ratio(path string) float64 {
	return ReadRatio(path)
}
