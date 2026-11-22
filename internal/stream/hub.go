// Package stream fans out log lines to WebSocket subscribers without blocking
// publishers.
package stream

import (
	"context"
	"log/slog"
	"sync"
)

// Hub keeps a bounded replay ring and lossy per-subscriber channels.
type Hub struct {
	mu    sync.Mutex
	cap   int
	lines []string
	subs  map[chan string]struct{}
}

// New creates a hub replaying at most cap lines.
func New(cap int) *Hub {
	if cap <= 0 {
		cap = 200
	}
	return &Hub{cap: cap, lines: make([]string, 0, cap), subs: make(map[chan string]struct{})}
}

// Publish records one line and offers it to every subscriber, dropping the
// oldest queued line when a subscriber is slow.
func (h *Hub) Publish(line string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lines = append(h.lines, line)
	if len(h.lines) > h.cap {
		h.lines = append(h.lines[:0], h.lines[len(h.lines)-h.cap:]...)
	}
	for ch := range h.subs {
		select {
		case ch <- line:
		default:
			select {
			case <-ch:
			default:
			}
			ch <- line
		}
	}
}

// Lines returns the replay buffer in chronological order.
func (h *Hub) Lines() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.lines...)
}

// Subscribe returns a line channel and an unsubscribe function.
func (h *Hub) Subscribe() (<-chan string, func()) {
	ch := make(chan string, 64)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		delete(h.subs, ch)
		h.mu.Unlock()
	}
}

// Handler mirrors slog records into the hub while keeping the next handler.
type Handler struct {
	next slog.Handler
	hub  *Hub
}

// NewHandler wraps next so each handled record is also published.
func NewHandler(next slog.Handler, hub *Hub) *Handler {
	return &Handler{next: next, hub: hub}
}

// Enabled delegates to the wrapped handler.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

// Handle records the line then delegates to the wrapped handler.
func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	h.hub.Publish(record.Time.Format("15:04:05") + " " + record.Level.String() + " " + record.Message)
	return h.next.Handle(ctx, record)
}

// WithAttrs delegates to the wrapped handler.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{next: h.next.WithAttrs(attrs), hub: h.hub}
}

// WithGroup delegates to the wrapped handler.
func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{next: h.next.WithGroup(name), hub: h.hub}
}
