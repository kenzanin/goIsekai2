// Package browser calls the browser-control MCP server over HTTP JSON-RPC.
package browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client posts tools/call requests to one browser server endpoint.
type Client struct {
	// URL is the endpoint accepting JSON-RPC POST requests.
	URL string
	// Timeout bounds each call; zero means no timeout.
	Timeout time.Duration
}

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
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Text string `json:"text"`
}

type callResponse struct {
	Result callResult `json:"result"`
	Error  *rpcError  `json:"error"`
}

// Call invokes the named tool with args and returns its first text content block.
func (c *Client) Call(tool string, args any) (string, error) {
	body, err := json.Marshal(callRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  callParams{Name: tool, Arguments: args},
	})
	if err != nil {
		return "", fmt.Errorf("encode tools/call %s: %w", tool, err)
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.Proxy = nil // loopback: never route through a proxy from the environment
	hc := &http.Client{Timeout: c.Timeout, Transport: tr}
	req, err := http.NewRequest(http.MethodPost, c.URL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build tools/call %s: %w", tool, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("post tools/call %s: %w", tool, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tools/call %s: http %d", tool, resp.StatusCode)
	}
	// ponytail: JSON responses only; add text/event-stream parsing if the endpoint becomes a streamable MCP server.
	var out callResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode tools/call %s: %w", tool, err)
	}
	if out.Error != nil {
		return "", fmt.Errorf("tools/call %s: rpc %d: %s", tool, out.Error.Code, out.Error.Message)
	}
	if len(out.Result.Content) == 0 {
		return "", fmt.Errorf("tools/call %s: empty content", tool)
	}
	return out.Result.Content[0].Text, nil
}
