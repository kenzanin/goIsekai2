package config

import (
	"sort"
	"strings"
	"sync"
)

var appliers = struct {
	sync.RWMutex
	fn map[string]Applier
}{fn: make(map[string]Applier)}

// Applier applies one changed knob to a live subsystem without a restart.
type Applier func(value string) error

// RegisterLiveApplier registers fn for fullKey, replacing any earlier applier
// for that key. Registration is process-global and safe for concurrent use.
func RegisterLiveApplier(fullKey string, fn Applier) {
	appliers.Lock()
	appliers.fn[fullKey] = fn
	appliers.Unlock()
}

// HasLiveApplier reports whether fullKey can be applied without a restart.
func HasLiveApplier(fullKey string) bool {
	appliers.RLock()
	_, ok := appliers.fn[fullKey]
	appliers.RUnlock()
	return ok
}

// ApplySnapshot calls the registered applier for every resolved knob that has
// one. Keys without an applier, and appliers that fail, are returned sorted so
// the save response can name the knobs that still require a restart.
func ApplySnapshot(s *Snapshot) []string {
	if s == nil {
		return nil
	}
	var restart []string
	for _, res := range s.All() {
		appliers.RLock()
		fn := appliers.fn[res.Entry.Name()]
		appliers.RUnlock()
		if fn == nil {
			restart = append(restart, res.Entry.Name())
			continue
		}
		if err := fn(res.Value.String()); err != nil {
			restart = append(restart, res.Entry.Name())
		}
	}
	sort.Strings(restart)
	return restart
}

// ResetAppliers clears all registered appliers. It is intended for tests so
// one case cannot leak live appliers into another.
func ResetAppliers() {
	appliers.Lock()
	appliers.fn = make(map[string]Applier)
	appliers.Unlock()
}

// RestartNames returns the comma-joined list used in API responses.
func RestartNames(keys []string) string {
	return strings.Join(keys, ",")
}
