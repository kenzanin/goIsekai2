package browser

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// Supervisor owns at most one browser child process. A foreign instance is
// never stopped.
type Supervisor struct {
	URL    string
	Binary string
	Port   int

	mu      sync.Mutex
	cmd     *exec.Cmd
	owned   bool
	respawn bool
	dead    bool
}

// Probe reports whether a backend already answers on URL.
func (s *Supervisor) Probe(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return false
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.Proxy = nil
	resp, err := (&http.Client{Timeout: 2 * time.Second, Transport: tr}).Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < http.StatusInternalServerError
}

// Start reuses a healthy instance or spawns an owned child. Owned reports
// whether the child was spawned by this supervisor.
func (s *Supervisor) Start(ctx context.Context) (owned bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dead {
		return false, ErrDead
	}
	if s.Probe(ctx) {
		return false, nil
	}
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(s.Port))
	cmd := exec.Command(s.Binary, "serve", "--host", "127.0.0.1", "--port", strconv.Itoa(s.Port))
	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("browser: spawn %s: %w", s.Binary, err)
	}
	s.cmd, s.owned = cmd, true
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if s.Probe(ctx) {
			return true, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = cmd.Process.Kill()
	s.cmd, s.owned = nil, false
	return false, fmt.Errorf("browser: %s did not become healthy", addr)
}

// ControlURL returns the CDP websocket URL for the configured port.
func (s *Supervisor) ControlURL() string {
	return fmt.Sprintf("ws://127.0.0.1:%d/", s.Port)
}

// Respawn restarts an owned child once. A second failure marks the backend
// dead for the rest of the process.
func (s *Supervisor) Respawn(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dead {
		return ErrDead
	}
	if s.respawn || !s.owned {
		s.dead = true
		return ErrDead
	}
	s.respawn = true
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_, _ = s.cmd.Process.Wait()
	}
	s.cmd = nil
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(s.Port))
	cmd := exec.Command(s.Binary, "serve", "--host", "127.0.0.1", "--port", strconv.Itoa(s.Port))
	if err := cmd.Start(); err != nil {
		s.dead = true
		return fmt.Errorf("browser: respawn %s: %w", s.Binary, err)
	}
	s.cmd = cmd
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if s.Probe(ctx) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = cmd.Process.Kill()
	s.dead = true
	return fmt.Errorf("browser: respawn at %s did not become healthy", addr)
}

// Stop terminates only a child this supervisor owns.
func (s *Supervisor) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.owned || s.cmd == nil || s.cmd.Process == nil {
		return nil
	}
	_ = s.cmd.Process.Signal(syscall.SIGTERM)
	done := make(chan struct{})
	go func() { _, _ = s.cmd.Process.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = s.cmd.Process.Kill()
		<-done
	case <-ctx.Done():
		_ = s.cmd.Process.Kill()
		<-done
	}
	s.owned = false
	s.cmd = nil
	return nil
}
