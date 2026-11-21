// Package events provides a typed event bus backed by gookit/event,
// with a single owner goroutine, panic recovery, and recorder support.
package events

import (
	"sync"
	"sync/atomic"

	"github.com/gookit/event"
)

// ConfigChanged is the topic name for configuration change events.
const ConfigChanged = "config.changed"

// PluginLifecycle is the topic name for plugin lifecycle changes.
const PluginLifecycle = "plugin.lifecycle"

// ConfigChangedPayload holds the payload for ConfigChanged events.
type ConfigChangedPayload struct {
	Key        string
	Old        string
	New        string
	Generation uint64
}

// PluginLifecyclePayload reports one plugin's persisted lifecycle state.
type PluginLifecyclePayload struct {
	PluginID string
	State    string
	Error    string
}

type reqType int

const (
	reqSubscribe reqType = iota
	reqPublish
	reqClose
)

type request struct {
	typ     reqType
	topic   string
	payload any
	fn      func(any)
	done    chan struct{}
}

// Bus is a concurrent-safe event bus managed by a single owner goroutine.
type Bus struct {
	mgr       *event.Manager
	reqChan   chan request
	doneCh    chan struct{}
	panics    atomic.Uint64
	closeOnce sync.Once
}

// New creates and starts a new Bus.
func New() *Bus {
	b := &Bus{
		mgr:     event.NewManager("events"),
		reqChan: make(chan request),
		doneCh:  make(chan struct{}),
	}
	go b.loop()
	return b
}

func (b *Bus) loop() {
	for {
		select {
		case <-b.doneCh:
			return
		case req := <-b.reqChan:
			switch req.typ {
			case reqSubscribe:
				// Wrap listener with panic recovery and payload extraction.
				listener := event.ListenerFunc(func(e event.Event) error {
					defer func() {
						if r := recover(); r != nil {
							b.panics.Add(1)
						}
					}()
					val := e.Get("payload")
					req.fn(val)
					return nil
				})
				b.mgr.AddListener(req.topic, listener)
				if req.done != nil {
					close(req.done)
				}
			case reqPublish:
				_, _ = b.mgr.Fire(req.topic, event.M{"payload": req.payload})
				if req.done != nil {
					close(req.done)
				}
			case reqClose:
				if req.done != nil {
					close(req.done)
				}
				return
			}
		}
	}
}

// Subscribe registers a handler function for a given topic.
func (b *Bus) Subscribe(topic string, fn func(any)) {
	done := make(chan struct{})
	select {
	case <-b.doneCh:
		return
	case b.reqChan <- request{
		typ:   reqSubscribe,
		topic: topic,
		fn:    fn,
		done:  done,
	}:
		<-done
	}
}

// Publish sends a payload to all subscribers of a topic, waiting for handlers to complete.
func (b *Bus) Publish(topic string, payload any) {
	done := make(chan struct{})
	select {
	case <-b.doneCh:
		return
	case b.reqChan <- request{
		typ:     reqPublish,
		topic:   topic,
		payload: payload,
		done:    done,
	}:
		<-done
	}
}

// Close stops the owner goroutine and signals closed state.
func (b *Bus) Close() {
	b.closeOnce.Do(func() {
		close(b.doneCh)
		done := make(chan struct{})
		select {
		case b.reqChan <- request{typ: reqClose, done: done}:
			<-done
		default:
		}
	})
}

// Recorder records published payloads for a topic.
type Recorder struct {
	mu    sync.Mutex
	items []any
}

// Record creates a Recorder attached to the bus for the given topic.
func (b *Bus) Record(topic string) *Recorder {
	r := &Recorder{}
	b.Subscribe(topic, func(val any) {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.items = append(r.items, val)
	})
	return r
}

// Items returns a copy of all recorded payloads.
func (r *Recorder) Items() []any {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]any, len(r.items))
	copy(out, r.items)
	return out
}
