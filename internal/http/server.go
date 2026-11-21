// Package httpserver hosts the daemon's HTTP routes and server-rendered pages.
package httpserver

import (
	"context"
	"embed"
	"encoding/json"
	"html/template"
	"net"
	"net/http"
	"os"
	"sort"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"

	"goisekai2/internal/config"
	"goisekai2/internal/events"
)

//go:embed templates/*.html
var templateFiles embed.FS

// Page is one implemented page; routing and navigation share this list.
type Page struct {
	Path  string
	Title string
}

var pages = []Page{{Path: "/settings", Title: "Settings"}}

// Server owns the HTTP listener, router and template set.
type Server struct {
	reg        *config.Registry
	iniPath    string
	templates  *template.Template
	httpServer *http.Server
	bus        *events.Bus
	generation atomic.Uint64
}

// New builds a server for reg and the ini file at iniPath. A nil bus disables
// domain-event publishing (used by tests that only exercise rendering).
func New(reg *config.Registry, iniPath string, bus *events.Bus) *Server {
	t := template.Must(template.ParseFS(templateFiles, "templates/*.html"))
	s := &Server{
		reg:       reg,
		iniPath:   iniPath,
		templates: t,
		bus:       bus,
	}
	s.httpServer = &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

// Pages returns the implemented pages in navigation order.
func (s *Server) Pages() []Page {
	return append([]Page(nil), pages...)
}

// Handler returns the chi router; handlers stay plain http.HandlerFunc.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	for _, p := range pages {
		page := p
		r.Get(page.Path, func(w http.ResponseWriter, r *http.Request) {
			s.renderSettings(w, page)
		})
	}
	r.Get("/config", s.getConfig)
	r.Post("/config", s.postConfig)
	return r
}

// Serve accepts connections on l until Shutdown is called.
func (s *Server) Serve(l net.Listener) error {
	return s.httpServer.Serve(l)
}

// Shutdown stops accepting new requests and waits for in-flight handlers.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

type row struct {
	FullKey  string
	Key      string
	Doc      string
	Value    string
	Source   string
	Hint     string
	Disabled bool
	Restart  bool
	Invalid  bool
}

type group struct {
	Section string
	Rows    []row
}

type pageData struct {
	Active Page
	Pages  []Page
	Groups []group
	Error  string
}

func (s *Server) renderSettings(w http.ResponseWriter, active Page) {
	projections, err := config.Project(s.reg, s.iniPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bySection := make(map[string][]row)
	var sections []string
	for _, p := range projections {
		if _, seen := bySection[p.Section]; !seen {
			sections = append(sections, p.Section)
		}
		r := row{
			FullKey:  p.FullKey,
			Key:      p.Key,
			Doc:      p.Doc,
			Value:    p.Value,
			Source:   p.Source,
			Invalid:  p.Invalid,
			Restart:  !config.HasLiveApplier(p.FullKey),
			Disabled: p.Source == "env" || p.Source == "flag",
		}
		if r.Disabled {
			r.Hint = "set by " + p.Source
		}
		bySection[p.Section] = append(bySection[p.Section], r)
	}
	sort.Strings(sections)
	groups := make([]group, 0, len(sections))
	for _, name := range sections {
		groups = append(groups, group{Section: name, Rows: bySection[name]})
	}
	data := pageData{Active: active, Pages: s.Pages(), Groups: groups}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) getConfig(w http.ResponseWriter, _ *http.Request) {
	projections, err := config.Project(s.reg, s.iniPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, projections)
}

func (s *Server) postConfig(w http.ResponseWriter, r *http.Request) {
	var values map[string]string
	if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	old, _ := config.Project(s.reg, s.iniPath)
	written, err := config.MergeSave(s.reg, s.iniPath, values)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	gen := s.generation.Add(1)
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if s.bus != nil {
		for _, k := range keys {
			oldValue := ""
			for _, p := range old {
				if p.FullKey == k {
					oldValue = p.Value
					break
				}
			}
			s.bus.Publish(events.ConfigChanged, events.ConfigChangedPayload{
				Key: k, Old: oldValue, New: values[k], Generation: gen,
			})
		}
	}

	ini, err := config.LoadIni(s.iniPath)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "written": written, "error": err.Error()})
		return
	}
	snap, err := config.Resolve(s.reg, ini, os.Getenv, nil)
	restart := []string{}
	if err == nil {
		restart = config.ApplySnapshot(snap)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "written": written, "restartRequired": restart,
	})
}

// writeJSON emits a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
