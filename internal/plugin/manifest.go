// Package plugin loads manga-site scraper plugins: one folder per site under
// the plugins root, a main.json manifest in each, addressed by folder id.
//
// The manifest contract and folder layout mirror cppIsekai exactly; the
// runtime backends (lua/js/wasm) live in later work packages.
package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// runtimes is the set of manifest runtime values this host can execute.
var runtimes = map[string]bool{"lua": true, "js": true, "wasm": true}

// Manifest is the plugin main.json contract. Field names match the cppIsekai
// manifest ABI; logo is a host-side addition absent from the C++ schema.
type Manifest struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Runtime      string   `json:"runtime"`
	Capabilities []string `json:"capabilities"`
	API          int      `json:"api"`
	Entry        string   `json:"entry"`
	Logo         string   `json:"logo"`
	Website      string   `json:"website"`
	ImageReferer string   `json:"image_referer"`
}

// LoadManifest reads and validates the manifest file at path. Entry resolves
// against the manifest's own folder. It rejects api != 1, a runtime outside
// lua/js/wasm, a declared id differing from the folder name, and a missing
// entry file; every error names path.
func LoadManifest(path string) (*Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if base := filepath.Base(filepath.Dir(path)); m.ID != "" && base != m.ID {
		return nil, fmt.Errorf("%s: folder %q does not match manifest id %q", path, base, m.ID)
	}
	if m.API != 1 {
		return nil, fmt.Errorf("%s: unsupported manifest api %d", path, m.API)
	}
	if !runtimes[m.Runtime] {
		return nil, fmt.Errorf("%s: unknown runtime %q", path, m.Runtime)
	}
	if m.Entry == "" {
		return nil, fmt.Errorf("%s: entry \"\"", path)
	}
	entry := filepath.Join(filepath.Dir(path), m.Entry)
	if _, err := os.Stat(entry); err != nil {
		return nil, fmt.Errorf("%s: entry %q: %w", path, m.Entry, err)
	}
	return &m, nil
}
