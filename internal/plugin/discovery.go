package plugin

import (
	"log/slog"
	"os"
	"path/filepath"
)

// Entry is one discovered plugin: the folder name id and its manifest.
type Entry struct {
	ID       string
	Manifest *Manifest
}

// Discover scans dir's immediate subfolders in name order and loads each
// main.json; the folder name is the plugin id. Folders with a missing or
// broken manifest are skipped with a warning. A directory that cannot be
// read is an error.
func Discover(dir string) ([]Entry, error) {
	dirEntries, err := os.ReadDir(dir) // name-sorted
	if err != nil {
		return nil, err
	}
	var found []Entry
	for _, d := range dirEntries {
		if !d.IsDir() {
			continue
		}
		id := d.Name()
		m, err := LoadManifest(filepath.Join(dir, id, "main.json"))
		if err != nil {
			slog.Warn("plugin: skipping broken manifest", "id", id, "err", err)
			continue
		}
		found = append(found, Entry{ID: id, Manifest: m})
	}
	return found, nil
}
