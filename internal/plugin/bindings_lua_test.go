package plugin

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"goisekai2/internal/plugin/tools"
	"goisekai2/internal/store"
)

func TestLuaHostBindings(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer st.Close()

	const fixture = `
local json = require "json"
function handle(req)
  host.log.info("binding probe")
  host.storage.set("cursor", "42")
  local got, gerr = host.storage.get("cursor")
  local absent, aerr = host.storage.get("absent")
  local call = json.decode(host.call("now", "[]"))
  local bad = json.decode(host.call("nope", "[]"))
  return json.encode({
    now = host.now() > 0,
    got = got,
    get_err = gerr,
    absent = absent,
    absent_err = aerr ~= nil,
    call_ok = call.ok,
    call_data = call.data,
    bad_ok = bad.ok,
    bad_err = bad.error,
  })
end`

	rt, err := NewLuaRuntime(fixture)
	if err != nil {
		t.Fatal(err)
	}
	rt.SetTools(&tools.ToolSet{Store: st, PluginID: "probe"})
	out, err := rt.Call(context.Background(), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("out = %s: %v", out, err)
	}
	if m["now"] != true {
		t.Errorf("now = %v", m["now"])
	}
	if m["got"] != "42" {
		t.Errorf("got = %v, want 42", m["got"])
	}
	if m["get_err"] != nil {
		t.Errorf("get_err = %v", m["get_err"])
	}
	if m["absent"] != nil {
		t.Errorf("absent = %v", m["absent"])
	}
	if m["absent_err"] != true {
		t.Errorf("absent_err = %v, want true", m["absent_err"])
	}
	if m["call_ok"] != true {
		t.Errorf("call_ok = %v", m["call_ok"])
	}
	if d, ok := m["call_data"].(float64); !ok || d <= 0 {
		t.Errorf("call_data = %v, want positive number", m["call_data"])
	}
	if m["bad_ok"] != false {
		t.Errorf("bad_ok = %v, want false", m["bad_ok"])
	}
	if s, ok := m["bad_err"].(string); !ok || !strings.Contains(s, "unknown tool") {
		t.Errorf("bad_err = %v, want unknown-tool error", m["bad_err"])
	}
}
