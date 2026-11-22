package plugin

import (
	"encoding/json"
	"fmt"

	lua "github.com/yuin/gopher-lua"

	"goisekai2/internal/plugin/tools"
)

// bindLua registers the global host table on a fresh state: host.call as the
// generic JSON envelope over the tool set plus typed wrappers for now, log
// and storage.
func bindLua(L *lua.LState, ts *tools.ToolSet) {
	host := L.NewTable()
	L.SetGlobal("host", host)
	L.SetField(host, "call", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		var args []any
		if err := json.Unmarshal([]byte(L.CheckString(2)), &args); err != nil {
			L.Push(lua.LString(hostResultJSON(false, nil, err)))
			return 1
		}
		v, err := ts.Dispatch(name, args...)
		L.Push(lua.LString(hostResultJSON(err == nil, v, err)))
		return 1
	}))
	L.SetField(host, "now", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(ts.Now()))
		return 1
	}))
	log := L.NewTable()
	L.SetField(log, "info", L.NewFunction(func(L *lua.LState) int {
		ts.LogInfo(L.CheckString(1))
		return 0
	}))
	L.SetField(log, "error", L.NewFunction(func(L *lua.LState) int {
		ts.LogError(L.CheckString(1))
		return 0
	}))
	L.SetField(host, "log", log)
	storage := L.NewTable()
	L.SetField(storage, "get", L.NewFunction(func(L *lua.LState) int {
		v, err := ts.StorageGet(L.CheckString(1))
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(v))
		return 1
	}))
	L.SetField(storage, "set", L.NewFunction(func(L *lua.LState) int {
		if err := ts.StorageSet(L.CheckString(1), L.CheckString(2)); err != nil {
			L.RaiseError("host.storage.set: %v", err)
		}
		return 0
	}))
	L.SetField(host, "storage", storage)
	text := L.NewTable()
	L.SetField(text, "normalize_status", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(tools.NormalizeStatus(L.CheckString(1))))
		return 1
	}))
	L.SetField(text, "clean", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(tools.Clean(L.CheckString(1))))
		return 1
	}))
	L.SetField(text, "strip_html", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(tools.Clean(L.CheckString(1))))
		return 1
	}))
	L.SetField(host, "text", text)
	plugin := L.NewTable()
	L.SetField(plugin, "param", L.NewFunction(func(L *lua.LState) int {
		if v, ok := ts.Params[L.CheckString(1)]; ok {
			L.Push(goToLua(L, v))
		} else {
			L.Push(lua.LNil)
		}
		return 1
	}))
	L.SetField(host, "plugin", plugin)
	L.SetField(host, "url_encode", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LString(tools.URLEncode(L.CheckString(1))))
		return 1
	}))
	L.SetField(host, "url_decode", L.NewFunction(func(L *lua.LState) int {
		v, err := tools.URLDecode(L.CheckString(1))
		if err != nil {
			L.RaiseError("host.url_decode: %v", err)
		}
		L.Push(lua.LString(v))
		return 1
	}))
	httpTable := L.NewTable()
	for _, name := range []string{"fetch", "fetch_body", "get_json", "post_json"} {
		tool := name
		L.SetField(httpTable, name, L.NewFunction(func(L *lua.LState) int {
			args := []any{L.CheckString(1)}
			if L.GetTop() >= 2 && L.CheckAny(2) != lua.LNil {
				args = append(args, luaValue(L.CheckAny(2)))
			}
			v, err := ts.Dispatch("http."+tool, args...)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			L.Push(goToLua(L, v))
			L.Push(lua.LNil)
			return 2
		}))
	}
	L.SetField(host, "http", httpTable)
}

func goToLua(L *lua.LState, v any) lua.LValue {
	switch x := v.(type) {
	case string:
		return lua.LString(x)
	case float64:
		return lua.LNumber(x)
	case int:
		return lua.LNumber(x)
	case bool:
		return lua.LBool(x)
	case map[string]any:
		t := L.NewTable()
		for k, v := range x {
			L.SetField(t, k, goToLua(L, v))
		}
		return t
	case []any:
		t := L.NewTable()
		for _, v := range x {
			t.Append(goToLua(L, v))
		}
		return t
	default:
		return lua.LString(fmt.Sprint(x))
	}
}

// hostResult is the JSON envelope returned by host.call.
type hostResult struct {
	Ok    bool   `json:"ok"`
	Data  any    `json:"data"`
	Error string `json:"error,omitempty"`
}

// hostResultJSON renders the frozen {ok,data,error} envelope.
func hostResultJSON(ok bool, data any, err error) string {
	res := hostResult{Ok: ok, Data: data}
	if err != nil {
		res.Error = err.Error()
	}
	b, marshalErr := json.Marshal(res)
	if marshalErr != nil {
		return `{"ok":false,"error":"host: envelope marshal"}`
	}
	return string(b)
}
