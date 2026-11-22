package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Read returns the contents of the file at name resolved inside the plugin
// directory dir. name must be relative and, once cleaned, must stay within
// dir: absolute paths and ".."-escaping paths are rejected.
func Read(dir, name string) (string, error) {
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("plugin file %q: path must be relative", name)
	}
	rel := filepath.Clean(name)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("plugin file %q: path escapes the plugin directory", name)
	}
	b, err := os.ReadFile(filepath.Join(dir, rel))
	return string(b), err
}

// ReadLib reads a file from the plugin's confined lib directory.
func ReadLib(dir, name string) (string, error) {
	return Read(dir, filepath.Join("lib", name))
}
