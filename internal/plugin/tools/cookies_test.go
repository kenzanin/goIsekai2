package tools

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestJarPathScopesPlugins(t *testing.T) {
	dir := t.TempDir()
	a := &ToolSet{DataDir: dir, PluginID: "alpha"}
	b := &ToolSet{DataDir: dir, PluginID: "beta"}
	if want := filepath.Join(dir, "plugin-state", "alpha.cookies.txt"); a.JarPath() != want {
		t.Fatalf("JarPath() = %q, want %q", a.JarPath(), want)
	}
	if a.JarPath() == b.JarPath() {
		t.Fatal("distinct plugin ids share one jar path")
	}
}

func TestCookieRoundTripAndIsolation(t *testing.T) {
	dir := t.TempDir()
	a := &ToolSet{DataDir: dir, PluginID: "alpha"}
	b := &ToolSet{DataDir: dir, PluginID: "beta"}

	jarA := a.LoadJar()
	exU, _ := url.Parse("https://example.com/")
	otU, _ := url.Parse("https://other.org/")
	jarA.SetCookies(exU, []*http.Cookie{{Name: "sid", Value: "abc"}})
	jarA.SetCookies(otU, []*http.Cookie{{Name: "t", Value: "1"}})
	if err := a.SaveJar(jarA); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(a.JarPath())
	if err != nil {
		t.Fatal(err)
	}
	if want := "example.com\tsid\tabc\nother.org\tt\t1\n"; string(raw) != want {
		t.Fatalf("saved file = %q, want %q", raw, want)
	}

	back := a.LoadJar()
	cs := back.Cookies(exU)
	if len(cs) != 1 || cs[0].Name != "sid" || cs[0].Value != "abc" {
		t.Fatalf("round trip lost cookie: %v", cs)
	}

	// Beta's own empty jar must not see alpha's cookies.
	if err := b.SaveJar(b.LoadJar()); err != nil {
		t.Fatal(err)
	}
	if cs := b.LoadJar().Cookies(exU); len(cs) != 0 {
		t.Fatalf("cookie leaked across plugins: %v", cs)
	}
	if raw, err := os.ReadFile(b.JarPath()); err != nil || len(raw) != 0 {
		t.Fatalf("beta state = %q, %v; want empty", raw, err)
	}

	// A loaded jar re-saved keeps its entries (hosts recorded on load).
	if err := a.SaveJar(a.LoadJar()); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(a.JarPath())
	if err != nil {
		t.Fatal(err)
	}
	if want := "example.com\tsid\tabc\nother.org\tt\t1\n"; string(raw) != want {
		t.Fatalf("re-saved file = %q, want %q", raw, want)
	}
}
