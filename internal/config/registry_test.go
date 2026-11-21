package config

import (
	"strings"
	"testing"
)

func TestKindString(t *testing.T) {
	tests := []struct {
		kind Kind
		want string
	}{
		{KindInt, "int"},
		{KindInt64, "int64"},
		{KindBool, "bool"},
		{KindString, "string"},
		{KindEnum, "enum"},
		{Kind(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("Kind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestEntryName(t *testing.T) {
	tests := []struct {
		entry Entry
		want  string
	}{
		{Entry{Section: "server", Key: "timeout"}, "server.timeout"},
		{Entry{Section: "db_pool", Key: "max_size"}, "db_pool.max_size"},
	}
	for _, tt := range tests {
		if got := tt.entry.Name(); got != tt.want {
			t.Errorf("Name() = %q, want %q", got, tt.want)
		}
	}
}

func TestFlagName(t *testing.T) {
	tests := []struct {
		name  string
		entry Entry
		want  string
	}{
		{"default", Entry{Section: "server", Key: "read_timeout"}, "--server-read-timeout"},
		{"custom", Entry{Section: "server", Key: "read_timeout", Flag: "-r"}, "-r"},
		{"underscores to dashes", Entry{Section: "log_level", Key: "default_val"}, "--log-level-default-val"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entry.FlagName(); got != tt.want {
				t.Errorf("FlagName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnvName(t *testing.T) {
	tests := []struct {
		name  string
		entry Entry
		want  string
	}{
		{"default", Entry{Section: "server", Key: "timeout"}, "GOISEKAI_SERVER_TIMEOUT"},
		{"custom", Entry{Section: "server", Key: "timeout", Env: "MY_ENV"}, "MY_ENV"},
		{"dots and dashes", Entry{Section: "a-b.c", Key: "d_e"}, "GOISEKAI_A_B_C_D_E"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entry.EnvName(); got != tt.want {
				t.Errorf("EnvName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		r, err := New(Entry{Section: "s", Key: "k", Kind: KindBool, Default: "true", Doc: "d"})
		if err != nil || r == nil {
			t.Fatalf("New() err = %v, want nil", err)
		}
	})
	t.Run("duplicate name", func(t *testing.T) {
		_, err := New(
			Entry{Section: "s", Key: "k", Kind: KindBool, Doc: "d"},
			Entry{Section: "s", Key: "k", Kind: KindString, Doc: "d2"},
		)
		if err == nil || !strings.Contains(err.Error(), `duplicate entry name "s.k"`) {
			t.Errorf("err = %v, want duplicate entry name", err)
		}
	})
	t.Run("empty doc", func(t *testing.T) {
		_, err := New(Entry{Section: "s", Key: "k", Kind: KindBool, Doc: ""})
		if err == nil || !strings.Contains(err.Error(), "doc cannot be empty") {
			t.Errorf("err = %v, want empty doc error", err)
		}
	})
	t.Run("enum without allowed", func(t *testing.T) {
		_, err := New(Entry{Section: "s", Key: "k", Kind: KindEnum, Doc: "d"})
		if err == nil || !strings.Contains(err.Error(), "requires allowed values") {
			t.Errorf("err = %v, want enum error", err)
		}
	})
	t.Run("invalid default", func(t *testing.T) {
		_, err := New(Entry{Section: "s", Key: "k", Kind: KindInt, Default: "nope", Doc: "d"})
		if err == nil || !strings.Contains(err.Error(), `invalid default "nope"`) {
			t.Errorf("err = %v, want invalid default error", err)
		}
	})
}

func TestLookup(t *testing.T) {
	r, err := New(
		Entry{Section: "a", Key: "x", Kind: KindInt, Default: "1", Doc: "d"},
		Entry{Section: "a", Key: "y", Kind: KindBool, Doc: "d"},
	)
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	got, ok := r.Lookup("a.x")
	if !ok || got.Name() != "a.x" || got.Kind != KindInt {
		t.Errorf("Lookup hit: got (%v, %v), want a.x entry", got, ok)
	}
	if _, ok := r.Lookup("a.z"); ok {
		t.Error("Lookup miss: ok = true, want false")
	}
}

func TestDefault(t *testing.T) {
	r, err := New(Entry{Section: "s", Key: "n", Kind: KindInt, Default: "42", Doc: "d"})
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	v, err := r.Default("s.n")
	if err != nil || v.Int != 42 || v.Kind != KindInt {
		t.Errorf("Default() = (%v, %v), want int 42", v, err)
	}
	if _, err := r.Default("s.missing"); err == nil {
		t.Error("Default miss: err = nil, want error")
	}
}

func TestValidateIntBounds(t *testing.T) {
	r, err := New(Entry{Section: "s", Key: "n", Kind: KindInt, Min: 1, Max: 10, Default: "5", Doc: "d"})
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	for _, tt := range []struct {
		raw  string
		want int64
		bad  bool
	}{
		{"1", 1, false},
		{"10", 10, false},
		{"5", 5, false},
		{"0", 0, true},
		{"11", 11, true},
		{"abc", 0, true},
	} {
		v, err := r.Validate("s.n", tt.raw)
		if tt.bad {
			if err == nil {
				t.Errorf("Validate(%q) err = nil, want error", tt.raw)
			}
			continue
		}
		if err != nil || v.Int != tt.want {
			t.Errorf("Validate(%q) = (%v, %v), want int %d", tt.raw, v, err, tt.want)
		}
	}
}

func TestValidateBool(t *testing.T) {
	r, err := New(Entry{Section: "s", Key: "b", Kind: KindBool, Default: "true", Doc: "d"})
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	for _, tt := range []struct {
		raw  string
		want bool
		bad  bool
	}{
		{"true", true, false},
		{"1", true, false},
		{"TRUE", true, false},
		{"false", false, false},
		{"0", false, false},
		{"False", false, false},
		{"yes", false, true},
		{"", true, false}, // empty falls back to Default "true"
	} {
		v, err := r.Validate("s.b", tt.raw)
		if tt.bad {
			if err == nil {
				t.Errorf("Validate(%q) err = nil, want error", tt.raw)
			}
			continue
		}
		if err != nil || v.B != tt.want || v.Kind != KindBool {
			t.Errorf("Validate(%q) = (%v, %v), want bool %v", tt.raw, v, err, tt.want)
		}
	}
}

func TestValidateEnum(t *testing.T) {
	r, err := New(Entry{Section: "log", Key: "level", Kind: KindEnum, Allowed: []string{"debug", "info"}, Default: "info", Doc: "d"})
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	v, err := r.Validate("log.level", "debug")
	if err != nil || v.Str != "debug" || v.Kind != KindEnum {
		t.Errorf("allowed: (%v, %v), want debug", v, err)
	}
	if _, err := r.Validate("log.level", "warn"); err == nil {
		t.Error("rejected: err = nil, want error")
	}
}

func TestValidateString(t *testing.T) {
	r, err := New(Entry{Section: "s", Key: "str", Kind: KindString, Default: "x", Doc: "d"})
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	v, err := r.Validate("s.str", " hello ")
	if err != nil || v.Str != "hello" || v.Kind != KindString {
		t.Errorf("(%v, %v), want trimmed hello", v, err)
	}
}

func TestValidateUnknownKind(t *testing.T) {
	r, err := New(Entry{Section: "s", Key: "u", Kind: Kind(42), Doc: "d"})
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	if _, err := r.Validate("s.u", "x"); err == nil || !strings.Contains(err.Error(), "unknown kind") {
		t.Errorf("err = %v, want unknown kind", err)
	}
}

func TestEntryNameRoundTrip(t *testing.T) {
	tests := []struct {
		section, key string
	}{
		{"server", "timeout"},
		{"db", "max_conns"},
		{"a", "b"},
	}
	for _, tt := range tests {
		e := Entry{Section: tt.section, Key: tt.key}
		name := e.Name()
		got, ok := func() (Entry, bool) {
			return e, name == tt.section+"."+tt.key
		}()
		if !ok {
			t.Fatalf("round-trip build failed for %q", name)
		}
		if got.Section != tt.section || got.Key != tt.key {
			t.Errorf("round-trip %q: got section=%q key=%q", name, got.Section, got.Key)
		}
		re := Entry{Section: tt.section, Key: tt.key}
		if re.Name() != name {
			t.Errorf("round-trip: rebuilt %q != %q", re.Name(), name)
		}
	}
}
