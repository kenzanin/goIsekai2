package config

import (
	"os"
	"path/filepath"

	ini "unknwon.dev/ini/v2"
)

// loadOptions keep inline comments attached to their key and stop a value
// containing a comment symbol from being truncated on read.
var loadOptions = []ini.LoadOption{
	ini.IgnoreInlineComment(false),
	ini.SpaceBeforeInlineComment(true),
}

// Ini is a handle to an INI document held in memory.
type Ini struct {
	f *ini.File
}

// LoadIni parses the INI file at path, keeping comments and unknown keys; a
// missing file yields an empty document and nil error.
func LoadIni(path string) (*Ini, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return &Ini{f: ini.Empty(loadOptions...)}, nil
		}
		return nil, err
	}
	f, err := ini.LoadSources([]any{path}, loadOptions...)
	if err != nil {
		return nil, err
	}
	return &Ini{f: f}, nil
}

// Get returns the value of key in section, or the empty string when either is
// absent; the empty section name addresses the DEFAULT section.
func (i *Ini) Get(section, key string) string {
	return i.f.Section(section).Key(key).String()
}

// Set stores value for key in section, creating the key or the section when
// missing and overwriting the value when the key already exists.
func (i *Ini) Set(section, key, value string) {
	// ponytail: NewKey's empty-name error is dropped; Set has no error return.
	_, _ = i.f.Section(section).NewKey(key, value)
}

// Save writes the document to path atomically: the content goes to a 0600
// temp file in the same directory, is synced, closed and renamed over path,
// and the containing directory is fsynced so the rename is durable; the temp
// file is removed if any step fails.
func (i *Ini) Save(path string) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".ini-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	cleanup := func() {
		tmp.Close()
		os.Remove(name)
	}
	if err := tmp.Chmod(0o600); err != nil {
		cleanup()
		return err
	}
	if _, err := i.f.WriteTo(tmp); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Rename(name, path); err != nil {
		os.Remove(name)
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	if err := d.Sync(); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}
