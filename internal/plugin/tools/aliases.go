package tools

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

// resolveAlias returns the canonical tool name and any fixed leading argument.
func resolveAlias(name string) (string, string) {
	a, ok := aliases[name]
	if !ok {
		return name, ""
	}
	return a.canonical, a.fixedArg
}
