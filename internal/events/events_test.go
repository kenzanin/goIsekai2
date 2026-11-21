package events

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBus_FIFOAndPublish(t *testing.T) {
	b := New()
	defer b.Close()

	var order []int
	var mu sync.Mutex

	b.Subscribe("test.topic", func(v any) {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, v.(int))
	})

	b.Publish("test.topic", 1)
	b.Publish("test.topic", 2)
	b.Publish("test.topic", 3)

	// Give a brief moment or since Publish is synchronous, check immediately.
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []int{1, 2, 3}, order)
}

func TestBus_PanicRecoveryDoesNotStopLaterEvents(t *testing.T) {
	b := New()
	defer b.Close()

	var order []string
	var mu sync.Mutex

	// Panicking subscriber
	b.Subscribe("test.panic", func(v any) {
		if v == "panic" {
			panic("oops")
		}
		mu.Lock()
		defer mu.Unlock()
		order = append(order, v.(string))
	})

	b.Publish("test.panic", "first")
	b.Publish("test.panic", "panic")
	b.Publish("test.panic", "third")

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []string{"first", "third"}, order)
	assert.Equal(t, uint64(1), b.panics.Load())
}

func TestBus_CloseLeavesNoGoroutine(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()

	b := New()
	b.Subscribe("some.topic", func(any) {})
	b.Close()

	// Wait briefly for goroutine exit
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > initialGoroutines && time.Now().Before(deadline) {
		time.Sleep(10 * ms)
	}

	// Allow small delta or assert exact if stable
	assert.LessOrEqual(t, runtime.NumGoroutine(), initialGoroutines+1)
}

func TestBus_RecordConfigChangedPayload(t *testing.T) {
	b := New()
	defer b.Close()

	rec := b.Record(ConfigChanged)

	p := ConfigChangedPayload{
		Key:        "theme",
		Old:        "light",
		New:        "dark",
		Generation: 42,
	}

	b.Publish(ConfigChanged, p)

	items := rec.Items()
	require.Len(t, items, 1)

	got, ok := items[0].(ConfigChangedPayload)
	require.True(t, ok)
	assert.Equal(t, p, got)
}

const ms = time.Millisecond
