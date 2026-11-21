package tools

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"goisekai2/internal/config"
	"goisekai2/internal/store"
)

func newTestSet(t *testing.T) *ToolSet {
	t.Helper()
	st, err := store.Open(t.TempDir() + "/lib.db")
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	snap, err := config.Resolve(config.Knobs(), nil, nil, nil)
	if err != nil {
		t.Fatalf("config.Resolve: %v", err)
	}
	return &ToolSet{Store: st, Snapshot: snap, PluginID: "probe"}
}

func TestStorageScopedPerPlugin(t *testing.T) {
	ts := newTestSet(t)
	if err := ts.StorageSet("last_sync", "42"); err != nil {
		t.Fatalf("StorageSet: %v", err)
	}
	if got, err := ts.StorageGet("last_sync"); err != nil || got != "42" {
		t.Fatalf("StorageGet after set = %q,%v; want 42,nil", got, err)
	}
	if err := ts.StorageSet("last_sync", "43"); err != nil {
		t.Fatalf("StorageSet overwrite: %v", err)
	}
	if got, _ := ts.StorageGet("last_sync"); got != "43" {
		t.Errorf("StorageGet after overwrite = %q; want 43", got)
	}
	other := *ts
	other.PluginID = "other"
	if _, err := other.StorageGet("last_sync"); err == nil {
		t.Errorf("other plugin sees probe's state; want scoped miss")
	}
}

func TestConfigGet(t *testing.T) {
	ts := newTestSet(t)
	if v, ok := ts.ConfigGet("server.port"); !ok || v != "8090" {
		t.Errorf(`ConfigGet("server.port") = %q,%v; want "8090",true`, v, ok)
	}
	if v, ok := ts.ConfigGet("nope.nope"); ok || v != "" {
		t.Errorf(`ConfigGet("nope.nope") = %q,%v; want "",false`, v, ok)
	}
}

func TestLogAndClock(t *testing.T) {
	var buf bytes.Buffer
	ts := &ToolSet{PluginID: "probe", Log: nil} // nil Log must not panic
	ts.LogInfo("dropped")
	ts.Log = slog.New(slog.NewTextHandler(&buf, nil))
	ts.LogError("boom")
	for _, want := range []string{"level=ERROR", "msg=boom", "plugin=probe"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("log output missing %q: %s", want, buf.String())
		}
	}
	if got := ts.Now(); got < time.Now().Unix()-2 || got > time.Now().Unix()+2 {
		t.Errorf("Now() = %d; want within 2s of %d", got, time.Now().Unix())
	}
}
