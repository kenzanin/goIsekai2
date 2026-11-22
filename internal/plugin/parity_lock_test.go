package plugin

import (
	"context"
	"encoding/json"
	"testing"

	"goisekai2/internal/plugin/tools"
)

func TestLuaHostSurfaceParity(t *testing.T) {
	const source = `function handle(req)
  local checks = {
    log_debug = type(host.log.debug) == "function",
    config_get = type(host.config.get) == "function",
    plugin_param = type(host.plugin.param) == "function",
    text_chapter = type(host.text.chapter_num) == "function",
    text_strip = type(host.text.strip_html) == "function",
    codecs_b64 = type(host.codecs.base64_encode) == "function",
    crypto_vrf = type(host.crypto.vrf_sign) == "function",
    re2_gmatch = type(host.re2.gmatch) == "function",
    re2_callable = type(host.re2) == "table" and getmetatable(host.re2).__call ~= nil,
    html_select = type(host.html.select) == "function",
    browser_call = type(host.browser.call) == "function",
    fs_read_lib = type(host.fs.read_lib) == "function",
  }
  return checks
end`
	rt, err := NewLuaRuntimeWithTools(source, &tools.ToolSet{})
	if err != nil {
		t.Fatal(err)
	}
	out, err := rt.Call(context.Background(), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	assertAllTrue(t, out)
}

func TestJSHostSurfaceParity(t *testing.T) {
	const source = `function handle(req) { return {
  log_debug: typeof host.log.debug === "function",
  config_get: typeof host.config.get === "function",
  plugin_param: typeof host.plugin.param === "function",
  text_chapter: typeof host.text.chapter_num === "function",
  text_strip: typeof host.text.strip_html === "function",
  codecs_b64: typeof host.codecs.base64_encode === "function",
  crypto_vrf: typeof host.crypto.vrf_sign === "function",
  re2_gmatch: typeof host.re2.gmatch === "function",
  html_select: typeof host.html.select === "function",
  browser_call: typeof host.browser.call === "function",
  fs_read_lib: typeof host.fs.read_lib === "function"
}; }`
	rt, err := NewJSRuntime(source)
	if err != nil {
		t.Fatal(err)
	}
	rt.SetTools(&tools.ToolSet{})
	out, err := rt.Call(context.Background(), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	assertAllTrue(t, out)
}

func assertAllTrue(t *testing.T, out []byte) {
	t.Helper()
	var got map[string]bool
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("surface result = %s: %v", out, err)
	}
	for key, ok := range got {
		if !ok {
			t.Errorf("%s is false", key)
		}
	}
	if len(got) == 0 {
		t.Fatal("empty surface result")
	}
}

func TestDispatcherRegistryCoversBaselineNames(t *testing.T) {
	want := []string{
		"crypto.vrf_sign", "crypto.aes_gcm_decrypt", "codecs.base64_encode",
		"re2.gmatch", "text.chapter_num", "text.date_to_iso", "plugin.param",
		"json.encode", "json.decode", "browser.call", "html.xpath",
	}
	names := map[string]bool{}
	for _, n := range tools.Names() {
		names[n] = true
	}
	for _, n := range want {
		if !names[n] {
			t.Errorf("dispatcher registry missing %q", n)
		}
	}
}
