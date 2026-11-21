package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectAndMergeSave(t *testing.T) {
	tmpDir := t.TempDir()
	iniPath := filepath.Join(tmpDir, "config.ini")

	r := Knobs()

	// 1. MergeSave on absent file creates from dump template and sets values
	count, err := MergeSave(r, iniPath, map[string]string{
		"server.port": "9090",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify file exists and has content
	content, err := os.ReadFile(iniPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "port")
	assert.Contains(t, string(content), "9090")

	// 2. Malformed knob yields exactly one Invalid row with strict boot failing on same input
	err = os.WriteFile(iniPath, []byte("[server]\nport = invalid\n"), 0600)
	require.NoError(t, err)

	projs, err := Project(r, iniPath)
	require.NoError(t, err)

	var invalidCount int
	var invalidFullKey string
	for _, p := range projs {
		if p.Invalid {
			invalidCount++
			invalidFullKey = p.FullKey
		}
	}
	assert.Equal(t, 1, invalidCount)
	assert.Equal(t, "server.port", invalidFullKey)

	// Strict boot check (Resolve fails on invalid ini value)
	iniFile, err := LoadIni(iniPath)
	require.NoError(t, err)
	_, resolveErr := Resolve(r, iniFile, os.Getenv, nil)
	assert.Error(t, resolveErr)

	// 3. Merge-save of one key preserves comments and every other value
	err = os.WriteFile(iniPath, []byte("; TCP listen port\n[server]\nport = 8090\n\n; Outbound HTTP(S) proxy URL\nproxy = http://proxy\n"), 0600)
	require.NoError(t, err)

	count, err = MergeSave(r, iniPath, map[string]string{
		"server.port": "8091",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	updatedContent, err := os.ReadFile(iniPath)
	require.NoError(t, err)
	assert.Contains(t, string(updatedContent), "; TCP listen port")
	assert.Contains(t, string(updatedContent), "port")
	assert.Contains(t, string(updatedContent), "8091")
	assert.Contains(t, string(updatedContent), "; Outbound HTTP(S) proxy URL")
	assert.Contains(t, string(updatedContent), "proxy = http://proxy")

	// 4. Rejected key leaves file byte-identical
	beforeHash, err := os.ReadFile(iniPath)
	require.NoError(t, err)

	_, err = MergeSave(r, iniPath, map[string]string{
		"server.port": "not-an-int",
	})
	assert.Error(t, err)

	afterHash, err := os.ReadFile(iniPath)
	require.NoError(t, err)
	assert.Equal(t, beforeHash, afterHash)
}
