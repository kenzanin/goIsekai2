package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"modernc.org/lua"

	"goisekai2/internal/plugin/tools"
)

// LuaRuntime executes a Lua 5.5 plugin with the frozen JSON envelope.
type LuaRuntime struct {
	source string

	// Tools, when non-nil, is bound into every call state as host.*.
	Tools *tools.ToolSet
	Dir   string
}

// SetTools attaches the host tool set exposed to scripts; Call binds it into
// every fresh state.
func (r *LuaRuntime) SetTools(ts *tools.ToolSet) { r.Tools = ts }

// SetDir sets the plugin directory used by the confined require shim.
func (r *LuaRuntime) SetDir(dir string) { r.Dir = dir }

const unsupportedModule = "unsupported module (Lua 5.5 runtime)"

// newVM builds a sandboxed VM with the frozen host bridge and helper chunks.
func (r *LuaRuntime) newVM() (*lua.VM, error) {
	m, err := lua.NewVM()
	if err != nil {
		return nil, fmt.Errorf("plugin: lua vm: %w", err)
	}
	for _, name := range []string{"os", "io", "debug", "dofile", "loadfile", "package"} {
		if err := m.SetGlobal(name, nil); err != nil {
			_ = m.Close()
			return nil, fmt.Errorf("plugin: lua sandbox: %w", err)
		}
	}
	if err := bindLua(m, r.Tools); err != nil {
		_ = m.Close()
		return nil, fmt.Errorf("plugin: lua bindings: %w", err)
	}
	if err := m.SetGlobal("require", r.requireFunc(m)); err != nil {
		_ = m.Close()
		return nil, fmt.Errorf("plugin: lua require: %w", err)
	}
	for _, helper := range []string{luaJSONHelper, luaHostHelper} {
		if _, err := m.EvalChunk("=goisekai-helper", helper); err != nil {
			_ = m.Close()
			return nil, fmt.Errorf("plugin: lua helper: %w", err)
		}
	}
	return m, nil
}

// requireFunc resolves plugin-local modules and returns their JSON encoding.
// The Lua wrapper decodes that string back into a table.
func (r *LuaRuntime) requireFunc(m *lua.VM) func(string) (string, error) {
	cache := map[string]string{}
	return func(name string) (string, error) {
		if name == "ffi" || name == "bit" {
			return "", fmt.Errorf("%s: %s", unsupportedModule, name)
		}
		if encoded, ok := cache[name]; ok {
			return encoded, nil
		}
		if r.Dir == "" {
			return "", fmt.Errorf("plugin: lua: module %s not found", name)
		}
		base, err := filepath.Abs(r.Dir)
		if err != nil {
			return "", fmt.Errorf("plugin: lua: module %s: %w", name, err)
		}
		path, err := filepath.Abs(filepath.Join(base, name+".lua"))
		if err != nil {
			return "", fmt.Errorf("plugin: lua: module %s: %w", name, err)
		}
		rel, err := filepath.Rel(base, path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("plugin: lua: module %s outside plugin directory", name)
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("plugin: lua: module %s not found: %w", name, err)
		}
		values, err := m.EvalChunk("@"+path, string(src))
		if err != nil {
			return "", fmt.Errorf("plugin: lua: module %s: %w", name, err)
		}
		if len(values) == 0 || values[0] == nil {
			return "null", nil
		}
		encoded, err := m.Call("__go_encode", values[0])
		if err != nil || len(encoded) == 0 {
			if err == nil {
				err = fmt.Errorf("no encoded result")
			}
			return "", fmt.Errorf("plugin: lua: module %s encode: %w", name, err)
		}
		text, ok := encoded[0].(string)
		if !ok {
			return "", fmt.Errorf("plugin: lua: module %s encode: unexpected result", name)
		}
		cache[name] = text
		return text, nil
	}
}

// NewLuaRuntime compiles source and rejects scripts without a handle function.
func NewLuaRuntime(source string) (*LuaRuntime, error) {
	return NewLuaRuntimeWithTools(source, nil)
}

// NewLuaRuntimeWithTools compiles source with tools bound, so top-level host
// calls during validation see the same bridge as a live call.
func NewLuaRuntimeWithTools(source string, ts *tools.ToolSet) (*LuaRuntime, error) {
	r := &LuaRuntime{source: source, Tools: ts}
	m, err := r.newVM()
	if err != nil {
		return nil, err
	}
	defer m.Close()
	if err := m.SetGlobal("__source", source); err != nil {
		return nil, fmt.Errorf("plugin: lua source: %w", err)
	}
	values, err := m.EvalChunk("=validate", "return load(__source, '=plugin', 't')")
	if err != nil {
		return nil, fmt.Errorf("plugin: lua source: %w", err)
	}
	if len(values) == 0 || values[0] == nil {
		msg := "invalid syntax"
		if len(values) > 1 {
			msg = fmt.Sprint(values[1])
		}
		return nil, fmt.Errorf("plugin: lua source: %s", msg)
	}
	if !strings.Contains(source, "handle") {
		return nil, fmt.Errorf("plugin: lua: missing handle function")
	}
	return r, nil
}

// Call invokes handle(requestJSON) with a fresh sandboxed state. The context
// deadline interrupts a runaway loop.
func (r *LuaRuntime) Call(ctx context.Context, request []byte) ([]byte, error) {
	m, err := r.newVM()
	if err != nil {
		return nil, err
	}
	defer m.Close()

	done := make(chan struct{})
	defer close(done)
	if deadline, ok := ctx.Deadline(); ok {
		m.SetEvalTimeout(time.Until(deadline))
	}
	go func() {
		select {
		case <-ctx.Done():
			m.Interrupt()
		case <-done:
		}
	}()

	slog.Debug("plugin lua call start", "dir", r.Dir, "request_bytes", len(request))
	if _, err := m.EvalChunk("=plugin", r.source); err != nil {
		return nil, fmt.Errorf("plugin: lua source: %w", err)
	}
	if _, err := m.EvalChunk("=goisekai-invoke", luaInvokeHelper); err != nil {
		return nil, fmt.Errorf("plugin: lua invoke: %w", err)
	}

	var decoded any
	if err := json.Unmarshal(request, &decoded); err != nil {
		return nil, fmt.Errorf("plugin: lua: decode request: %w", err)
	}
	values, err := m.Call("__invoke", decoded)
	if err != nil {
		return nil, fmt.Errorf("plugin: lua call: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("plugin: lua call: %w", err)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("plugin: lua call: no result")
	}
	slog.Debug("plugin lua result", "type", fmt.Sprintf("%T", values[0]))
	if text, ok := values[0].(string); ok {
		return []byte(text), nil
	}
	body, err := json.Marshal(values[0])
	if err != nil {
		return nil, fmt.Errorf("plugin: lua: encode result: %w", err)
	}
	return body, nil
}
