package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"

	"goisekai2/internal/plugin/tools"
)

// LuaRuntime executes a Lua 5.1 plugin with the frozen JSON envelope.
type LuaRuntime struct {
	source string

	// Tools, when non-nil, is bound into every call state as host.*.
	Tools *tools.ToolSet
	Dir   string
}

// SetTools attaches the host tool set exposed to scripts; Call binds it into
// each fresh state.
func (r *LuaRuntime) SetTools(ts *tools.ToolSet) { r.Tools = ts }

// SetDir sets the folder used to resolve plugin-local Lua modules.
func (r *LuaRuntime) SetDir(dir string) { r.Dir = dir }

func configureLuaDir(L *lua.LState, dir string) {
	configureLua(L)
	if dir == "" {
		return
	}
	L.SetGlobal("require", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		path := filepath.Join(dir, name+".lua")
		src, err := os.ReadFile(path)
		if err != nil {
			L.RaiseError("module %s not found: %v", name, err)
			return 0
		}
		if err := L.DoString(string(src)); err != nil {
			L.RaiseError("module %s: %v", name, err)
			return 0
		}
		value := L.Get(-1)
		L.Pop(1)
		L.Push(value)
		return 1
	}))
}

const unsupportedModule = "unsupported module (Lua 5.1 runtime)"

func configureLua(L *lua.LState) {
	L.OpenLibs()
	L.SetGlobal("os", lua.LNil)
	L.SetGlobal("debug", lua.LNil)
	L.SetGlobal("io", lua.LNil)
	L.PreloadModule("ffi", func(*lua.LState) int {
		L.RaiseError("%s: ffi", unsupportedModule)
		return 0
	})
	L.PreloadModule("bit", func(*lua.LState) int {
		L.RaiseError("%s: bit", unsupportedModule)
		return 0
	})
	L.PreloadModule("json", func(L *lua.LState) int {
		mod := L.NewTable()
		L.SetField(mod, "encode", L.NewFunction(func(L *lua.LState) int {
			v, err := json.Marshal(luaValue(L.CheckAny(1)))
			if err != nil {
				L.RaiseError("json encode: %v", err)
			}
			L.Push(lua.LString(v))
			return 1
		}))
		L.SetField(mod, "decode", L.NewFunction(func(L *lua.LState) int {
			var v any
			if err := json.Unmarshal([]byte(L.CheckString(1)), &v); err != nil {
				L.RaiseError("json decode: %v", err)
			}
			L.Push(goValue(L, v))
			return 1
		}))
		L.Push(mod)
		return 1
	})
}

func luaValue(v lua.LValue) any {
	switch value := v.(type) {
	case lua.LString:
		return string(value)
	case lua.LNumber:
		return float64(value)
	case lua.LBool:
		return bool(value)
	case *lua.LTable:
		if n := value.MaxN(); n > 0 {
			out := make([]any, n)
			for i := 1; i <= n; i++ {
				out[i-1] = luaValue(value.RawGetInt(i))
			}
			return out
		}
		out := make(map[string]any)
		value.ForEach(func(k, val lua.LValue) {
			out[fmt.Sprint(luaValue(k))] = luaValue(val)
		})
		return out
	default:
		return nil
	}
}

func goValue(L *lua.LState, v any) lua.LValue {
	switch value := v.(type) {
	case string:
		return lua.LString(value)
	case float64:
		return lua.LNumber(value)
	case bool:
		return lua.LBool(value)
	case map[string]any:
		t := L.NewTable()
		for k, item := range value {
			L.SetField(t, k, goValue(L, item))
		}
		return t
	case []any:
		t := L.NewTable()
		for _, item := range value {
			t.Append(goValue(L, item))
		}
		return t
	default:
		return lua.LNil
	}
}

// NewLuaRuntime compiles source and rejects scripts without a handle function.
func NewLuaRuntime(source string) (*LuaRuntime, error) {
	L := lua.NewState()
	defer L.Close()
	configureLua(L)
	// ponytail: validation compiles syntax only; top-level host calls run at Call time.
	if _, err := L.LoadString(source); err != nil {
		return nil, fmt.Errorf("plugin: lua source: %w", err)
	}
	if !strings.Contains(source, "handle") {
		return nil, fmt.Errorf("plugin: lua: missing handle function")
	}
	return &LuaRuntime{source: source}, nil
}

// NewLuaRuntimeWithTools compiles source with tools bound, so top-level host
// calls can resolve during validation.
func NewLuaRuntimeWithTools(source string, ts *tools.ToolSet) (*LuaRuntime, error) {
	L := lua.NewState()
	defer L.Close()
	configureLua(L)
	if ts != nil {
		bindLua(L, ts)
	}
	if _, err := L.LoadString(source); err != nil {
		return nil, fmt.Errorf("plugin: lua source: %w", err)
	}
	if !strings.Contains(source, "handle") {
		return nil, fmt.Errorf("plugin: lua: missing handle function")
	}
	return &LuaRuntime{source: source, Tools: ts}, nil
}

// Call invokes handle(requestJSON) with a fresh sandboxed state. The context
// deadline cancels a runaway loop.
func (r *LuaRuntime) Call(ctx context.Context, request []byte) ([]byte, error) {
	L := lua.NewState()
	defer L.Close()
	L.SetContext(ctx)
	configureLuaDir(L, r.Dir)
	if r.Tools != nil {
		bindLua(L, r.Tools)
	}
	slog.Debug("plugin lua call start", "dir", r.Dir, "request_bytes", len(request))
	if err := L.DoString(r.source); err != nil {
		return nil, fmt.Errorf("plugin: lua source: %w", err)
	}
	fn := L.GetGlobal("handle")
	if fn == lua.LNil {
		return nil, fmt.Errorf("plugin: lua: missing handle function")
	}
	var decoded any
	if err := json.Unmarshal(request, &decoded); err != nil {
		return nil, fmt.Errorf("plugin: lua: decode request: %w", err)
	}
	luaRequest := goValue(L, decoded)
	slog.Debug("plugin lua request decoded", "type", fmt.Sprintf("%T", luaRequest))
	if err := L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, luaRequest); err != nil {
		return nil, fmt.Errorf("plugin: lua call: %w", err)
	}
	value := L.Get(-1)
	slog.Debug("plugin lua result", "type", fmt.Sprintf("%T", value))
	if text, ok := value.(lua.LString); ok {
		return []byte(text), nil
	}
	body, err := json.Marshal(luaValue(value))
	if err != nil {
		return nil, fmt.Errorf("plugin: lua: encode result: %w", err)
	}
	return body, nil
}
