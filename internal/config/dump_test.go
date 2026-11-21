package config

import (
	"strings"
	"testing"
)

func TestDumpListsEveryKnobOnce(t *testing.T) {
	reg := Knobs()
	out := reg.DumpString()
	for _, e := range reg.Entries() {
		if got := strings.Count(out, "\n"+e.Key+" ="); got != 1 {
			t.Errorf("key %s appears %d times, want 1\ndump:\n%s", e.Name(), got, out)
		}
	}
}

func TestDumpDocAndSection(t *testing.T) {
	reg := Knobs()
	out := reg.DumpString()
	for _, e := range reg.Entries() {
		if n := strings.Count(out, "["+e.Section+"]\n"); n != 1 {
			t.Errorf("section %s header appears %d times, want 1", e.Section, n)
		}
		want := "; " + e.Doc + "\n" + e.Key + " = " + e.Default + "\n"
		if e.Default == "" {
			want = "; " + e.Doc + "\n" + e.Key + " =\n"
		}
		if !strings.Contains(out, want) {
			t.Errorf("missing doc+default line for %s, want %q", e.Name(), want)
		}
	}
}

func TestDumpRoundTripsThroughIni(t *testing.T) {
	reg := Knobs()
	path := t.TempDir() + "/dump.ini"
	first, err := LoadIni(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range reg.Entries() {
		first.Set(e.Section, e.Key, e.Default)
	}
	if err := first.Save(path); err != nil {
		t.Fatal(err)
	}
	// Every dumped default must resolve to itself with source ini.
	loaded, err := LoadIni(path)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := Resolve(reg, loaded, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, res := range snap.All() {
		if res.Entry.Default == "" {
			// An empty default cannot be written as an ini value: an empty
			// value counts as unset, so it resolves back to the default.
			if res.Source != SourceDefault {
				t.Errorf("%s source = %s, want default", res.Entry.Name(), res.Source)
			}
			continue
		}
		if res.Source != SourceIni {
			t.Errorf("%s source = %s, want ini", res.Entry.Name(), res.Source)
			continue
		}
		if res.Value.String() != res.Entry.Default {
			t.Errorf("%s = %q, want %q", res.Entry.Name(), res.Value.String(), res.Entry.Default)
		}
	}
}
