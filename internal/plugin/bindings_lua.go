package plugin

import (
	"encoding/json"
	"fmt"

	"modernc.org/lua"

	"goisekai2/internal/plugin/tools"
)

// bindLua registers the host table and the JSON bridge on a fresh VM. Typed
// wrappers are added by luaHostHelper after these Go-backed functions exist.
func bindLua(m *lua.VM, ts *tools.ToolSet) error {
	host, err := m.NewTable()
	if err != nil {
		return err
	}
	if err := m.SetGlobal("host", host); err != nil {
		return err
	}
	if err := host.Set("call", func(name, argsJSON string) string {
		var args []any
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return hostResultJSON(false, nil, err)
		}
		if ts == nil {
			return hostResultJSON(false, nil, fmt.Errorf("host tools unavailable"))
		}
		v, err := ts.Dispatch(name, args...)
		return hostResultJSON(err == nil, v, err)
	}); err != nil {
		return err
	}

	params, err := m.NewTable()
	if err != nil {
		return err
	}
	if err := params.Set("param", func(key string) any {
		if ts == nil {
			return nil
		}
		return ts.Params[key]
	}); err != nil {
		return err
	}
	if err := host.Set("plugin", params); err != nil {
		return err
	}

	textTable, err := m.NewTable()
	if err != nil {
		return err
	}
	if err := textTable.Set("normalize_status", func(s string) string { return tools.NormalizeStatus(s) }); err != nil {
		return err
	}
	if err := host.Set("text", textTable); err != nil {
		return err
	}

	js, err := m.NewTable()
	if err != nil {
		return err
	}
	if err := js.Set("decode", func(s string) (any, error) {
		var v any
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			return nil, err
		}
		return v, nil
	}); err != nil {
		return err
	}
	return m.SetGlobal("json", js)
}

type hostResult struct {
	Ok    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
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
