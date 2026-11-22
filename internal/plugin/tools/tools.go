// Package tools is the host tool surface handed to plugin runtimes: clock,
// log, plugin-scoped persistence and resolved-config lookup.
package tools

import (
	"context"
	"log/slog"
	"time"

	"goisekai2/internal/browser"
	"goisekai2/internal/config"
	"goisekai2/internal/store"
)

// ToolSet bundles the host capabilities for exactly one plugin: Store and
// Snapshot are shared daemon state, PluginID scopes storage rows and log
// lines, Dir is the plugin's own directory.
type ToolSet struct {
	Store          *store.Store
	Snapshot       *config.Snapshot
	Dir            string
	DataDir        string
	PluginID       string
	Log            *slog.Logger
	Browser        browser.Browser
	BrowserAllowed func(string) bool
}

// Now returns the current Unix time in seconds.
func (t *ToolSet) Now() int64 { return time.Now().Unix() }

// LogInfo records one info-level line attributed to the plugin.
func (t *ToolSet) LogInfo(msg string) { t.log(slog.LevelInfo, msg) }

// LogError records one error-level line attributed to the plugin.
func (t *ToolSet) LogError(msg string) { t.log(slog.LevelError, msg) }

// StorageSet upserts a persistent value scoped to the plugin id.
func (t *ToolSet) StorageSet(key, value string) error {
	return t.Store.PutPluginState(context.Background(), t.PluginID, key, value)
}

// StorageGet reads a value stored by StorageSet; a missing key is an error.
func (t *ToolSet) StorageGet(key string) (string, error) {
	return t.Store.GetPluginState(context.Background(), t.PluginID, key)
}

// ConfigGet returns the resolved value of a dotted knob name and whether the
// knob exists in the registry.
func (t *ToolSet) ConfigGet(key string) (string, bool) {
	res, ok := t.Snapshot.Get(key)
	if !ok {
		return "", false
	}
	return res.Value.String(), ok
}

// log emits through the daemon logger, falling back to the default one.
func (t *ToolSet) log(level slog.Level, msg string) {
	logger := t.Log
	if logger == nil {
		logger = slog.Default()
	}
	logger.Log(context.Background(), level, msg, "plugin", t.PluginID)
}
