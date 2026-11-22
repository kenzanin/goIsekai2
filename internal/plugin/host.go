package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"goisekai2/internal/config"
	"goisekai2/internal/events"
	"goisekai2/internal/plugin/tools"
	"goisekai2/internal/store"
)

// HostTools is the only capability surface a plugin can reach. The store and
// event bus are deliberately absent.
type HostTools interface {
	// Fetch performs an outbound HTTP request for the plugin.
	Fetch(ctx context.Context, url string, headers map[string]string) ([]byte, error)
	// Log records a plugin message.
	Log(level, message string)
	// StorageGet reads a plugin-scoped value.
	StorageGet(key string) (string, bool)
	// StorageSet writes a plugin-scoped value.
	StorageSet(key, value string)
	// Hash hashes a value with the host's crypto helper.
	Hash(value string) string
}

// noopTools is the default until the host-tools capability lands.
type noopTools struct{}

// Fetch reports that the default tool set has no network backend.
func (noopTools) Fetch(context.Context, string, map[string]string) ([]byte, error) {
	return nil, errors.New("plugin: http.fetch is not available")
}

// Log discards plugin log lines until host tools land.
func (noopTools) Log(string, string) {}

// StorageGet reports no stored value.
func (noopTools) StorageGet(string) (string, bool) { return "", false }

// StorageSet discards plugin writes.
func (noopTools) StorageSet(string, string) {}

// Hash returns the value unchanged until a crypto tool is wired.
func (noopTools) Hash(value string) string { return value }

func hasCapability(m *Manifest, want string) bool {
	for _, c := range m.Capabilities {
		if strings.EqualFold(c, want) {
			return true
		}
	}
	return false
}

func snapshotOf(ts *tools.ToolSet) *config.Snapshot {
	if ts == nil {
		return nil
	}
	return ts.Snapshot
}

func dataDirOf(ts *tools.ToolSet, dir string) string {
	if ts != nil && ts.DataDir != "" {
		return ts.DataDir
	}
	return filepath.Dir(dir)
}

// HasCapability reports whether a plugin declares the named capability.
func (h *Host) HasCapability(id, capability string) bool {
	inst, err := h.instance(id)
	if err != nil {
		return false
	}
	for _, c := range inst.Manifest.Capabilities {
		if strings.EqualFold(c, capability) {
			return true
		}
	}
	return false
}

// ImageReferer resolves the referer for a plugin's image fetches. An empty
// result means no Referer header; "none" is the explicit sentinel.
func (h *Host) ImageReferer(id string) string {
	inst, err := h.instance(id)
	if err != nil {
		return ""
	}
	if inst.Manifest.ImageReferer == "none" {
		return ""
	}
	if inst.Manifest.ImageReferer != "" {
		return inst.Manifest.ImageReferer
	}
	u, err := url.Parse(inst.Manifest.Website)
	if err != nil {
		return ""
	}
	return u.Scheme + "://" + u.Host + "/"
}

// Health is the last call outcome; it is never a lifecycle state.
type Health struct {
	OK    bool
	Error string
	At    time.Time
}

// Instance is one loaded plugin with its own call gate.
type Instance struct {
	Manifest  *Manifest
	Runtime   Runtime
	Lifecycle *Lifecycle
	health    Health
	sem       chan struct{}
	mu        sync.RWMutex
}

// Host owns all plugin instances and their lifecycle reconciliation.
type Host struct {
	store     *store.Store
	bus       *events.Bus
	dir       string
	tools     *tools.ToolSet
	mu        sync.RWMutex
	instances map[string]*Instance
}

// NewHost creates a host; nil tools get the safe noop implementation.
func NewHost(st *store.Store, bus *events.Bus, dir string, ts *tools.ToolSet) *Host {
	return &Host{store: st, bus: bus, dir: dir, tools: ts, instances: make(map[string]*Instance)}
}

func loadRuntime(m *Manifest, entry string) (Runtime, error) {
	raw, err := os.ReadFile(entry)
	if err != nil {
		return nil, err
	}
	switch m.Runtime {
	case "lua":
		return NewLuaRuntime(string(raw))
	case "js":
		return NewJSRuntime(string(raw))
	case "wasm":
		return NewWasmRuntime(raw)
	default:
		return nil, fmt.Errorf("plugin: unknown runtime %q", m.Runtime)
	}
}

// Boot discovers plugins, reconciles their persisted state and prunes folders
// that disappeared. A missing directory logs an error and prunes nothing.
func (h *Host) Boot(ctx context.Context) error {
	info, err := os.Stat(h.dir)
	if err != nil || !info.IsDir() {
		slog.Warn("plugin: plugins directory unavailable; skipping discovery and prune", "dir", h.dir)
		return nil
	}
	manifests, err := Discover(h.dir)
	if err != nil {
		return err
	}
	next := make(map[string]*Instance)
	seen := make(map[string]bool)
	for _, entry := range manifests {
		m := entry.Manifest
		folder := entry.ID
		seen[folder] = true
		row, err := h.store.GetPlugin(ctx, folder)
		if err != nil {
			row = &store.PluginRow{Name: m.Name, PluginID: folder, Version: m.Version,
				Capabilities: fmt.Sprint(m.Capabilities), Runtime: m.Runtime}
			_ = h.store.UpsertPlugin(ctx, row)
		}
		initial := StateDisabled
		if s, err := h.store.GetPluginState(ctx, folder, "lifecycle"); err == nil && s != "" {
			initial = s
		} else if row.Enabled == 1 {
			initial = StateEnabled
		}
		rt, err := loadRuntime(m, filepath.Join(h.dir, folder, m.Entry))
		if err != nil {
			slog.Warn("plugin: entry failed to load", "plugin", folder, "error", err)
			initial = StateFailed
			rt = nil
		}
		lc := NewLifecycle(h.store, h.bus, folder, initial)
		ts := &tools.ToolSet{
			Store: h.store, Snapshot: snapshotOf(h.tools), Dir: folder,
			DataDir: dataDirOf(h.tools, h.dir), PluginID: folder,
		}
		if h.tools != nil {
			ts.Browser = h.tools.Browser
		}
		ts.BrowserAllowed = func(id string) bool { return hasCapability(m, "browser") }
		if binder, ok := rt.(interface{ SetTools(*tools.ToolSet) }); ok {
			binder.SetTools(ts)
		}
		next[folder] = &Instance{Manifest: m, Runtime: rt, Lifecycle: lc, sem: make(chan struct{}, 1)}
	}
	rows, err := h.store.AllPlugins(ctx)
	if err == nil {
		for _, row := range rows {
			if !seen[row.PluginID] {
				if err := h.store.DeletePluginRows(ctx, row.PluginID); err != nil {
					slog.Warn("plugin: prune failed", "plugin", row.PluginID, "error", err)
				}
			}
		}
	}
	h.mu.Lock()
	h.instances = next
	h.mu.Unlock()
	return nil
}

func (h *Host) instance(id string) (*Instance, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	inst, ok := h.instances[id]
	if !ok {
		return nil, fmt.Errorf("plugin: unknown plugin %q", id)
	}
	return inst, nil
}

// Call serializes one plugin call per instance and honours the caller deadline
// while queued.
func (h *Host) Call(ctx context.Context, id, method string, params json.RawMessage) ([]byte, error) {
	if err := ValidateMethod(method); err != nil {
		return nil, err
	}
	inst, err := h.instance(id)
	if err != nil {
		return nil, err
	}
	state, err := inst.Lifecycle.State(ctx)
	if err != nil {
		return nil, err
	}
	if len(state) < 7 || state[:7] != "Enabled" {
		return nil, errors.New("plugin disabled")
	}
	select {
	case inst.sem <- struct{}{}:
		defer func() { <-inst.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if inst.Runtime == nil {
		return nil, errors.New("plugin failed")
	}
	_ = inst.Lifecycle.Fire(ctx, TriggerCallStarted)
	envelope, _ := json.Marshal(Envelope{API: HostAPI, Method: method, Params: params})
	out, err := inst.Runtime.Call(ctx, envelope)
	now := time.Now()
	inst.mu.Lock()
	if err != nil {
		inst.health = Health{Error: err.Error(), At: now}
		_ = inst.Lifecycle.Fire(ctx, TriggerCallFailed)
	} else {
		inst.health = Health{OK: true, At: now}
		_ = inst.Lifecycle.Fire(ctx, TriggerCallFinished)
	}
	inst.mu.Unlock()
	return out, err
}

func sliceSort(rows []HostRow) {
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
}

// HostRow is the page-facing view of one plugin instance.
type HostRow struct {
	ID           string
	Name         string
	Version      string
	Runtime      string
	Capabilities []string
	TitleCount   int
	Status       string // active|error|disabled|failed
	Error        string
	Enabled      bool
	HasLogo      bool
	Logo         string
}

// Dir returns the discovered plugins directory.
func (h *Host) Dir() string { return h.dir }

// Rows returns one page row per loaded plugin.
func (h *Host) Rows(ctx context.Context) []HostRow {
	h.mu.RLock()
	defer h.mu.RUnlock()
	rows := make([]HostRow, 0, len(h.instances))
	for id, inst := range h.instances {
		state, _ := inst.Lifecycle.State(ctx)
		health := h.Health(id)
		row := HostRow{
			ID: id, Name: inst.Manifest.Name, Version: inst.Manifest.Version,
			Runtime: inst.Manifest.Runtime, Capabilities: inst.Manifest.Capabilities,
			Enabled: len(state) >= 7 && state[:7] == "Enabled",
			Logo:    inst.Manifest.Logo,
		}
		switch {
		case state == StateFailed:
			row.Status = "failed"
		case !row.Enabled:
			row.Status = "disabled"
		case !health.OK && health.Error != "":
			row.Status, row.Error = "error", health.Error
		default:
			row.Status = "active"
		}
		if inst.Manifest.Logo != "" {
			if _, err := os.Stat(filepath.Join(h.dir, id, inst.Manifest.Logo)); err == nil {
				row.HasLogo = true
			}
		}
		count, _ := h.store.CountSavedByPlugin(ctx, id)
		row.TitleCount = count
		rows = append(rows, row)
	}
	sliceSort(rows)
	return rows
}

// SetEnabled changes one plugin's lifecycle state, waiting for a busy call.
func (h *Host) SetEnabled(ctx context.Context, id string, enabled bool) error {
	inst, err := h.instance(id)
	if err != nil {
		return err
	}
	if !enabled {
		select {
		case inst.sem <- struct{}{}:
			defer func() { <-inst.sem }()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	trigger := TriggerDisable
	if enabled {
		trigger = TriggerEnable
	}
	if err := inst.Lifecycle.Fire(ctx, trigger); err != nil {
		return err
	}
	return h.store.SetPluginEnabled(ctx, id, enabled)
}

// Reload rescans plugins with the boot pipeline and keeps persisted states.
func (h *Host) Reload(ctx context.Context) error {
	h.mu.RLock()
	instances := make([]*Instance, 0, len(h.instances))
	for _, inst := range h.instances {
		instances = append(instances, inst)
	}
	h.mu.RUnlock()
	for _, inst := range instances {
		inst.sem <- struct{}{}
		<-inst.sem
	}
	return h.Boot(ctx)
}

// Health returns the last call outcome for one plugin.
func (h *Host) Health(id string) Health {
	inst, err := h.instance(id)
	if err != nil {
		return Health{}
	}
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	return inst.health
}
