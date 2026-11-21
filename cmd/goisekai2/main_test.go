package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runCLI(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // tests never want the daemon to wait for a real signal
	var out, errb bytes.Buffer
	code = run(ctx, args, &out, &errb)
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

func TestDumpMatchesGolden(t *testing.T) {
	code, out, _ := runCLI(t, "--dump-config")
	if code != 0 {
		t.Fatalf("dump exit = %d", code)
	}
	golden, err := os.ReadFile("testdata/default.ini")
	if err != nil {
		t.Fatal(err)
	}
	if out != string(golden) {
		t.Fatalf("dump differs from golden:\n--- got ---\n%s\n--- want ---\n%s", out, golden)
	}
}

func TestDaemonEndToEnd(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	dir := t.TempDir()
	iniPath := filepath.Join(dir, "cfg.ini")
	if err := os.WriteFile(iniPath, []byte("[server]\n; keep this comment\nport = 8090\n\n[log]\nlevel = info\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan int, 1)
	go func() {
		var out, errb bytes.Buffer
		done <- run(ctx, []string{"--config", iniPath, "--port", fmt.Sprint(port)}, &out, &errb)
	}()

	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	var resp *http.Response
	for i := 0; i < 40; i++ {
		resp, err = client.Get(base + "/settings")
		if err == nil {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("daemon never listened: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), `href="/settings" aria-current="page"`) {
		t.Fatalf("settings page missing active nav: %s", body)
	}

	resp, err = client.Get(base + "/config")
	if err != nil {
		t.Fatal(err)
	}
	var projections []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&projections); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if len(projections) == 0 || projections[0]["fullKey"] != "server.port" {
		t.Fatalf("projection = %v", projections)
	}

	req, _ := http.NewRequest(http.MethodPost, base+"/config", strings.NewReader(`{"log.level":"debug"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if result["ok"] != true || result["written"] != float64(1) {
		t.Fatalf("save result = %v", result)
	}

	onDisk, err := os.ReadFile(iniPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(onDisk), "; keep this comment") || !strings.Contains(string(onDisk), "level = debug") {
		t.Fatalf("file not merged correctly: %q", onDisk)
	}

	resp, err = client.Get(base + "/settings")
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), `value="debug"`) {
		t.Fatalf("settings page did not re-read saved value: %s", body)
	}
	if !strings.Contains(string(body), "restart required") {
		t.Fatalf("settings page missing restart badge: %s", body)
	}

	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("daemon exit = %d", code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("daemon did not shut down")
	}
}
