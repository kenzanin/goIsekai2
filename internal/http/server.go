// Package httpserver hosts the daemon's HTTP routes and server-rendered pages.
package httpserver

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"

	"goisekai2/internal/config"
	"goisekai2/internal/events"
	"goisekai2/internal/http/components"
	"goisekai2/internal/library"
	"goisekai2/internal/plugin"
	"goisekai2/internal/store"
)

//go:embed assets/*
var assetFiles embed.FS

var assetRoot, _ = fs.Sub(assetFiles, "assets")

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
	httpServer *http.Server
	bus        *events.Bus
	host       *plugin.Host
	lib        *library.Service
	store      *store.Store
	generation atomic.Uint64
}

// New builds a server for reg and the ini file at iniPath. A nil bus disables
// domain-event publishing (used by tests that only exercise rendering).
func New(reg *config.Registry, iniPath string, bus *events.Bus, host *plugin.Host, lib *library.Service, st *store.Store) *Server {
	s := &Server{
		reg:     reg,
		iniPath: iniPath,
		bus:     bus,
		host:    host,
		lib:     lib,
		store:   st,
	}
	s.httpServer = &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

// Pages returns the implemented pages in navigation order.
func (s *Server) Pages() []Page {
	out := append([]Page(nil), pages...)
	out = append(out, Page{Path: "/library", Title: "Library"})
	out = append(out, Page{Path: "/search", Title: "Search"})
	if s.host != nil {
		out = append(out, Page{Path: "/plugins", Title: "Plugins"})
	}
	return out
}

// Handler returns the chi router; handlers stay plain http.HandlerFunc.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	for _, p := range pages {
		r.Get(p.Path, s.renderSettings)
	}
	r.Get("/config", s.getConfig)
	r.Post("/config", s.postConfig)
	r.Get("/", s.renderLibrary)
	r.Get("/library", s.renderLibrary)
	r.Get("/search", s.renderSearch)
	r.Get("/api/library", s.listLibrary)
	r.Post("/api/library", s.saveLibrary)
	r.Post("/api/library/remove", s.removeLibrary)
	if s.host != nil {
		r.Get("/plugins", s.renderPlugins)
		r.Post("/plugins/{id}/toggle", s.togglePlugin)
		r.Post("/plugins/{id}/reload", s.reloadPlugin)
		r.Get("/plugins/{id}/logo", s.pluginLogo)
	}
	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.FS(assetRoot))))
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

func parseStringList(raw string) []string {
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		return out
	}
	return strings.Fields(raw)
}

func (s *Server) libraryRows(ctx context.Context) []components.LibraryRow {
	if s.lib == nil {
		return nil
	}
	rows, err := s.lib.List(ctx)
	if err != nil {
		return nil
	}
	out := make([]components.LibraryRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, components.LibraryRow{
			Plugin: row.Plugin, Slug: row.Slug, Name: row.Name, Cover: row.Cover,
			Status: row.Status, StatusFlags: row.StatusFlags, DupGroup: row.DupGroup,
			Authors: parseStringList(row.Authors), Artists: parseStringList(row.Artists),
			Tags: parseStringList(row.Tags),
		})
	}
	return out
}

func (s *Server) renderLibrary(w http.ResponseWriter, r *http.Request) {
	active := components.Page{Path: r.URL.Path, Title: "Library"}
	if active.Path != "/library" {
		active.Path = "/"
	}
	pages := make([]components.Page, 0, len(s.Pages()))
	for _, p := range s.Pages() {
		pages = append(pages, components.Page{Path: p.Path, Title: p.Title})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.Library(active, pages, s.libraryRows(r.Context())).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderSearch(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	onlyPlugin := r.FormValue("plugin")
	plugins := make([]components.SearchPlugin, 0)
	if s.host != nil {
		for _, row := range s.host.Rows(r.Context()) {
			if row.Enabled {
				plugins = append(plugins, components.SearchPlugin{ID: row.ID, Name: row.Name})
			}
		}
	}
	hits := make([]components.SearchHit, 0)
	if query != "" && s.store != nil {
		found, err := s.store.SearchTitles(r.Context(), query)
		if err == nil {
			for _, hit := range found {
				if onlyPlugin == "" || hit.Plugin == onlyPlugin {
					hits = append(hits, components.SearchHit{Plugin: hit.Plugin, Slug: hit.Slug, Name: hit.Name})
				}
			}
		}
	}
	active := components.Page{Path: "/search", Title: "Search"}
	pages := make([]components.Page, 0, len(s.Pages()))
	for _, p := range s.Pages() {
		pages = append(pages, components.Page{Path: p.Path, Title: p.Title})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.Search(active, pages, plugins, hits, query).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) listLibrary(w http.ResponseWriter, r *http.Request) {
	if s.lib == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	rows, err := s.lib.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	category := r.FormValue("category")
	dup := r.FormValue("dup")
	plugin := r.FormValue("plugin")
	filtered := make([]store.SavedManga, 0, len(rows))
	bit := map[string]int{
		"ongoing": store.StatusOngoing, "hiatus": store.StatusHiatus,
		"canceled": store.StatusCanceled, "completed": store.StatusCompleted,
		"new": store.StatusNew, "done": store.StatusDone,
	}[category]
	for _, row := range rows {
		if bit != 0 && row.StatusFlags&bit == 0 {
			continue
		}
		if dup == "1" && row.DupGroup == 0 {
			continue
		}
		if plugin != "" && row.Plugin != plugin {
			continue
		}
		filtered = append(filtered, row)
	}
	writeJSON(w, http.StatusOK, filtered)
}

func (s *Server) saveLibrary(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Plugin string `json:"plugin"`
		Slug   string `json:"slug"`
		Name   string `json:"name"`
		Cover  string `json:"cover"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if s.lib == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "library unavailable"})
		return
	}
	if err := s.lib.Save(r.Context(), body.Plugin, body.Slug, body.Name, body.Cover); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": true})
}

func (s *Server) removeLibrary(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Plugin string `json:"plugin"`
		Slug   string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if s.lib == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "library unavailable"})
		return
	}
	if err := s.lib.Remove(r.Context(), body.Plugin, body.Slug); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": false})
}

func (s *Server) renderPlugins(w http.ResponseWriter, r *http.Request) {
	active := components.Page{Path: "/plugins", Title: "Plugins"}
	pages := make([]components.Page, 0, len(s.Pages()))
	for _, p := range s.Pages() {
		pages = append(pages, components.Page{Path: p.Path, Title: p.Title})
	}
	rows := make([]components.PluginRow, 0)
	for _, row := range s.host.Rows(r.Context()) {
		rows = append(rows, pluginRowFromHost(row))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.Plugins(active, pages, rows, s.host.Dir()).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func pluginRowFromHost(row plugin.HostRow) components.PluginRow {
	return components.PluginRow{
		ID: row.ID, Name: row.Name, Version: row.Version, Runtime: row.Runtime,
		Capabilities: row.Capabilities, TitleCount: row.TitleCount, Status: row.Status,
		Error: row.Error, Enabled: row.Enabled, HasLogo: row.HasLogo,
	}
}

func (s *Server) pluginRow(id string) (components.PluginRow, bool) {
	for _, row := range s.host.Rows(context.Background()) {
		if row.ID == id {
			return pluginRowFromHost(row), true
		}
	}
	return components.PluginRow{}, false
}

func (s *Server) renderPluginStatus(w http.ResponseWriter, r *http.Request, id string, actionErr error) {
	row, ok := s.pluginRow(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if actionErr != nil {
		row.Status, row.Error = "error", actionErr.Error()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.PluginStatusFragment(row).Render(r.Context(), w)
}

func (s *Server) togglePlugin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	row, ok := s.pluginRow(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	err := s.host.SetEnabled(r.Context(), id, !row.Enabled)
	s.renderPluginStatus(w, r, id, err)
}

func (s *Server) reloadPlugin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := s.host.Reload(r.Context())
	s.renderPluginStatus(w, r, id, err)
}

func (s *Server) pluginLogo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	for _, row := range s.host.Rows(r.Context()) {
		if row.ID == id && row.HasLogo && row.Logo != "" {
			http.ServeFile(w, r, filepath.Join(s.host.Dir(), id, row.Logo))
			return
		}
	}
	http.NotFound(w, r)
}

func (s *Server) renderSettings(w http.ResponseWriter, r *http.Request) {
	projections, err := config.Project(s.reg, s.iniPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bySection := make(map[string][]components.Row)
	var sections []string
	for _, p := range projections {
		if _, seen := bySection[p.Section]; !seen {
			sections = append(sections, p.Section)
		}
		r := components.Row{
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
	groups := make([]components.Group, 0, len(sections))
	for _, name := range sections {
		groups = append(groups, components.Group{Section: name, Rows: bySection[name]})
	}
	pages := make([]components.Page, 0, len(pages))
	for _, p := range s.Pages() {
		pages = append(pages, components.Page{Path: p.Path, Title: p.Title})
	}
	active := components.Page{Path: "/settings", Title: "Settings"}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.Settings(active, pages, groups).Render(r.Context(), w); err != nil {
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
	htmx := r.Header.Get("HX-Request") == "true"
	respond := func(status int, ok bool, written int, errMsg string) {
		if htmx {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_ = components.StatusFragment(ok, written, errMsg).Render(r.Context(), w)
			return
		}
		if ok {
			writeJSON(w, status, map[string]any{"ok": true, "written": written})
			return
		}
		writeJSON(w, status, map[string]any{"ok": false, "error": errMsg})
	}

	values := make(map[string]string)
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			respond(http.StatusBadRequest, false, 0, err.Error())
			return
		}
		for key := range r.PostForm {
			values[key] = r.PostForm.Get(key)
		}
	} else if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
		respond(http.StatusBadRequest, false, 0, err.Error())
		return
	}

	old, _ := config.Project(s.reg, s.iniPath)
	written, err := config.MergeSave(s.reg, s.iniPath, values)
	if err != nil {
		respond(http.StatusBadRequest, false, 0, err.Error())
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

	if ini, err := config.LoadIni(s.iniPath); err == nil {
		if snap, err := config.Resolve(s.reg, ini, os.Getenv, nil); err == nil {
			config.ApplySnapshot(snap)
		}
	}
	respond(http.StatusOK, true, written, "")
}

// writeJSON emits a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
