package tools

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// JarPath returns the plugin-scoped cookie state file, the
// "host<TAB>name<TAB>value" lines written by SaveJar.
func (t *ToolSet) JarPath() string {
	return filepath.Join(t.DataDir, "plugin-state", t.PluginID+".cookies.txt")
}

// cookieJar is the jar LoadJar returns: the stdlib jar plus the host set
// SaveJar persists, since cookiejar.Jar cannot enumerate its own contents.
type cookieJar struct {
	*cookiejar.Jar
	hosts map[string]struct{}
}

// SetCookies records the URL host for the next SaveJar, then delegates to the
// wrapped jar.
func (j *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if u != nil && u.Host != "" {
		j.hosts[u.Host] = struct{}{}
	}
	j.Jar.SetCookies(u, cookies)
}

// LoadJar returns a cookie jar holding the cookies of the last SaveJar, one
// "host<TAB>name<TAB>value" line per cookie. A missing state file yields an
// empty jar; an unreadable one only logs. Every restored cookie is host-scoped
// on path "/" — the format keeps no path or expiry.
func (t *ToolSet) LoadJar() http.CookieJar {
	jar, _ := cookiejar.New(nil) // never fails with nil options
	j := &cookieJar{Jar: jar, hosts: map[string]struct{}{}}
	raw, err := os.ReadFile(t.JarPath())
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			t.log(slog.LevelWarn, "cookie state unreadable: "+err.Error())
		}
		return j
	}
	for _, line := range strings.Split(string(raw), "\n") {
		f := strings.Split(line, "\t")
		if len(f) != 3 || f[0] == "" || f[1] == "" {
			continue
		}
		j.SetCookies(&url.URL{Scheme: "https", Host: strings.Trim(f[0], "."), Path: "/"},
			[]*http.Cookie{{Name: f[1], Value: f[2]}})
	}
	return j
}

// SaveJar writes every cookie jar holds back to JarPath, one
// "host<TAB>name<TAB>value" line per cookie, creating the state dir. Jar must
// come from LoadJar; anything else is an error, as is a failed write.
func (t *ToolSet) SaveJar(jar http.CookieJar) error {
	j, ok := jar.(*cookieJar)
	if !ok {
		return fmt.Errorf("cookie jar must come from LoadJar, got %T", jar)
	}
	var b strings.Builder
	for _, host := range slices.Sorted(maps.Keys(j.hosts)) {
		for _, c := range j.Cookies(&url.URL{Scheme: "https", Host: host, Path: "/"}) {
			// ponytail: a tab or newline inside name/value corrupts the TSV;
			// escape here when a real plugin hits it.
			b.WriteString(host + "\t" + c.Name + "\t" + c.Value + "\n")
		}
	}
	if err := os.MkdirAll(filepath.Dir(t.JarPath()), 0o755); err != nil {
		return fmt.Errorf("cookie state dir: %w", err)
	}
	return os.WriteFile(t.JarPath(), []byte(b.String()), 0o644)
}
