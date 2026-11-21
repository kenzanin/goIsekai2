package plugin

import (
	"encoding/json"

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
