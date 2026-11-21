package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// defaultProfile is the fallback TLS fingerprint, matching the V1 default.
const defaultProfile = "chrome_146"

// defaultTimeoutMS applies when timeout_ms is absent or not positive.
const defaultTimeoutMS = 30000

// clientProfiles maps plugin-visible profile names to tls-client fingerprints.
var clientProfiles = map[string]profiles.ClientProfile{
	"chrome_146":  profiles.Chrome_146,
	"chrome_133":  profiles.Chrome_133,
	"chrome_124":  profiles.Chrome_124,
	"chrome_120":  profiles.Chrome_120,
	"firefox_148": profiles.Firefox_148,
	"firefox_133": profiles.Firefox_133,
	"firefox_117": profiles.Firefox_117,
}

// fetchOpts is the parsed form of the loose Fetch options map.
type fetchOpts struct {
	method    string
	body      string
	profile   string
	timeoutMS int
	headers   map[string]string
}

// Fetch performs one outbound request with a browser-like TLS fingerprint and
// returns status (int), body (string) and lowercased headers; any HTTP status,
// including non-2xx, is returned as data, only transport failures error.
func (t *ToolSet) Fetch(url string, options map[string]any) (map[string]any, error) {
	return t.fetch(url, parseOptions(options))
}

// FetchBody returns the response body, treating every status but 200 as an
// "upstream status <code>" error.
func (t *ToolSet) FetchBody(url string, options map[string]any) (string, error) {
	res, err := t.fetch(url, parseOptions(options))
	if err != nil {
		return "", err
	}
	if status, _ := res["status"].(int); status != 200 {
		return "", fmt.Errorf("upstream status %d", status)
	}
	body, _ := res["body"].(string)
	return body, nil
}

// GetJSON fetches url and decodes the JSON response; options follow Fetch.
func (t *ToolSet) GetJSON(url string, options map[string]any) (any, error) {
	return t.jsonFetch("GET", url, "", options)
}

// PostJSON posts body (marshalled to JSON) and decodes the JSON response;
// options follow Fetch, so a bare header map is accepted too.
func (t *ToolSet) PostJSON(url string, body any, options map[string]any) (any, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("json encode: %w", err)
	}
	return t.jsonFetch("POST", url, string(data), options)
}

// jsonFetch runs the request with a forced method and decodes the body as JSON.
func (t *ToolSet) jsonFetch(method, url, body string, options map[string]any) (any, error) {
	o := parseOptions(options)
	o.method = method
	if body != "" {
		o.body = body
		if !hasHeader(o.headers, "Content-Type") {
			o.headers["Content-Type"] = "application/json"
		}
	}
	res, err := t.fetch(url, o)
	if err != nil {
		return nil, err
	}
	var out any
	if err := json.Unmarshal([]byte(res["body"].(string)), &out); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}
	return out, nil
}

// fetch builds the one-shot client, performs the request and collects the
// result envelope.
func (t *ToolSet) fetch(url string, o fetchOpts) (map[string]any, error) {
	client, err := t.httpClient(o.profile, o.timeoutMS)
	if err != nil {
		return nil, fmt.Errorf("http client: %w", err)
	}
	var body io.Reader
	if o.body != "" {
		body = strings.NewReader(o.body)
	}
	req, err := fhttp.NewRequest(o.method, url, body)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	for k, v := range o.headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http body: %w", err)
	}
	hdrs := map[string]string{}
	for k, vs := range resp.Header {
		for _, v := range vs { // last value wins, keys lowercased
			hdrs[strings.ToLower(k)] = v
		}
	}
	return map[string]any{"status": resp.StatusCode, "body": string(raw), "headers": hdrs}, nil
}

// httpClient builds a one-shot tls-client; an unknown non-empty profile warns
// and falls back to the default fingerprint.
func (t *ToolSet) httpClient(profile string, timeoutMS int) (tls_client.HttpClient, error) {
	p, ok := clientProfiles[profile]
	if !ok {
		if profile != "" {
			t.log(slog.LevelWarn, "unknown tls profile "+profile+", using "+defaultProfile)
		}
		p = clientProfiles[defaultProfile]
	}
	if timeoutMS <= 0 {
		timeoutMS = defaultTimeoutMS
	}
	return tls_client.NewHttpClient(nil,
		tls_client.WithTimeoutMilliseconds(timeoutMS),
		tls_client.WithClientProfile(p))
}

// parseOptions maps the loose options onto fetchOpts; a map without any known
// key is taken as a bare header map (cpp get_json/post_json parity).
func parseOptions(options map[string]any) fetchOpts {
	o := fetchOpts{headers: map[string]string{}}
	known := 0
	for k, v := range options {
		switch k {
		case "method":
			o.method, known = asString(v), known+1
		case "body":
			o.body, known = asString(v), known+1
		case "profile":
			o.profile, known = asString(v), known+1
		case "timeout_ms":
			o.timeoutMS, known = asInt(v), known+1
		case "headers":
			known++
			switch h := v.(type) {
			case map[string]any:
				for hk, hv := range h {
					o.headers[hk] = asString(hv)
				}
			case map[string]string:
				for hk, hv := range h {
					o.headers[hk] = hv
				}
			}
		}
	}
	if known == 0 {
		for k, v := range options {
			o.headers[k] = asString(v)
		}
	}
	return o
}

// asString coerces an option value to string without failing.
func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

// asInt coerces a numeric option value to int, tolerating JSON decoding.
func asInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

// hasHeader reports whether the map already carries the named header, ignoring case.
func hasHeader(h map[string]string, want string) bool {
	for k := range h {
		if strings.EqualFold(k, want) {
			return true
		}
	}
	return false
}
