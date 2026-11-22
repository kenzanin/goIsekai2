package tools

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// NormalizeStatus canonicalises an upstream status label for enrichment
// candidates; unknown labels pass through unchanged.
func NormalizeStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "unknown":
		return "unknown"
	case "ongoing", "publishing":
		return "ongoing"
	case "hiatus", "on hold":
		return "hiatus"
	case "canceled", "cancelled":
		return "canceled"
	case "completed", "complete", "finished":
		return "completed"
	default:
		return status
	}
}

// URLEncode percent-encodes s for use in query strings and path segments.
func URLEncode(s string) string { return url.QueryEscape(s) }

// URLDecode reverses URLEncode; malformed escapes are an error.
func URLDecode(s string) (string, error) { return url.QueryUnescape(s) }

// JSONBlob serializes v to compact JSON with deterministic key order
// (struct fields in declaration order, map keys sorted); it returns ""
// when v is not marshalable.
func JSONBlob(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// DateEpoch parses s as RFC3339 or as a plain 2006-01-02 date and returns
// the corresponding Unix seconds.
func DateEpoch(s string) (int64, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Unix(), nil
		}
	}
	return 0, &time.ParseError{Value: s, Message: ": not RFC3339 nor 2006-01-02"}
}

// Clean strips non-space control characters, collapses every whitespace run
// to a single space and trims the result.
func Clean(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsControl(r) && !unicode.IsSpace(r) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
