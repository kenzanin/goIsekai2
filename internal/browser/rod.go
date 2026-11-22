package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Rod implements Browser on a rod-connected CDP browser such as lightpanda.
type Rod struct {
	browser *rod.Browser
	mu      sync.Mutex
	pages   map[string]*rod.Page
	next    int
}

var _ Browser = (*Rod)(nil)

type rodArgs struct {
	Session  string `json:"session"`
	ID       string `json:"id"`
	URL      string `json:"url"`
	Script   string `json:"script"`
	Selector string `json:"selector"`
	State    string `json:"state"`
}

// ConnectRod opens one CDP connection and creates the default page session.
func ConnectRod(ctx context.Context, controlURL string, timeout time.Duration) (*Rod, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	b := rod.New().Context(ctx).ControlURL(controlURL)
	if err := b.Connect(); err != nil {
		return nil, fmt.Errorf("browser: connect %s: %w", controlURL, err)
	}
	page, err := b.Page(proto.TargetCreateTarget{})
	if err != nil {
		_ = b.Close()
		return nil, fmt.Errorf("browser: default page: %w", err)
	}
	page = page.Timeout(timeout)
	return &Rod{browser: b, pages: map[string]*rod.Page{"default": page}}, nil
}

// Close ends the CDP connection.
func (r *Rod) Close() error { return r.browser.Close() }

func (r *Rod) page(id string) *rod.Page {
	if id == "" {
		id = "default"
	}
	if page, ok := r.pages[id]; ok {
		return page
	}
	return r.pages["default"]
}

// Call dispatches one browser tool to rod.
func (r *Rod) Call(tool string, raw any) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var args rodArgs
	if b, err := json.Marshal(raw); err == nil {
		_ = json.Unmarshal(b, &args)
	}
	page := r.page(args.Session)
	switch tool {
	case "goto":
		if args.URL == "" {
			return "", fmt.Errorf("browser: goto: missing url")
		}
		if err := page.Navigate(args.URL); err != nil {
			return "",
				classifyRod(tool, err)
		}
		if err := page.WaitLoad(); err != nil {
			return "", classifyRod(tool, err)
		}
		return page.MustInfo().URL, nil
	case "evaluate":
		res, err := page.Eval(evalScript(args.Script))
		if err != nil {
			return "", classifyRod(tool, err)
		}
		return res.Value.String(), nil
	case "waitForSelector":
		if _, err := page.Element(args.Selector); err != nil {
			return "", classifyRod(tool, err)
		}
		return "ok", nil
	case "waitForState":
		switch args.State {
		case "stable":
			if err := page.WaitStable(300 * time.Millisecond); err != nil {
				return "", classifyRod(tool, err)
			}
		default:
			if err := page.WaitLoad(); err != nil {
				return "", classifyRod(tool, err)
			}
		}
		return "ok", nil
	case "waitForScript":
		if err := page.Wait(rod.Eval(args.Script)); err != nil {
			return "", classifyRod(tool, err)
		}
		return "ok", nil
	case "getCookies":
		cookies, err := page.Cookies(nil)
		if err != nil {
			return "", classifyRod(tool, err)
		}
		b, err := json.Marshal(cookies)
		return string(b), err
	case "session_new":
		p, err := r.browser.Page(proto.TargetCreateTarget{})
		if err != nil {
			return "", classifyRod(tool, err)
		}
		id := fmt.Sprintf("s%d", r.next)
		r.next++
		r.pages[id] = p.Timeout(30 * time.Second)
		return id, nil
	case "session_list":
		ids := make([]string, 0, len(r.pages))
		for id := range r.pages {
			ids = append(ids, id)
		}
		b, _ := json.Marshal(ids)
		return string(b), nil
	case "session_close":
		if args.ID == "default" || args.ID == "" {
			return "", fmt.Errorf("browser: cannot close default session")
		}
		p, ok := r.pages[args.ID]
		if !ok {
			return "", fmt.Errorf("%w: %s", ErrSessionExpired, args.ID)
		}
		delete(r.pages, args.ID)
		return "ok", p.Close()
	case "login":
		if args.URL != "" {
			if err := page.Navigate(args.URL); err != nil {
				return "", classifyRod(tool, err)
			}
			_ = page.WaitLoad()
		}
		if args.Script != "" {
			res, err := page.Eval(evalScript(args.Script))
			if err != nil {
				return "", classifyRod(tool, err)
			}
			return res.Value.String(), nil
		}
		return "ok", nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnknownTool, tool)
	}
}

func evalScript(script string) string {
	trimmed := strings.TrimSpace(script)
	if strings.HasPrefix(trimmed, "function") || strings.HasPrefix(trimmed, "()") {
		return trimmed
	}
	if strings.Contains(trimmed, "return ") {
		return "function(){" + trimmed + "}"
	}
	return "function(){ return (" + trimmed + ") }"
}

func classifyRod(tool string, err error) error {
	msg := err.Error()
	if strings.Contains(strings.ToLower(msg), "session") && strings.Contains(strings.ToLower(msg), "expired") {
		return fmt.Errorf("%w: %s", ErrSessionExpired, msg)
	}
	return fmt.Errorf("browser: %s: %w", tool, err)
}
