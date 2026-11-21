package config

import (
	"bytes"
	"fmt"
	"os"

	ini "unknwon.dev/ini/v2"
)

// Projection represents a resolved configuration knob for status/ui inspection.
type Projection struct {
	Section string `json:"section"`
	Key     string `json:"key"`
	FullKey string `json:"fullKey"`
	Doc     string `json:"doc"`
	Type    string `json:"type"`
	Flag    string `json:"flag"`
	Env     string `json:"env"`
	Default string `json:"default"`
	Value   string `json:"value"`
	Source  string `json:"source"`
	Invalid bool   `json:"invalid,omitempty"`
}

// Project re-reads the ini and resolves default<ini<env (no writes), setting
// Invalid=true, Value=Default, Source="default" only for the offending knob.
func Project(r *Registry, path string) ([]Projection, error) {
	iniFile, err := LoadIni(path)
	if err != nil {
		return nil, err
	}
	entries := r.Entries()
	projs := make([]Projection, len(entries))

	for i, e := range entries {
		fullKey := e.Name()
		val := e.Default
		source := "default"
		invalid := false

		iniVal := iniFile.Get(e.Section, e.Key)
		if iniVal != "" {
			if _, err := r.Validate(fullKey, iniVal); err != nil {
				invalid = true
			} else {
				val = iniVal
				source = "ini"
			}
		}

		envVal := os.Getenv(e.EnvName())
		if envVal != "" {
			if _, err := r.Validate(fullKey, envVal); err != nil {
				invalid = true
			} else {
				val = envVal
				source = "env"
			}
		}

		if invalid {
			val = e.Default
			source = "default"
		}

		projs[i] = Projection{
			Section: e.Section,
			Key:     e.Key,
			FullKey: fullKey,
			Doc:     e.Doc,
			Type:    e.Kind.String(),
			Flag:    e.FlagName(),
			Env:     e.EnvName(),
			Default: e.Default,
			Value:   val,
			Source:  source,
			Invalid: invalid,
		}
	}
	return projs, nil
}

// MergeSave validates every "section.key", creates the file from the dump template
// via atomic Save when absent, sets only touched keys on the existing Ini, calls Save,
// and returns written count; errors name the offending key.
func MergeSave(r *Registry, path string, values map[string]string) (int, error) {
	for k := range values {
		if _, ok := r.Lookup(k); !ok {
			return 0, fmt.Errorf("unknown key %q", k)
		}
	}
	for k, v := range values {
		if _, err := r.Validate(k, v); err != nil {
			return 0, fmt.Errorf("invalid value for %s: %w", k, err)
		}
	}

	var iniFile *Ini
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		var buf bytes.Buffer
		if err := r.Dump(&buf); err != nil {
			return 0, err
		}
		f, err := ini.LoadSources([]any{buf.Bytes()}, loadOptions...)
		if err != nil {
			return 0, err
		}
		iniFile = &Ini{f: f}
	} else {
		var err error
		iniFile, err = LoadIni(path)
		if err != nil {
			return 0, err
		}
	}

	for k, v := range values {
		e, _ := r.Lookup(k)
		iniFile.Set(e.Section, e.Key, v)
	}

	if err := iniFile.Save(path); err != nil {
		return 0, err
	}
	return len(values), nil
}
