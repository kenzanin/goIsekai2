package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// LuaRuntime executes a Lua 5.1 plugin with the frozen JSON envelope.
type LuaRuntime struct {
	source string
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
		out := make(map[string]any)
		value.ForEach(func(k, val lua.LValue) {
			out[fmt.Sprint(k)] = luaValue(val)
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
	if err := L.DoString(source); err != nil {
		return nil, fmt.Errorf("plugin: lua source: %w", err)
	}
	if L.GetGlobal("handle") == lua.LNil {
		return nil, fmt.Errorf("plugin: lua: missing handle function")
	}
	return &LuaRuntime{source: source}, nil
}

// Call invokes handle(requestJSON) with a fresh sandboxed state. The context
// deadline cancels a runaway loop.
func (r *LuaRuntime) Call(ctx context.Context, request []byte) ([]byte, error) {
	L := lua.NewState()
	defer L.Close()
	L.SetContext(ctx)
	configureLua(L)
	if err := L.DoString(r.source); err != nil {
		return nil, fmt.Errorf("plugin: lua source: %w", err)
	}
	fn := L.GetGlobal("handle")
	if fn == lua.LNil {
		return nil, fmt.Errorf("plugin: lua: missing handle function")
	}
	if err := L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, lua.LString(string(request))); err != nil {
		return nil, fmt.Errorf("plugin: lua call: %w", err)
	}
	result, ok := L.Get(-1).(lua.LString)
	if !ok {
		return nil, fmt.Errorf("plugin: lua: handle returned no string")
	}
	return []byte(result), nil
}
