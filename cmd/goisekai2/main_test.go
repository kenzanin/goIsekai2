package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestHelpWithoutConfig(t *testing.T) {
	code, out, _ := runCLI(t, "--help")
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	for _, want := range []string{"--config", "--dump-config", "log-level"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q", want)
		}
	}
}

func TestMissingConfigExits2WithoutCreatingData(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(wd) })

	code, _, stderr := runCLI(t)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, "--config") {
		t.Fatalf("stderr = %q, want a --config hint", stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, "data")); !os.IsNotExist(err) {
		t.Fatalf("data dir was created: %v", err)
	}
}

func TestDumpConfigStdoutAndRefuseOverwrite(t *testing.T) {
	code, out, _ := runCLI(t, "--dump-config")
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(out, "port = 8090") || !strings.Contains(out, "[server]") {
		t.Fatalf("dump = %q", out)
	}

	path := filepath.Join(t.TempDir(), "gen.ini")
	code, _, _ = runCLI(t, "--dump-config", path)
	if code != 0 {
		t.Fatalf("file dump exit = %d, want 0", code)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("sentinel\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runCLI(t, "--dump-config", path)
	if code != 1 || !strings.Contains(stderr, "refusing to overwrite") {
		t.Fatalf("exit = %d stderr = %q, want refusal", code, stderr)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != "sentinel\n" {
		t.Fatalf("file changed on refusal: %q", after)
	}
	if len(first) == 0 {
		t.Fatal("first dump empty")
	}
}

func TestBootResolvesLayersAndFailsStrict(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.ini")
	if err := os.WriteFile(path, []byte("[server]\nport = 8081\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOISEKAI_PORT", "9111")
	code, out, _ := runCLI(t, "--config", path)
	if code != 0 || !strings.Contains(out, ":9111") {
		t.Fatalf("exit = %d out = %q, want env port", code, out)
	}

	t.Setenv("GOISEKAI_PORT", "http")
	code, _, stderr := runCLI(t, "--config", path)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr, "server.port") || !strings.Contains(stderr, "env") {
		t.Fatalf("stderr = %q, want key and layer", stderr)
	}
}

func TestFlagBeatsEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.ini")
	if err := os.WriteFile(path, []byte("[server]\nport = 8081\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOISEKAI_PORT", "9111")
	code, out, _ := runCLI(t, "--config", path, "--port", "7777")
	if code != 0 || !strings.Contains(out, ":7777") {
		t.Fatalf("exit = %d out = %q, want flag port", code, out)
	}
}
