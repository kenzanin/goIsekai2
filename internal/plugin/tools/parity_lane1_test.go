package tools

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAliasResolution(t *testing.T) {
	cases := []struct {
		name, canonical, fixed string
	}{
		{"codecs.base64_encode", "crypto.b64enc", ""},
		{"crypto.sha256_hex", "crypto.hash", "sha256"},
		{"text.clean", "text.clean", ""},
	}
	for _, tc := range cases {
		gotName, gotFixed := resolveAlias(tc.name)
		if gotName != tc.canonical || gotFixed != tc.fixed {
			t.Errorf("resolveAlias(%q) = (%q, %q), want (%q, %q)", tc.name, gotName, gotFixed, tc.canonical, tc.fixed)
		}
	}
}

func TestAliasEqualsCanonical(t *testing.T) {
	ts := &ToolSet{}
	alias, err := ts.Dispatch("codecs.base64_encode", "hello")
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := ts.Dispatch("crypto.b64enc", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if alias != canonical {
		t.Fatalf("alias = %v, canonical = %v", alias, canonical)
	}
	alias, err = ts.Dispatch("crypto.sha256_hex", "hello")
	if err != nil {
		t.Fatal(err)
	}
	canonical, err = ts.Dispatch("crypto.hash", "sha256", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if alias != canonical {
		t.Fatalf("sha256 alias = %v, canonical = %v", alias, canonical)
	}
}

func TestNewBaselineNames(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "lib"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "lib", "shared.lua"), []byte("return 1"), 0o600); err != nil {
		t.Fatal(err)
	}
	ts := &ToolSet{Dir: dir, Params: map[string]any{"api": "demo"}}

	cases := []struct {
		name string
		args []any
		want any
	}{
		{"text.normalize_status", []any{"Ongoing"}, "ongoing"},
		{"plugin.param", []any{"api"}, "demo"},
		{"text.chapter_num", []any{"Chapter 12.5"}, 12.5},
		{"text.date_to_iso", []any{"2026-11-23"}, "2026-11-23T00:00:00Z"},
		{"text.trim", []any{"  hi  "}, "hi"},
		{"text.unescape", []any{"a&amp;b"}, "a&b"},
		{"text.strip_html", []any{"<p>Hello <b>world</b></p>"}, "Hello world"},
		{"fs.read_lib", []any{"shared.lua"}, "return 1"},
		{"log.debug", []any{"trace"}, nil},
	}
	for _, tc := range cases {
		got, err := ts.Dispatch(tc.name, tc.args...)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s = %#v, want %#v", tc.name, got, tc.want)
		}
	}

	if _, err := ts.Dispatch("fs.read_lib", "../secrets"); err == nil {
		t.Fatal("fs.read_lib escape unexpectedly succeeded")
	}

	matches, err := ts.Dispatch("re2.gmatch", `(\d+)`, "a1 b22")
	if err != nil {
		t.Fatal(err)
	}
	rows, ok := matches.([][]string)
	if !ok || len(rows) != 2 || rows[0][1] != "1" || rows[1][1] != "22" {
		t.Fatalf("re2.gmatch = %#v", matches)
	}
}
