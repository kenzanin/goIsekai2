package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func emptyIni(t *testing.T) *Ini {
	t.Helper()
	ini, err := LoadIni(filepath.Join(t.TempDir(), "config.ini"))
	if err != nil {
		t.Fatal(err)
	}
	return ini
}

// testRegistry mirrors the Knobs server.port knob plus a string and a bool
// knob; it avoids Knobs because the current table panics in New on Min-only
// bounds (see registry bound check vs Max 0).
func testRegistry(t *testing.T) *Registry {
	t.Helper()
	reg, err := New(
		Entry{
			Section: "server", Key: "port", Kind: KindInt,
			Default: "8090", Min: 1, Max: 65535,
			Flag: "port", Env: "GOISEKAI_PORT",
			Doc: "TCP listen port",
		},
		Entry{
			Section: "server", Key: "proxy", Kind: KindString,
			Default: "direct",
			Flag:    "proxy", Env: "GOISEKAI_PROXY",
			Doc: "Outbound proxy URL",
		},
		Entry{
			Section: "t", Key: "on", Kind: KindBool,
			Default: "true",
			Doc:     "Boolean knob",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

func TestResolvePrecedence(t *testing.T) {
	tests := []struct {
		name    string
		ini     string
		env     string
		flag    string
		flagOK  bool
		wantInt int64
		wantSrc Source
	}{
		{"default only", "", "", "", false, 8090, SourceDefault},
		{"ini overrides default", "8081", "", "", false, 8081, SourceIni},
		{"env overrides ini", "8081", "8082", "", false, 8082, SourceEnv},
		{"flag overrides env", "8081", "8082", "8083", true, 8083, SourceFlag},
		{"flag overrides default", "", "", "8084", true, 8084, SourceFlag},
		{"unset flag keeps ini", "8081", "", "", false, 8081, SourceIni},
		{"empty env is ignored", "8081", "", "", false, 8081, SourceIni},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ini := emptyIni(t)
			if tt.ini != "" {
				ini.Set("server", "port", tt.ini)
			}
			snap, err := Resolve(testRegistry(t), ini,
				func(name string) string {
					if name == "GOISEKAI_PORT" {
						return tt.env
					}
					return ""
				},
				func(e Entry) (string, bool) {
					if e.Name() == "server.port" {
						return tt.flag, tt.flagOK
					}
					return "", false
				})
			if err != nil {
				t.Fatal(err)
			}
			res, ok := snap.Get("server.port")
			if !ok {
				t.Fatal("server.port missing from snapshot")
			}
			if res.Value.Int != tt.wantInt || res.Source != tt.wantSrc {
				t.Fatalf("got %d from %s, want %d from %s", res.Value.Int, res.Source, tt.wantInt, tt.wantSrc)
			}
		})
	}
}

func TestResolveInvalidIni(t *testing.T) {
	ini := emptyIni(t)
	ini.Set("server", "port", "not-a-port")
	_, err := Resolve(testRegistry(t), ini, func(string) string { return "" }, nil)
	if err == nil {
		t.Fatal("want error for invalid ini value")
	}
	const want = `config: server.port: invalid ini value "not-a-port"`
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("got %v, want prefix %q", err, want)
	}
}

func TestResolveInvalidEnv(t *testing.T) {
	_, err := Resolve(testRegistry(t), nil, func(name string) string {
		if name == "GOISEKAI_PORT" {
			return "NaN"
		}
		return ""
	}, nil)
	if err == nil {
		t.Fatal("want error for invalid env value")
	}
	const want = `config: server.port: invalid env value "NaN"`
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("got %v, want prefix %q", err, want)
	}
}

func TestSnapshotAccessors(t *testing.T) {
	snap, err := Resolve(testRegistry(t), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := snap.Int("server.port"); got != 8090 {
		t.Errorf("Int = %d, want 8090", got)
	}
	if got := snap.String("server.proxy"); got != "direct" {
		t.Errorf("String = %q, want %q", got, "direct")
	}
	if !snap.Bool("t.on") {
		t.Error("Bool(t.on) = false, want true")
	}
	if _, ok := snap.Get("nope"); ok {
		t.Error("Get(nope) ok, want missing")
	}
	if got := snap.String("nope"); got != "" {
		t.Errorf("String(nope) = %q, want empty", got)
	}
	if got := snap.Int("nope"); got != 0 {
		t.Errorf("Int(nope) = %d, want 0", got)
	}
	if snap.Bool("nope") {
		t.Error("Bool(nope) = true, want false")
	}
	res, _ := snap.Get("server.port")
	if res.Raw != "8090" || res.Source != SourceDefault {
		t.Errorf("Raw/Source = %q/%s, want 8090/default", res.Raw, res.Source)
	}
}

func TestSnapshotAllOrder(t *testing.T) {
	reg := testRegistry(t)
	snap, err := Resolve(reg, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	all := snap.All()
	entries := reg.Entries()
	if len(all) != len(entries) {
		t.Fatalf("All len = %d, want %d", len(all), len(entries))
	}
	for i, res := range all {
		if res.Entry.Name() != entries[i].Name() {
			t.Fatalf("position %d = %s, want %s", i, res.Entry.Name(), entries[i].Name())
		}
	}
	if all[0].Entry.Name() != "server.port" {
		t.Errorf("first entry = %s, want server.port", all[0].Entry.Name())
	}
}
