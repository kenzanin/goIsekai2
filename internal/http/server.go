// Package httpserver hosts the daemon's HTTP routes and server-rendered pages.
package httpserver

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	coderwebsocket "github.com/coder/websocket"
	"github.com/go-chi/chi/v5"

	imagecache "goisekai2/internal/cache"
	"goisekai2/internal/config"
	"goisekai2/internal/detail"
	"goisekai2/internal/events"
	"goisekai2/internal/http/components"
	"goisekai2/internal/library"
	"goisekai2/internal/plugin"
	"goisekai2/internal/reader"
	"goisekai2/internal/store"
	"goisekai2/internal/stream"
)

//go:embed assets/*
var assetFiles embed.FS

var assetRoot, _ = fs.Sub(assetFiles, "assets")

// Page is one implemented page; routing and navigation share this list.
type Page struct {
	Path  string
	Title string
}

var pages = []Page{
	{Path: "/settings", Title: "Settings"},
	{Path: "/log", Title: "Log"},
	{Path: "/about", Title: "About"},
}

// Server owns the HTTP listener, router and template set.
type Server struct {
	reg        *config.Registry
	iniPath    string
	httpServer *http.Server
	cacheStore *imagecache.Store
	detailSvc  *detail.Service
	streamHub  *stream.Hub
	bus        *events.Bus
	host       *plugin.Host
	lib        *library.Service
	reader     *reader.Service
	store      *store.Store
	generation atomic.Uint64
}

// New builds a server for reg and the ini file at iniPath. A nil bus disables
// domain-event publishing (used by tests that only exercise rendering).
func New(reg *config.Registry, iniPath string, bus *events.Bus, host *plugin.Host, lib *library.Service, rd *reader.Service, st *store.Store) *Server {
	s := &Server{
		reg:     reg,
		iniPath: iniPath,
		bus:     bus,
		host:    host,
		lib:     lib,
		reader:  rd,
		store:   st,
	}
	s.httpServer = &http.Server{
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

// SetCache attaches the image cache store used by /cache/img.
func (s *Server) SetCache(c *imagecache.Store) { s.cacheStore = c }

// SetDetail attaches the detail resolver used by the detail routes.
func (s *Server) SetDetail(d *detail.Service) { s.detailSvc = d }

// SetStream attaches the WebSocket log hub.
func (s *Server) SetStream(h *stream.Hub) { s.streamHub = h }

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
	r.Get("/ws/log", s.wsLog)
	r.Get("/log", s.renderLog)
	r.Get("/about", s.renderAbout)
	r.Get("/config", s.getConfig)
	r.Post("/config", s.postConfig)
	r.Get("/view/manga/{plugin}/{slug}", s.renderDetail)
	r.Get("/api/detail/{plugin}/{slug}", s.getDetail)
	r.Post("/api/refresh/{plugin}/{slug}", s.refreshDetail)
	r.Post("/api/refresh/{plugin}/{slug}/force", s.refreshDetail)
	r.Post("/api/detail/{plugin}/{slug}/reset", s.resetDetail)
	r.Post("/api/detail/{plugin}/{slug}/apply", s.applyCandidate)
	r.Post("/api/enrichment/refresh/{plugin}/{slug}", s.refreshEnrichment)
	r.Get("/cache/img", s.getCacheImage)
	r.Get("/", s.renderLibrary)
	r.Get("/library", s.renderLibrary)
	r.Get("/search", s.renderSearch)
	r.Get("/api/search", s.apiSearch)
	r.Get("/api/library", s.listLibrary)
	r.Get("/read/{plugin}/{slug}", s.renderRead)
	r.Get("/api/pages/{plugin}/{slug}/{ch}", s.getPages)
	r.Post("/reader/position", s.postReaderPosition)
	r.Post("/chapter/flags", s.postChapterFlags)
	r.Get("/api/chapters/{plugin}/{slug}", s.listChapters)
	r.Get("/chapters/{plugin}/{slug}", s.renderChapters)
	r.Post("/api/library", s.saveLibrary)
	r.Post("/api/library/remove", s.removeLibrary)
	if s.host != nil {
		r.Get("/plugins", s.renderPlugins)
		r.Post("/plugins/{id}/toggle", s.togglePlugin)
		r.Post("/plugins/{id}/reload", s.reloadPlugin)
		r.Get("/plugins/{id}/logo", s.pluginLogo)
		r.Get("/plugin/{id}/logo", s.pluginLogo)
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

func (s *Server) getDetail(w http.ResponseWriter, r *http.Request) {
	if s.detailSvc == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "detail unavailable"})
		return
	}
	data, err := s.detailSvc.Detail(r.Context(), chi.URLParam(r, "plugin"), chi.URLParam(r, "slug"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "unknown title"})
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) refreshDetail(w http.ResponseWriter, r *http.Request) {
	if s.detailSvc == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "detail unavailable"})
		return
	}
	if err := s.detailSvc.Refresh(r.Context(), chi.URLParam(r, "plugin"), chi.URLParam(r, "slug")); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) resetDetail(w http.ResponseWriter, r *http.Request) {
	if s.detailSvc == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "detail unavailable"})
		return
	}
	pluginID, slug := chi.URLParam(r, "plugin"), chi.URLParam(r, "slug")
	if err := s.detailSvc.ResetDetail(r.Context(), pluginID, slug); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if err := s.detailSvc.ResetCover(r.Context(), pluginID, slug); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) applyCandidate(w http.ResponseWriter, r *http.Request) {
	if s.detailSvc == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "detail unavailable"})
		return
	}
	id := r.FormValue("candidate")
	title := chi.URLParam(r, "slug")
	if row, err := s.store.GetSaved(r.Context(), chi.URLParam(r, "plugin"), title); err == nil && row.Name != "" {
		title = row.Name
	}
	for _, data := range s.detailSvc.Candidates(r.Context(), title, "detail") {
		if data["id"] == id {
			if err := s.detailSvc.ApplyCandidate(r.Context(), chi.URLParam(r, "plugin"), chi.URLParam(r, "slug"), data); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "unknown candidate"})
}

// refreshEnrichment is the explicit user action that fetches enrichment
// sources; GET routes never do this work.
func (s *Server) refreshEnrichment(w http.ResponseWriter, r *http.Request) {
	if s.detailSvc == nil || s.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "detail unavailable"})
		return
	}
	title := chi.URLParam(r, "slug")
	if row, err := s.store.GetSaved(r.Context(), chi.URLParam(r, "plugin"), title); err == nil && row.Name != "" {
		title = row.Name
	}
	s.detailSvc.RefreshCandidates(r.Context(), title, "detail")
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// imageSrc routes a remote image through the host cache so browser hotlink
// blocks do not break covers or reader pages.
func imageSrc(pluginID, slug, raw string, cover bool) string {
	if raw == "" {
		return ""
	}
	coverValue := "0"
	if cover {
		coverValue = "1"
	}
	return "/cache/img?plugin=" + url.QueryEscape(pluginID) +
		"&slug=" + url.QueryEscape(slug) +
		"&cover=" + coverValue +
		"&url=" + url.QueryEscape(raw)
}

// proxiedPages rewrites remote page URLs through the host image cache, keeping
// the chapter placement needed to lay each image out in the on-disk cache tree.
func proxiedPages(pluginID, slug, chapter, chapterID string, urls []string) []string {
	out := make([]string, len(urls))
	for i, raw := range urls {
		if raw == "" {
			continue
		}
		out[i] = imageSrc(pluginID, slug, raw, false) +
			"&ch=" + url.QueryEscape(chapter) +
			"&ch_id=" + url.QueryEscape(chapterID) +
			"&page=" + strconv.Itoa(i+1)
	}
	return out
}

func (s *Server) renderDetail(w http.ResponseWriter, r *http.Request) {
	if s.detailSvc == nil || s.store == nil {
		http.NotFound(w, r)
		return
	}
	pluginID, slug := chi.URLParam(r, "plugin"), chi.URLParam(r, "slug")
	data, err := s.detailSvc.Detail(r.Context(), pluginID, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	str := func(key string) string {
		if v, ok := data[key].(string); ok {
			return v
		}
		return ""
	}
	row := components.DetailRow{
		Plugin: pluginID, Slug: slug, Name: str("name"), Cover: str("cover"),
		Status: str("status"), Synopsis: str("synopsis"),
		ChosenCover: str("chosen_cover"), Cached: data["cached"] == true,
		Saved: data["saved"] == true,
	}
	row.Cover = imageSrc(pluginID, slug, row.Cover, true)
	row.ChosenCover = imageSrc(pluginID, slug, row.ChosenCover, true)
	if authors, ok := data["authors"].([]string); ok {
		row.Authors = authors
	} else if v, ok := data["authors"].(string); ok && v != "" {
		row.Authors = strings.Split(v, ", ")
	}
	if tags, ok := data["tags"].([]string); ok {
		row.Tags = tags
	} else if v, ok := data["tags"].(string); ok && v != "" {
		row.Tags = strings.Split(v, ", ")
	}
	var stored []store.Chapter
	if err := s.store.DB().WithContext(r.Context()).Where("plugin = ? AND slug = ?", pluginID, slug).
		Order("number, chapter_id").Find(&stored).Error; err != nil {
		stored = nil
	}
	totalPages := (len(stored) + 19) / 20
	if totalPages == 0 {
		totalPages = 1
	}
	detailPage, _ := strconv.Atoi(r.FormValue("page"))
	if detailPage < 1 {
		detailPage = 1
	}
	if detailPage > totalPages {
		detailPage = totalPages
	}
	start, end := (detailPage-1)*20, detailPage*20
	if end > len(stored) {
		end = len(stored)
	}
	chapters := make([]components.ChapterRow, 0, max(0, end-start))
	for _, ch := range stored[start:end] {
		chapters = append(chapters, components.ChapterRow{
			Plugin: pluginID, Slug: slug, ID: ch.ChapterID, Number: ch.Number,
			Title: ch.Title, Read: ch.Flags&store.ChapterRead != 0,
			Skip: ch.Flags&store.ChapterSkip != 0,
		})
	}
	candidates := make([]components.Candidate, 0)
	for _, data := range s.detailSvc.Candidates(r.Context(), row.Name, "detail") {
		field := func(key string) string {
			if v, ok := data[key].(string); ok {
				return v
			}
			return ""
		}
		candidates = append(candidates, components.Candidate{
			ID: field("id"), Name: field("name"),
			Cover: field("cover"), Synopsis: field("synopsis"),
		})
	}
	pages := make([]components.Page, 0, len(s.Pages()))
	for _, p := range s.Pages() {
		pages = append(pages, components.Page{Path: p.Path, Title: p.Title})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.Detail(components.Page{Path: "/view", Title: row.Name}, pages, row, chapters, candidates, detailPage, totalPages).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) getCacheImage(w http.ResponseWriter, r *http.Request) {
	if s.cacheStore == nil {
		http.NotFound(w, r)
		return
	}
	pluginID, slug, rawURL := r.FormValue("plugin"), r.FormValue("slug"), r.FormValue("url")
	if pluginID == "" || slug == "" || rawURL == "" {
		http.Error(w, "missing placement context", http.StatusBadRequest)
		return
	}
	root := s.cacheStore.Root
	isCover := r.FormValue("cover") == "1"
	var base string
	if isCover {
		base = strings.TrimSuffix(imagecache.CoverPath(root, pluginID, imagecache.Hash(rawURL), "bin"), ".bin")
	} else {
		ch := r.FormValue("ch")
		page, err := strconv.Atoi(r.FormValue("page"))
		if ch == "" || err != nil || page < 1 {
			http.Error(w, "missing placement context", http.StatusBadRequest)
			return
		}
		number, _ := strconv.ParseFloat(ch, 64)
		base = strings.TrimSuffix(imagecache.PagePath(root, pluginID, slug, number, r.FormValue("ch_id"), page, "bin"), ".bin")
	}
	matches, _ := filepath.Glob(base + ".*")
	if len(matches) == 0 || strings.HasSuffix(matches[0], ".json") {
		referer := ""
		if s.host != nil {
			referer = s.host.ImageReferer(pluginID)
		}
		body, err := s.cacheStore.GetOrFetchWithReferer(r.Context(), imagecache.Hash(rawURL), rawURL, referer)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		path, _, err := s.cacheStore.Put(base+".bin", body, isCover)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		matches = []string{path}
	}
	path := matches[0]
	body, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	s.cacheStore.Pool.Touch(path)
	w.Header().Set("Content-Type", contentTypeFor(path))
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	if ratio := imagecache.Ratio(path); ratio > 0 {
		w.Header().Set("X-Image-Ratio", strconv.FormatFloat(ratio, 'f', -1, 64))
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	_, _ = w.Write(body)
}

func contentTypeFor(path string) string {
	switch strings.TrimPrefix(filepath.Ext(path), ".") {
	case "webp":
		return "image/webp"
	case "avif":
		return "image/avif"
	case "jxl":
		return "image/jxl"
	default:
		return "application/octet-stream"
	}
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
	if query != "" {
		scope := onlyPlugin
		if scope == "" && len(plugins) > 0 {
			scope = plugins[0].ID
		}
		if scope != "" {
			hits = s.pluginHits(r.Context(), scope, query)
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

type searchItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Name     string `json:"name"`
	CoverURL string `json:"cover_url"`
	Cover    string `json:"cover"`
}

// pluginHits projects a plugin search envelope into UI hits, same data as the API.
func (s *Server) pluginHits(ctx context.Context, pluginID, query string) []components.SearchHit {
	params, _ := json.Marshal(map[string]any{"query": query, "page": 1})
	out, err := s.host.Call(ctx, pluginID, "search", params)
	if err != nil {
		slog.Debug("search projection failed", "plugin", pluginID, "error", err)
		return nil
	}
	var env struct {
		OK   bool `json:"ok"`
		Data struct {
			Items []searchItem `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &env); err != nil || !env.OK {
		return nil
	}
	hits := make([]components.SearchHit, 0, len(env.Data.Items))
	for _, item := range env.Data.Items {
		name := item.Title
		if name == "" {
			name = item.Name
		}
		hits = append(hits, components.SearchHit{Plugin: pluginID, Slug: item.ID, Name: name})
	}
	return hits
}

func (s *Server) apiSearch(w http.ResponseWriter, r *http.Request) {
	if s.host == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "plugin host unavailable"})
		return
	}
	pluginID := r.FormValue("plugin")
	if pluginID == "" {
		for _, row := range s.host.Rows(r.Context()) {
			if row.Enabled {
				pluginID = row.ID
				break
			}
		}
	}
	if pluginID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "no enabled plugin"})
		return
	}
	params, _ := json.Marshal(map[string]any{
		"query": r.FormValue("q"), "page": 1,
	})
	slog.Debug("plugin search api", "plugin", pluginID, "query", r.FormValue("q"))
	out, err := s.host.Call(r.Context(), pluginID, "search", params)
	slog.Debug("plugin search api result", "plugin", pluginID, "bytes", len(out), "error", err)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error(), "plugin": pluginID})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

func (s *Server) renderRead(w http.ResponseWriter, r *http.Request) {
	if s.reader == nil || s.store == nil {
		http.NotFound(w, r)
		return
	}
	pluginID, slug := chi.URLParam(r, "plugin"), chi.URLParam(r, "slug")
	number, _ := strconv.ParseFloat(r.FormValue("ch"), 64)
	result, err := s.reader.Pages(r.Context(), pluginID, slug, number, r.FormValue("ch_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	chapter := components.ReaderChapter{Number: number, ID: r.FormValue("ch_id")}
	if v, ok := result["chapter"].(float64); ok {
		chapter.Number = v
	}
	if urls, ok := result["pages"].([]string); ok {
		result["pages"] = proxiedPages(pluginID, slug, strconv.FormatFloat(number, 'f', -1, 64), r.FormValue("ch_id"), urls)
		for i, u := range result["pages"].([]string) {
			chapter.Pages = append(chapter.Pages, components.ReaderPage{URL: u, Index: i})
		}
	}
	f := func(key string) float64 {
		if v, ok := result[key].(float64); ok {
			return v
		}
		return 0
	}
	sv := func(key string) string {
		if v, ok := result[key].(string); ok {
			return v
		}
		return ""
	}
	chapter.PrevNumber, chapter.NextNumber = f("prev_ch"), f("next_ch")
	chapter.PrevID, chapter.NextID = sv("prev_ch_id"), sv("next_ch_id")

	mode, direction, page := "webtoon", "ltr", 0
	if pos, err := s.store.GetPosition(r.Context(), pluginID, slug); err == nil {
		mode, direction, page = pos.Mode, pos.Direction, pos.Page
	}
	if mode == "" {
		mode = "webtoon"
	}
	if direction == "" {
		direction = "ltr"
	}
	name := slug
	if s.lib != nil {
		if row, err := s.lib.Get(r.Context(), pluginID, slug); err == nil && row.Name != "" {
			name = row.Name
		}
	}
	allChapters := make([]components.ChapterOption, 0)
	if stored, err := s.reader.Chapters(r.Context(), pluginID, slug); err == nil {
		for _, row := range stored {
			allChapters = append(allChapters, components.ChapterOption{ID: row.ChapterID, Number: row.Number, Title: row.Title})
		}
	}
	active := components.Page{Path: "/read", Title: "Reader"}
	pages := make([]components.Page, 0, len(s.Pages()))
	for _, p := range s.Pages() {
		pages = append(pages, components.Page{Path: p.Path, Title: p.Title})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.Reader(active, pages, name, chapter, mode, direction, page, allChapters).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) getPages(w http.ResponseWriter, r *http.Request) {
	if s.reader == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "reader unavailable"})
		return
	}
	number, _ := strconv.ParseFloat(chi.URLParam(r, "ch"), 64)
	result, err := s.reader.Pages(r.Context(), chi.URLParam(r, "plugin"), chi.URLParam(r, "slug"), number, r.FormValue("ch_id"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "unknown chapter"})
		return
	}
	if urls, ok := result["pages"].([]string); ok {
		result["pages"] = proxiedPages(chi.URLParam(r, "plugin"), chi.URLParam(r, "slug"), chi.URLParam(r, "ch"), r.FormValue("ch_id"), urls)
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) postReaderPosition(w http.ResponseWriter, r *http.Request) {
	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if s.reader == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "reader unavailable"})
		return
	}
	plugin, _ := fields["plugin"].(string)
	slug, _ := fields["slug"].(string)
	delete(fields, "plugin")
	delete(fields, "slug")
	if err := s.reader.PutPosition(r.Context(), plugin, slug, fields); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	next, _ := fields["next_ch_id"].(string)
	prev, _ := fields["prev_ch_id"].(string)
	if chapterID, ok := fields["chapter_id"].(string); ok && chapterID != "" && next == "" && prev == "" {
		next, prev = s.reader.NeighbourIDs(r.Context(), plugin, slug, chapterID)
	}
	if next != "" || prev != "" {
		go s.reader.Prefetch(context.Background(), plugin, slug, next, prev)
	}
	if s.cacheStore != nil && s.store != nil {
		go s.prefetchWindow(context.Background(), plugin, slug)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) postChapterFlags(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Plugin     string   `json:"plugin"`
		Slug       string   `json:"slug"`
		ChapterID  string   `json:"chapter_id"`
		ChapterIDs []string `json:"chapter_ids"`
		Set        int      `json:"set"`
		Clear      int      `json:"clear"`
	}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		body.Plugin, body.Slug = r.FormValue("plugin"), r.FormValue("slug")
		body.ChapterIDs = r.Form["ch_id"]
		if len(body.ChapterIDs) == 0 {
			body.ChapterIDs = r.Form["chapter_id"]
		}
		if len(body.ChapterIDs) > 0 {
			body.ChapterID = body.ChapterIDs[0]
		}
		body.Set, _ = strconv.Atoi(r.FormValue("set"))
		body.Clear, _ = strconv.Atoi(r.FormValue("clear"))
		if r.FormValue("flag") == "skip" && len(body.ChapterIDs) == 1 {
			var chapter store.Chapter
			if err := s.store.DB().WithContext(r.Context()).
				Where("plugin = ? AND slug = ? AND chapter_id = ?", body.Plugin, body.Slug, body.ChapterID).First(&chapter).Error; err == nil {
				if chapter.Flags&store.ChapterSkip != 0 {
					body.Clear = store.ChapterSkip
				} else {
					body.Set = store.ChapterSkip
				}
			}
		}
	} else if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	if s.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "store unavailable"})
		return
	}
	ids := body.ChapterIDs
	if len(ids) == 0 && body.ChapterID != "" {
		ids = []string{body.ChapterID}
	}
	if len(ids) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "missing chapter id"})
		return
	}
	for _, id := range ids {
		if err := s.store.SetChapterFlags(r.Context(), body.Plugin, body.Slug, id, body.Set, body.Clear); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) listChapters(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "store unavailable"})
		return
	}
	var rows []store.Chapter
	if err := s.store.DB().WithContext(r.Context()).
		Where("plugin = ? AND slug = ?", chi.URLParam(r, "plugin"), chi.URLParam(r, "slug")).
		Order("number, chapter_id").Find(&rows).Error; err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ChapterID)
	}
	progress, _ := s.store.ReadProgresses(r.Context(), chi.URLParam(r, "plugin"), chi.URLParam(r, "slug"), ids)
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		var count int64
		_ = s.store.DB().WithContext(r.Context()).Model(&store.ChapterPage{}).
			Where("plugin = ? AND slug = ? AND chapter_id = ?", row.Plugin, row.Slug, row.ChapterID).Count(&count).Error
		out = append(out, map[string]any{
			"id": row.ChapterID, "number": row.Number, "title": row.Title,
			"read": row.Flags&store.ChapterRead != 0, "skip": row.Flags&store.ChapterSkip != 0,
			"pages": count, "read_pages": progress[row.ChapterID],
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) renderChapters(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.NotFound(w, r)
		return
	}
	pluginID, slug := chi.URLParam(r, "plugin"), chi.URLParam(r, "slug")
	var stored []store.Chapter
	if err := s.store.DB().WithContext(r.Context()).Where("plugin = ? AND slug = ?", pluginID, slug).
		Order("number, chapter_id").Find(&stored).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	allChapters := make([]components.ChapterOption, 0, len(stored))
	for _, row := range stored {
		allChapters = append(allChapters, components.ChapterOption{ID: row.ChapterID, Number: row.Number, Title: row.Title})
	}
	total := (len(stored) + 19) / 20
	if total == 0 {
		total = 1
	}
	page, _ := strconv.Atoi(r.FormValue("page"))
	if page < 1 {
		page = 1
	}
	if page > total {
		page = total
	}
	start, end := (page-1)*20, page*20
	if end > len(stored) {
		end = len(stored)
	}
	ids := make([]string, 0, end-start)
	for _, row := range stored[start:end] {
		ids = append(ids, row.ChapterID)
	}
	progress, _ := s.store.ReadProgresses(r.Context(), pluginID, slug, ids)
	rows := make([]components.ChapterRow, 0, end-start)
	for _, row := range stored[start:end] {
		var count int64
		_ = s.store.DB().WithContext(r.Context()).Model(&store.ChapterPage{}).
			Where("plugin = ? AND slug = ? AND chapter_id = ?", pluginID, slug, row.ChapterID).Count(&count).Error
		rows = append(rows, components.ChapterRow{
			Plugin: pluginID, Slug: slug, ID: row.ChapterID, Number: row.Number,
			Title: row.Title, Read: row.Flags&store.ChapterRead != 0,
			Skip: row.Flags&store.ChapterSkip != 0, Pages: int(count),
			Progress: progress[row.ChapterID],
		})
	}
	pages := make([]components.Page, 0, len(s.Pages()))
	for _, p := range s.Pages() {
		pages = append(pages, components.Page{Path: p.Path, Title: p.Title})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.Chapters(components.Page{Path: "/chapters", Title: "Chapters"}, pages, slug, rows, page, total, allChapters).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) prefetchWindow(ctx context.Context, plugin, slug string) {
	pos, err := s.store.GetPosition(ctx, plugin, slug)
	if err != nil {
		return
	}
	for _, raw := range s.reader.Window(ctx, plugin, slug, *pos, 10) {
		key := imagecache.Hash(raw)
		if s.cacheStore.Exists(key) {
			continue
		}
		body, err := s.cacheStore.GetOrFetch(ctx, key, raw)
		if err != nil {
			continue
		}
		path := imagecache.APIPath(s.cacheStore.Root, plugin, "prefetch", key)
		_, _, _ = s.cacheStore.Put(path, body, false)
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

func decodeLibraryRequest(r *http.Request, out any) error {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			return err
		}
		if fields, ok := out.(*struct {
			Plugin string `json:"plugin"`
			Slug   string `json:"slug"`
			Name   string `json:"name"`
			Cover  string `json:"cover"`
		}); ok {
			fields.Plugin, fields.Slug = r.FormValue("plugin"), r.FormValue("slug")
			fields.Name, fields.Cover = r.FormValue("name"), r.FormValue("cover")
			return nil
		}
		if fields, ok := out.(*struct {
			Plugin string `json:"plugin"`
			Slug   string `json:"slug"`
		}); ok {
			fields.Plugin, fields.Slug = r.FormValue("plugin"), r.FormValue("slug")
			return nil
		}
	}
	return json.NewDecoder(r.Body).Decode(out)
}

func (s *Server) saveLibrary(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Plugin string `json:"plugin"`
		Slug   string `json:"slug"`
		Name   string `json:"name"`
		Cover  string `json:"cover"`
	}
	if err := decodeLibraryRequest(r, &body); err != nil {
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
	if err := decodeLibraryRequest(r, &body); err != nil {
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

func (s *Server) wsLog(w http.ResponseWriter, r *http.Request) {
	if s.streamHub == nil {
		http.NotFound(w, r)
		return
	}
	conn, err := coderwebsocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.CloseNow()
	ctx := r.Context()
	for _, line := range s.streamHub.Lines() {
		if err := conn.Write(ctx, coderwebsocket.MessageText, []byte(line)); err != nil {
			return
		}
	}
	lines, unsubscribe := s.streamHub.Subscribe()
	defer unsubscribe()
	for {
		select {
		case <-ctx.Done():
			return
		case line := <-lines:
			if err := conn.Write(ctx, coderwebsocket.MessageText, []byte(line)); err != nil {
				return
			}
		}
	}
}

func (s *Server) renderLog(w http.ResponseWriter, r *http.Request) {
	nav := make([]components.Page, 0, len(s.Pages()))
	for _, p := range s.Pages() {
		nav = append(nav, components.Page{Path: p.Path, Title: p.Title})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.Log(components.Page{Path: "/log", Title: "Log"}, nav).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderAbout(w http.ResponseWriter, r *http.Request) {
	raw, err := assetFiles.ReadFile("assets/README.md")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	nav := make([]components.Page, 0, len(s.Pages()))
	for _, p := range s.Pages() {
		nav = append(nav, components.Page{Path: p.Path, Title: p.Title})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.About(components.Page{Path: "/about", Title: "About"}, nav, readmeHTML(string(raw))).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
