package tools

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
)

func argString(args []any, i int) (string, error) {
	if i >= len(args) {
		return "", fmt.Errorf("tools: missing argument %d", i)
	}
	switch v := args[i].(type) {
	case string:
		return v, nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return "", fmt.Errorf("tools: argument %d is not a string", i)
	}
}

func argMap(args []any, i int) (map[string]any, error) {
	if i >= len(args) || args[i] == nil {
		return nil, nil
	}
	switch v := args[i].(type) {
	case map[string]any:
		return v, nil
	case map[string]string:
		out := make(map[string]any, len(v))
		for k, item := range v {
			out[k] = item
		}
		return out, nil
	default:
		return nil, fmt.Errorf("tools: argument %d is not a map", i)
	}
}

// Dispatch invokes one frozen host tool by name. Runtime bindings are thin
// marshals over this single implementation.
func (t *ToolSet) Dispatch(name string, args ...any) (any, error) {
	canonical, fixed := resolveAlias(name)
	if fixed != "" {
		args = append([]any{fixed}, args...)
	}
	name = canonical
	slog.Debug("host tool dispatch", "tool", name, "args", len(args), "plugin", t.PluginID)
	switch name {
	case "now":
		return t.Now(), nil
	case "log.info":
		msg, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		t.LogInfo(msg)
		return nil, nil
	case "log.error":
		msg, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		t.LogError(msg)
		return nil, nil
	case "log.debug":
		msg, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		t.LogDebug(msg)
		return nil, nil
	case "http.fetch":
		u, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		opts, err := argMap(args, 1)
		if err != nil {
			return nil, err
		}
		return t.Fetch(u, opts)
	case "http.fetch_body":
		u, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		opts, err := argMap(args, 1)
		if err != nil {
			return nil, err
		}
		return t.FetchBody(u, opts)
	case "http.get_json":
		u, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		opts, err := argMap(args, 1)
		if err != nil {
			return nil, err
		}
		return t.GetJSON(u, opts)
	case "http.post_json":
		u, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		if len(args) < 2 {
			return nil, fmt.Errorf("tools: post_json needs a body")
		}
		opts, err := argMap(args, 2)
		if err != nil {
			return nil, err
		}
		return t.PostJSON(u, args[1], opts)
	case "storage.get":
		key, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		return t.StorageGet(key)
	case "storage.set":
		key, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		value, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		return nil, t.StorageSet(key, value)
	case "fs.read":
		path, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		return Read(t.Dir, path)
	case "fs.read_lib":
		path, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		return ReadLib(t.Dir, path)
	case "text.normalize_status":
		s, err := argString(args, 0)
		return NormalizeStatus(s), err
	case "plugin.param":
		key, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		return t.Params[key], nil
	case "config.get":
		key, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		value, ok := t.ConfigGet(key)
		if !ok {
			return nil, nil
		}
		return value, nil
	case "url_encode":
		s, err := argString(args, 0)
		return URLEncode(s), err
	case "url_decode":
		s, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		return URLDecode(s)
	case "url_abs":
		base, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		ref, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		b, err := url.Parse(base)
		if err != nil {
			return nil, err
		}
		r, err := url.Parse(ref)
		if err != nil {
			return nil, err
		}
		return b.ResolveReference(r).String(), nil
	case "re2.gmatch":
		p, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		s, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		return Gmatch(p, s)
	case "re2.match", "re2.find", "re2.capture", "re2.replace", "re2.find_index":
		p, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		s, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		switch name {
		case "re2.match":
			return Match(p, s)
		case "re2.find":
			return Find(p, s)
		case "re2.capture":
			return Capture(p, s)
		case "re2.replace":
			repl, err := argString(args, 2)
			if err != nil {
				return nil, err
			}
			return Replace(p, s, repl)
		case "browser.call":
			if t.BrowserAllowed == nil || !t.BrowserAllowed(t.PluginID) || t.Browser == nil {
				return nil, fmt.Errorf("tools: unknown tool %q", name)
			}
			tool, err := argString(args, 0)
			if err != nil {
				return nil, err
			}
			params, err := argMap(args, 1)
			if err != nil {
				return nil, err
			}
			return t.Browser.Call(tool, params)
		default:
			return FindIndex(p, s)
		}
	case "re2.find_all":
		p, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		s, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		limit := -1
		if len(args) > 2 {
			if n, ok := args[2].(int); ok {
				limit = n
			}
		}
		return FindAll(p, s, limit)
	case "re2.quote":
		s, err := argString(args, 0)
		return Quote(s), err
	case "html.select", "html.xpath":
		doc, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		q, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		if name == "html.xpath" {
			return XPath(doc, q)
		}
		return Select(doc, q)
	case "html.attr":
		doc, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		q, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		attr, err := argString(args, 2)
		if err != nil {
			return nil, err
		}
		return Attr(doc, q, attr)
	case "json.encode":
		if len(args) == 0 {
			return "null", nil
		}
		b, err := json.Marshal(args[0])
		return string(b), err
	case "json.decode":
		s, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		var v any
		err = json.Unmarshal([]byte(s), &v)
		return v, err
	case "text.json_blob":
		s, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		marker, err := argString(args, 1)
		if err != nil {
			marker = ""
		}
		blob, ok := JSONBlobFrom(s, marker)
		if !ok {
			return nil, nil
		}
		return blob, nil
	case "text.date_epoch":
		s, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		return DateEpoch(s)
	case "text.clean":
		s, err := argString(args, 0)
		return Clean(s), err
	case "text.chapter_num":
		s, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		return ChapterNum(s), nil
	case "text.date_to_iso":
		s, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		return DateToISO(s), nil
	case "text.trim":
		s, err := argString(args, 0)
		return Trim(s), err
	case "text.unescape":
		s, err := argString(args, 0)
		return Unescape(s), err
	case "text.strip_html":
		s, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		return StripHTML(s), nil
	case "crypto.hash", "crypto.hmac":
		algorithm, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		value, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		if name == "crypto.hash" {
			return Hash(algorithm, value)
		}
		key, err := argString(args, 2)
		if err != nil {
			return nil, err
		}
		return HMAC(algorithm, key, value)
	case "crypto.sign":
		payload, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		key, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		return Sign(payload, key)
	case "crypto.hmac_sha256_hex":
		data, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		key, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		return HMAC("sha256", key, data)
	case "crypto.vrf_sign":
		path, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		params, ok := asStringMap(args[1])
		if !ok {
			return nil, fmt.Errorf("tools: vrf_sign params must be an object")
		}
		stages, ok := asAnySlice(args[2])
		if !ok {
			return nil, fmt.Errorf("tools: vrf_sign stages must be an array")
		}
		return VRFSign(path, params, stages)
	case "crypto.aes_gcm_decrypt":
		iv, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		ct, err := argString(args, 1)
		if err != nil {
			return nil, err
		}
		tag, err := argString(args, 2)
		if err != nil {
			return nil, err
		}
		key, err := argString(args, 3)
		if err != nil {
			return nil, err
		}
		return AESGCMDecrypt(iv, ct, tag, key)
	case "crypto.b64enc":
		s, err := argString(args, 0)
		return Base64Encode(s), err
	case "crypto.b64dec":
		s, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		return Base64Decode(s)
	case "crypto.hexenc":
		s, err := argString(args, 0)
		return HexEncode(s), err
	case "crypto.hexdec":
		s, err := argString(args, 0)
		if err != nil {
			return nil, err
		}
		return HexDecode(s)
	default:
		return nil, fmt.Errorf("tools: unknown tool %q", name)
	}
}
