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
	seen := make(map[string]bool)
	seenPaths := make(map[string]string)
	for _, d := range dirEntries {
		if !d.IsDir() {
			continue
		}
		folder := d.Name()
		m, err := LoadManifest(filepath.Join(dir, folder, "main.json"))
		if err != nil {
			slog.Warn("plugin: skipping broken manifest", "id", folder, "err", err)
			continue
		}
		id := folder
		if m.ID != "" {
			id = m.ID
		}
		if seen[id] {
			slog.Warn("plugin: duplicate id skipped", "id", id, "first", seenPaths[id], "second", filepath.Join(dir, folder))
			continue
		}
		seen[id] = true
		seenPaths[id] = filepath.Join(dir, folder)
		found = append(found, Entry{ID: folder, Manifest: m})
	}
	return found, nil
}
