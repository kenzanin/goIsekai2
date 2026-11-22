package cache

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type entry struct {
	path string
	at   time.Time
	size int64
}

// Pool tracks cache files by recency and enforces one byte budget.
type Pool struct {
	Root      string
	MaxBytes  int64
	mu        sync.Mutex
	entries   map[string]entry
	totalSize int64
}

// NewPool creates an empty pool for root.
func NewPool(root string, maxBytes int64) *Pool {
	if maxBytes <= 0 {
		maxBytes = 2048 << 20
	}
	return &Pool{Root: root, MaxBytes: maxBytes, entries: make(map[string]entry)}
}

// Rebuild scans the cache tree and restores recency from file mtimes.
func (p *Pool) Rebuild() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.entries = make(map[string]entry)
	p.totalSize = 0
	return filepath.Walk(p.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() || filepath.Ext(path) == ".json" {
			return nil
		}
		p.add(path, info.Size(), info.ModTime())
		return nil
	})
}

func (p *Pool) add(path string, size int64, at time.Time) {
	if old, ok := p.entries[path]; ok {
		p.totalSize -= old.size
	}
	p.entries[path] = entry{path: path, at: at, size: size}
	p.totalSize += size
}

// Touch records a read or write and refreshes recency.
func (p *Pool) Touch(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.add(path, info.Size(), time.Now())
}

// Track registers a path after a successful write.
func (p *Pool) Track(path string) { p.Touch(path) }

// Evict removes oldest entries until the pool fits its budget.
func (p *Pool) Evict() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.totalSize <= p.MaxBytes {
		return nil
	}
	list := make([]entry, 0, len(p.entries))
	for _, e := range p.entries {
		list = append(list, e)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].at.Before(list[j].at) })
	for _, e := range list {
		if p.totalSize <= p.MaxBytes {
			break
		}
		for _, path := range []string{e.path, sidecarPath(e.path)} {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
		delete(p.entries, e.path)
		p.totalSize -= e.size
	}
	return nil
}

func sidecarPath(path string) string { return path + ".json" }
