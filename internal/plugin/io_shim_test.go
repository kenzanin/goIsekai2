package plugin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goisekai2/internal/plugin/tools"
)

func TestLuaConfinedIOShim(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("alpha\nbeta\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	const source = `function handle(req)
  local body = io.read("notes.txt")
  local lines = {}
  for line in io.lines("notes.txt") do lines[#lines + 1] = line end
  local escaped, err = io.read("../outside.txt")
  return { body = body, lines = lines, escaped = escaped ~= nil, err = err }
end`
	rt, err := NewLuaRuntimeWithTools(source, &tools.ToolSet{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	out, err := rt.Call(context.Background(), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	assertIOShim(t, out)
}

func TestJSConfinedIOShim(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("alpha\nbeta\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	const source = `function handle(req) {
  var body = host.io.read("notes.txt");
  var lines = [];
  var it = host.io.lines("notes.txt");
  for (var line = it(); line !== undefined; line = it()) lines.push(line);
  var escaped = false, err = "";
  try { host.io.read("../outside.txt"); escaped = true; }
  catch (e) { err = String(e); }
  return { body: body, lines: lines, escaped: escaped, err: err };
}`
	rt, err := NewJSRuntime(source)
	if err != nil {
		t.Fatal(err)
	}
	rt.SetTools(&tools.ToolSet{Dir: dir})
	out, err := rt.Call(context.Background(), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	assertIOShim(t, out)
}

func assertIOShim(t *testing.T, out []byte) {
	t.Helper()
	var got struct {
		Body    string   `json:"body"`
		Lines   []string `json:"lines"`
		Escaped bool     `json:"escaped"`
		Err     string   `json:"err"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("io result = %s: %v", out, err)
	}
	if got.Body != "alpha\nbeta\n" || len(got.Lines) != 2 || got.Lines[0] != "alpha" || got.Lines[1] != "beta" {
		t.Fatalf("io result = %+v", got)
	}
	if got.Escaped || !strings.Contains(got.Err, "escapes") {
		t.Fatalf("escape result = %+v", got)
	}
}
