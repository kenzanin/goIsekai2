package config

import "fmt"

// Source names the layer that produced a resolved configuration value.
type Source string

const (
	// SourceDefault is the value declared in the knob table.
	SourceDefault Source = "default"
	// SourceIni is a value read from the INI document.
	SourceIni Source = "ini"
	// SourceEnv is a value read from the process environment.
	SourceEnv Source = "env"
	// SourceFlag is a value given on the command line.
	SourceFlag Source = "flag"
)

// Resolved is one configuration entry paired with its winning value,
// the layer it came from and the raw text that produced the value.
type Resolved struct {
	Entry  Entry
	Value  Value
	Source Source
	Raw    string
}

// Snapshot is the fully resolved configuration: one Resolved per registry
// entry, in registry order.
type Snapshot struct {
	items []Resolved
}

// Resolve settles every registry entry to a single value by applying the
// default, INI, environment and flag layers in rising precedence; the last
// layer carrying a value wins. An empty INI value counts as unset. A nil
// ini, env or flagValue skips that layer.
func Resolve(r *Registry, ini *Ini, env func(string) string, flagValue func(Entry) (string, bool)) (*Snapshot, error) {
	entries := r.Entries()
	s := &Snapshot{items: make([]Resolved, 0, len(entries))}
	for _, e := range entries {
		res := Resolved{Entry: e}
		apply := func(src Source, raw string) error {
			v, err := r.Validate(e.Name(), raw)
			if err != nil {
				return fmt.Errorf("config: %s: invalid %s value %q: %w", e.Name(), src, raw, err)
			}
			res.Value, res.Source, res.Raw = v, src, raw
			return nil
		}
		if err := apply(SourceDefault, e.Default); err != nil {
			return nil, err
		}
		if ini != nil {
			if v := ini.Get(e.Section, e.Key); v != "" {
				if err := apply(SourceIni, v); err != nil {
					return nil, err
				}
			}
		}
		if env != nil {
			if v := env(e.EnvName()); v != "" {
				if err := apply(SourceEnv, v); err != nil {
					return nil, err
				}
			}
		}
		if flagValue != nil {
			if v, ok := flagValue(e); ok {
				if err := apply(SourceFlag, v); err != nil {
					return nil, err
				}
			}
		}
		s.items = append(s.items, res)
	}
	return s, nil
}

// Get returns the resolved value for the named entry.
func (s *Snapshot) Get(name string) (Resolved, bool) {
	for _, res := range s.items {
		if res.Entry.Name() == name {
			return res, true
		}
	}
	return Resolved{}, false
}

// All returns every resolved entry in registry order.
func (s *Snapshot) All() []Resolved {
	out := make([]Resolved, len(s.items))
	copy(out, s.items)
	return out
}

// String returns the value of name as a string, or "" when absent.
func (s *Snapshot) String(name string) string {
	if res, ok := s.Get(name); ok {
		return res.Value.String()
	}
	return ""
}

// Int returns the value of name as an integer, or 0 when absent.
func (s *Snapshot) Int(name string) int64 {
	if res, ok := s.Get(name); ok {
		return res.Value.Int
	}
	return 0
}

// Bool returns the value of name as a boolean, or false when absent.
func (s *Snapshot) Bool(name string) bool {
	if res, ok := s.Get(name); ok {
		return res.Value.B
	}
	return false
}
