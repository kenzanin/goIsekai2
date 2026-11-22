package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client posts tools/call requests to one browser server endpoint.
type Client struct {
	URL     string
	Timeout time.Duration

	mu      sync.Mutex
	session string
}

var _ Browser = (*Client)(nil)

type callParams struct {
	Name      string `json:"name"`
	Arguments any    `json:"arguments"`
}

type callRequest struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      int        `json:"id"`
	Method  string     `json:"method"`
	Params  callParams `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type callResult struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

type callResponse struct {
	Result callResult `json:"result"`
	Error  *rpcError  `json:"error"`
}

// Close forgets the cached session; no process is owned by the client.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session = ""
	return nil
}

// Call invokes a tool, recovering once from a session-expired response.
func (c *Client) Call(tool string, args any) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	text, err := c.callOnce(tool, args)
	if !errors.Is(err, ErrSessionExpired) {
		return text, err
	}
	c.session = ""
	text, err = c.callOnce(tool, args)
	if errors.Is(err, ErrSessionExpired) {
		return "", fmt.Errorf("%w: repeated", ErrSessionExpired)
	}
	return text, err
}

func (c *Client) transport() http.RoundTripper {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	host := ""
	if u, err := url.Parse(c.URL); err == nil {
		host = u.Hostname()
	}
	if host == "127.0.0.1" || host == "localhost" || host == "::1" {
		tr.Proxy = nil
	}
	return tr
}

func (c *Client) callOnce(tool string, args any) (string, error) {
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	body, err := json.Marshal(callRequest{
		JSONRPC: "2.0", ID: 1, Method: "tools/call",
		Params: callParams{Name: tool, Arguments: args},
	})
	if err != nil {
		return "", fmt.Errorf("encode tools/call %s: %w", tool, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build tools/call %s: %w", tool, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.session != "" {
		req.Header.Set("Mcp-Session-Id", c.session)
	}
	resp, err := (&http.Client{Timeout: timeout, Transport: c.transport()}).Do(req)
	if err != nil {
		return "", fmt.Errorf("post tools/call %s: %w", tool, err)
	}
	defer resp.Body.Close()
	if id := resp.Header.Get("Mcp-Session-Id"); id != "" {
		c.session = id
	}
	if resp.StatusCode != http.StatusOK {
		return "", classify(tool, fmt.Sprintf("http %d", resp.StatusCode))
	}
	var out callResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode tools/call %s: %w", tool, err)
	}
	if out.Error != nil {
		return "", classify(tool, out.Error.Message)
	}
	if len(out.Result.Content) == 0 {
		return "", fmt.Errorf("tools/call %s: empty content", tool)
	}
	return out.Result.Content[0].Text, nil
}

func classify(tool, message string) error {
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "unknown tool"):
		return fmt.Errorf("%w: %s: %s", ErrUnknownTool, tool, message)
	case strings.Contains(lower, "session") && strings.Contains(lower, "expired"):
		return fmt.Errorf("%w: %s", ErrSessionExpired, message)
	default:
		return fmt.Errorf("tools/call %s: %s", tool, message)
	}
}
