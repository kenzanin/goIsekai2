package tools

import (
	"encoding/json"
	stdhtml "html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	xhtml "golang.org/x/net/html"
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

// ChapterNum returns the first numeric token in s, or nil when none exists.
// It follows the baseline's optional-double contract for chapter labels.
func ChapterNum(s string) any {
	re := regexp.MustCompile(`[0-9]+(?:\.[0-9]+)?`)
	m := re.FindString(s)
	if m == "" {
		return nil
	}
	f, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return nil
	}
	return f
}

// DateToISO parses an RFC3339 or plain date and returns its ISO form, or nil.
func DateToISO(s string) any {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format(time.RFC3339)
		}
	}
	return nil
}

// Trim removes leading and trailing whitespace.
func Trim(s string) string { return strings.TrimSpace(s) }

// Unescape decodes HTML character references.
func Unescape(s string) string { return stdhtml.UnescapeString(s) }

// StripHTML returns the text nodes of s with whitespace cleaned.
func StripHTML(s string) string {
	z := xhtml.NewTokenizer(strings.NewReader(s))
	var b strings.Builder
	for {
		switch z.Next() {
		case xhtml.ErrorToken:
			return Clean(b.String())
		case xhtml.TextToken:
			b.Write(z.Text())
		}
	}
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
