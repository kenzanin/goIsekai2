package plugin

import (
	"encoding/json"
	"fmt"

	"modernc.org/quickjs"

	"goisekai2/internal/plugin/tools"
)

// bindJS registers the host bridge on a fresh VM: __host_call as the generic
// JSON envelope over the tool set, plus the global host object built on it.
func bindJS(vm *quickjs.VM, ts *tools.ToolSet) error {
	err := vm.RegisterHostFunc("__host_call", func(args []any) (any, error) {
		if len(args) != 2 {
			return hostResultJSON(false, nil, fmt.Errorf("host: __host_call(name, jsonArgs)")), nil
		}
		name, okName := args[0].(string)
		raw, okJSON := args[1].(string)
		if !okName || !okJSON {
			return hostResultJSON(false, nil, fmt.Errorf("host: __host_call needs string arguments")), nil
		}
		var callArgs []any
		if err := json.Unmarshal([]byte(raw), &callArgs); err != nil {
			return hostResultJSON(false, nil, err), nil
		}
		v, err := ts.Dispatch(name, callArgs...)
		return hostResultJSON(err == nil, v, err), nil
	})
	if err != nil {
		return fmt.Errorf("plugin: js: bind __host_call: %w", err)
	}
	if _, err := vm.Eval(jsHostBridge, 0); err != nil {
		return fmt.Errorf("plugin: js: bind host: %w", err)
	}
	return nil
}

// jsHostBridge defines the global host object: call returns the raw envelope,
// every typed wrapper throws the envelope error and yields the data.
const jsHostBridge = `var host = (function(){
  function envelope(name, args){
    return JSON.parse(__host_call(name, JSON.stringify(args === undefined ? [] : args)));
  }
  function rpc(name, args){
    var res = envelope(name, args);
    if (!res.ok) throw new Error(res.error);
    return res.data;
  }
  return {
    call: envelope,
    now: function(){ return rpc("now"); },
    log: {
      info: function(msg){ rpc("log.info", [msg]); },
      error: function(msg){ rpc("log.error", [msg]); }
    },
    storage: {
      get: function(key){ return rpc("storage.get", [key]); },
      set: function(key, value){ rpc("storage.set", [key, value]); }
    }
  };
})();`
