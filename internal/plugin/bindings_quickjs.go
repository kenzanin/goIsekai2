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
      error: function(msg){ rpc("log.error", [msg]); },
      debug: function(msg){ rpc("log.debug", [msg]); }
    },
    storage: {
      get: function(key){ return rpc("storage.get", [key]); },
      set: function(key, value){ rpc("storage.set", [key, value]); }
    },
    text: {
      normalize_status: function(s){ return rpc("text.normalize_status", [s]); },
      chapter_num: function(s){ return rpc("text.chapter_num", [s]); },
      date_to_iso: function(s){ return rpc("text.date_to_iso", [s]); },
      date_epoch: function(s){ return rpc("text.date_epoch", [s]); },
      json_blob: function(s, marker){ return rpc("text.json_blob", [s, marker]); },
      trim: function(s){ return rpc("text.trim", [s]); },
      unescape: function(s){ return rpc("text.unescape", [s]); },
      strip_html: function(s){ return rpc("text.strip_html", [s]); },
      clean: function(s){ return rpc("text.clean", [s]); }
    },
    plugin: { param: function(key){ return rpc("plugin.param", [key]); } },
    config: { get: function(key){ return rpc("config.get", [key]); } },
    codecs: {
      base64_encode: function(s){ return rpc("codecs.base64_encode", [s]); },
      base64_decode: function(s){ return rpc("codecs.base64_decode", [s]); },
      hex_encode: function(s){ return rpc("codecs.hex_encode", [s]); },
      hex_decode: function(s){ return rpc("codecs.hex_decode", [s]); }
    },
    crypto: {
      vrf_sign: function(path, params, stages){ return rpc("crypto.vrf_sign", [path, params, stages]); },
      sha256_hex: function(s){ return rpc("crypto.sha256_hex", [s]); },
      md5_hex: function(s){ return rpc("crypto.md5_hex", [s]); },
      hmac_sha256_hex: function(data, key){ return rpc("crypto.hmac_sha256_hex", [data, key]); },
      aes_gcm_decrypt: function(iv, ct, tag, key){ return rpc("crypto.aes_gcm_decrypt", [iv, ct, tag, key]); }
    },
    re2: {
      find: function(p, t){ return rpc("re2.find", [p, t]); },
      match: function(p, t){ return rpc("re2.match", [p, t]); },
      find_all: function(p, t){ return rpc("re2.find_all", [p, t]); },
      gmatch: function(p, t){ return rpc("re2.gmatch", [p, t]); },
      replace: function(p, r, t){ return rpc("re2.replace", [p, r, t]); },
      quote: function(s){ return rpc("re2.quote", [s]); },
      find_index: function(p, t){ return rpc("re2.find_index", [p, t]); },
      capture: function(p, t){ return rpc("re2.capture", [p, t]); }
    },
    browser: { call: function(tool, args){ return rpc("browser.call", [tool, args]); } },
    html: {
      select: function(doc, q){ return rpc("html.select", [doc, q]); },
      attr: function(doc, q, a){ return rpc("html.attr", [doc, q, a]); },
      xpath: function(expr, doc){ return rpc("html.xpath", [expr, doc]); }
    },
    io: {
      read: function(name){ return rpc("fs.read", [name]); },
      lines: function(name){
        var body = rpc("fs.read", [name]);
        var pos = 0;
        return function(){
          if (pos >= body.length) return undefined;
          var nl = body.indexOf("\n", pos);
          if (nl < 0) { var rest = body.slice(pos); pos = body.length; return rest; }
          var line = body.slice(pos, nl);
          pos = nl + 1;
          return line;
        };
      }
    },
    url_encode: function(s){ return rpc("url_encode", [s]); },
    url_decode: function(s){ return rpc("url_decode", [s]); },
    fs: {
      read: function(name){ return rpc("fs.read", [name]); },
      read_lib: function(name){ return rpc("fs.read_lib", [name]); }
    },
    http: {
      fetch: function(url, opts){ return rpc("http.fetch", opts === undefined ? [url] : [url, opts]); },
      fetch_body: function(url, opts){ return rpc("http.fetch_body", opts === undefined ? [url] : [url, opts]); },
      get_json: function(url, opts){ return rpc("http.get_json", opts === undefined ? [url] : [url, opts]); },
      post_json: function(url, body, opts){ return rpc("http.post_json", opts === undefined ? [url, body] : [url, body, opts]); }
    }
  };
})();`
