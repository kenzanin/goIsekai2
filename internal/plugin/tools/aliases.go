package tools

import "sort"

// alias forwards a baseline host name to the single dispatcher implementation
// that owns its behavior. fixedArg is prepended when the target needs a
// constant selector (for example a hash algorithm).
type alias struct {
	canonical string
	fixedArg  string
}

// aliases maps cppIsekai-compatible names onto goIsekai2 canonical tools.
// Names not listed here are already canonical and resolve unchanged.
var aliases = map[string]alias{
	"codecs.base64_encode": {canonical: "crypto.b64enc"},
	"codecs.base64_decode": {canonical: "crypto.b64dec"},
	"codecs.hex_encode":    {canonical: "crypto.hexenc"},
	"codecs.hex_decode":    {canonical: "crypto.hexdec"},
	"crypto.sha256_hex":    {canonical: "crypto.hash", fixedArg: "sha256"},
	"crypto.md5_hex":       {canonical: "crypto.hash", fixedArg: "md5"},
}

// Names returns every dispatcher-visible host name, canonical and aliased.
// Runtime bindings use it as the parity registry in lock tests.
func Names() []string {
	names := []string{
		"now", "log.info", "log.error", "log.debug",
		"http.fetch", "http.fetch_body", "http.get_json", "http.post_json",
		"storage.get", "storage.set", "fs.read", "fs.read_lib", "config.get",
		"json.encode", "json.decode",
		"url_encode", "url_decode", "url_abs",
		"re2.match", "re2.find", "re2.capture", "re2.replace", "re2.find_index",
		"re2.find_all", "re2.quote", "re2.gmatch",
		"browser.call", "html.select", "html.attr", "html.xpath",
		"text.json_blob", "text.date_epoch", "text.clean", "text.chapter_num",
		"text.date_to_iso", "text.trim", "text.unescape", "text.strip_html",
		"text.normalize_status", "plugin.param",
		"crypto.sha256_hex", "crypto.md5_hex", "crypto.hmac_sha256_hex",
		"crypto.vrf_sign", "crypto.aes_gcm_decrypt",
	}
	for name := range aliases {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// resolveAlias returns the canonical tool name and any fixed leading argument.
func resolveAlias(name string) (string, string) {
	a, ok := aliases[name]
	if !ok {
		return name, ""
	}
	return a.canonical, a.fixedArg
}
