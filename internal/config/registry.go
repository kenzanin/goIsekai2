package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Kind represents the data type of a configuration entry.
type Kind uint8

const (
	// KindInt is a 32-bit signed integer knob.
	KindInt Kind = iota
	// KindInt64 is a 64-bit signed integer knob, used for byte and millisecond limits.
	KindInt64
	// KindBool is a boolean knob.
	KindBool
	// KindString is a free-form string knob.
	KindString
	// KindEnum is a string knob restricted to a fixed set of allowed values.
	KindEnum
)

// String returns the string representation of the Kind.
func (k Kind) String() string {
	switch k {
	case KindInt:
		return "int"
	case KindInt64:
		return "int64"
	case KindBool:
		return "bool"
	case KindString:
		return "string"
	case KindEnum:
		return "enum"
	}
	return "unknown"
}

// Entry defines a configuration setting's metadata, constraints, and defaults.
type Entry struct {
	Section string
	Key     string
	Kind    Kind
	Default string
	Min     int64
	Max     int64
	Allowed []string
	Flag    string
	Env     string
	Doc     string
}

// Name returns the dot-separated section and key name.
func (e Entry) Name() string {
	return e.Section + "." + e.Key
}

// FlagName returns the custom flag name or generates one from section and key.
func (e Entry) FlagName() string {
	if e.Flag != "" {
		return e.Flag
	}
	return "--" + strings.ReplaceAll(e.Section, "_", "-") + "-" + strings.ReplaceAll(e.Key, "_", "-")
}

// EnvName returns the custom environment variable name or generates one from section and key.
func (e Entry) EnvName() string {
	if e.Env != "" {
		return e.Env
	}
	s := strings.ToUpper(e.Section + "_" + e.Key)
	return "GOISEKAI_" + strings.ReplaceAll(strings.ReplaceAll(s, ".", "_"), "-", "_")
}

// Value holds a parsed configuration value of a specific kind.
type Value struct {
	Kind Kind
	Int  int64
	Str  string
	B    bool
}

// String returns the canonical text representation of the Value.
func (v Value) String() string {
	switch v.Kind {
	case KindInt, KindInt64:
		return strconv.FormatInt(v.Int, 10)
	case KindBool:
		return strconv.FormatBool(v.B)
	case KindString, KindEnum:
		return v.Str
	}
	return ""
}

// Registry stores configuration entries and provides lookup and validation.
type Registry struct {
	entries []Entry
	index   map[string]int
}

// New creates and validates a new Registry from a slice of entries.
func New(entries ...Entry) (*Registry, error) {
	index := make(map[string]int, len(entries))
	for i, e := range entries {
		if e.Section == "" || e.Key == "" || e.Doc == "" {
			return nil, fmt.Errorf("entry %d: section, key, and doc cannot be empty", i)
		}
		name := e.Name()
		if _, exists := index[name]; exists {
			return nil, fmt.Errorf("duplicate entry name %q", name)
		}
		if e.Kind == KindEnum && len(e.Allowed) == 0 {
			return nil, fmt.Errorf("entry %q: enum kind requires allowed values", name)
		}
		if err := validateRaw(e, e.Default); err != nil && e.Default != "" {
			return nil, fmt.Errorf("entry %q: invalid default %q: %w", name, e.Default, err)
		}
		index[name] = i
	}
	cp := make([]Entry, len(entries))
	copy(cp, entries)
	return &Registry{entries: cp, index: index}, nil
}

// Lookup finds an entry by its full name.
func (r *Registry) Lookup(name string) (Entry, bool) {
	if r == nil {
		return Entry{}, false
	}
	idx, ok := r.index[name]
	if !ok {
		return Entry{}, false
	}
	return r.entries[idx], true
}

// Entries returns a copy of all entries in the registry.
func (r *Registry) Entries() []Entry {
	if r == nil {
		return nil
	}
	cp := make([]Entry, len(r.entries))
	copy(cp, r.entries)
	return cp
}

// Default returns the parsed default value for the named entry.
func (r *Registry) Default(name string) (Value, error) {
	e, ok := r.Lookup(name)
	if !ok {
		return Value{}, fmt.Errorf("entry %q not found", name)
	}
	return r.Validate(name, e.Default)
}

// Validate parses and validates a raw string value against the named entry's constraints.
func (r *Registry) Validate(name, raw string) (Value, error) {
	e, ok := r.Lookup(name)
	if !ok {
		return Value{}, fmt.Errorf("entry %q not found", name)
	}
	return parseValue(e, raw)
}

func validateRaw(e Entry, raw string) error {
	_, err := parseValue(e, raw)
	return err
}

func parseValue(e Entry, raw string) (Value, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = e.Default
	}
	name := e.Name()

	switch e.Kind {
	case KindInt, KindInt64:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return Value{}, fmt.Errorf("entry %q: invalid value %q for %s: %w", name, raw, e.Kind, err)
		}
		if (e.Min != 0 || e.Max != 0) && (v < e.Min || v > e.Max) {
			return Value{}, fmt.Errorf("entry %q: value %d out of bounds [%d, %d]", name, v, e.Min, e.Max)
		}
		return Value{Kind: e.Kind, Int: v}, nil

	case KindBool:
		switch strings.ToLower(raw) {
		case "true", "1":
			return Value{Kind: e.Kind, B: true}, nil
		case "false", "0":
			return Value{Kind: e.Kind, B: false}, nil
		default:
			return Value{}, fmt.Errorf("entry %q: invalid boolean value %q", name, raw)
		}

	case KindString:
		return Value{Kind: e.Kind, Str: raw}, nil

	case KindEnum:
		for _, a := range e.Allowed {
			if raw == a {
				return Value{Kind: e.Kind, Str: raw}, nil
			}
		}
		return Value{}, fmt.Errorf("entry %q: invalid enum value %q, expected one of %v", name, raw, e.Allowed)
	}

	return Value{}, fmt.Errorf("entry %q: unknown kind %v", name, e.Kind)
}
