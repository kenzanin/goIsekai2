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
	"strconv"
	"strings"
	"sync/atomic"
	"time"

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
	cacheStore *imagecache.Store
	detailSvc  *detail.Service
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
	r.Get("/view/manga/{plugin}/{slug}", s.renderDetail)
	r.Get("/api/detail/{plugin}/{slug}", s.getDetail)
	r.Post("/api/refresh/{plugin}/{slug}", s.refreshDetail)
	r.Post("/api/refresh/{plugin}/{slug}/force", s.refreshDetail)
	r.Post("/api/detail/{plugin}/{slug}/reset", s.resetDetail)
	r.Post("/api/detail/{plugin}/{slug}/apply", s.applyCandidate)
	r.Get("/cache/img", s.getCacheImage)
	r.Get("/", s.renderLibrary)
	r.Get("/library", s.renderLibrary)
	r.Get("/search", s.renderSearch)
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
	for _, data := range s.detailSvc.Candidates(r.Context(), chi.URLParam(r, "slug"), "detail") {
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
	}
	if authors, ok := data["authors"].([]string); ok {
		row.Authors = authors
	}
	if tags, ok := data["tags"].([]string); ok {
		row.Tags = tags
	}
	chapters := make([]components.ChapterRow, 0)
	var stored []store.Chapter
	_ = s.store.DB().WithContext(r.Context()).Where("plugin = ? AND slug = ?", pluginID, slug).
		Order("number, chapter_id").Find(&stored)
	for _, ch := range stored {
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
	if err := components.Detail(components.Page{Path: "/view", Title: row.Name}, pages, row, chapters, candidates).Render(r.Context(), w); err != nil {
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
		body, err := s.cacheStore.GetOrFetch(r.Context(), imagecache.Hash(rawURL), rawURL)
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
		for i, u := range urls {
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
	active := components.Page{Path: "/read", Title: "Reader"}
	pages := make([]components.Page, 0, len(s.Pages()))
	for _, p := range s.Pages() {
		pages = append(pages, components.Page{Path: p.Path, Title: p.Title})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.Reader(active, pages, name, chapter, mode, direction, page).Render(r.Context(), w); err != nil {
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
		Plugin    string `json:"plugin"`
		Slug      string `json:"slug"`
		ChapterID string `json:"chapter_id"`
		Set       int    `json:"set"`
		Clear     int    `json:"clear"`
	}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		body.Plugin, body.Slug = r.FormValue("plugin"), r.FormValue("slug")
		body.ChapterID = r.FormValue("ch_id")
		if body.ChapterID == "" {
			body.ChapterID = r.FormValue("chapter_id")
		}
		if r.FormValue("flag") == "skip" {
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
	if err := s.store.SetChapterFlags(r.Context(), body.Plugin, body.Slug, body.ChapterID, body.Set, body.Clear); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
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
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		var count int64
		_ = s.store.DB().WithContext(r.Context()).Model(&store.ChapterPage{}).
			Where("plugin = ? AND slug = ? AND chapter_id = ?", row.Plugin, row.Slug, row.ChapterID).Count(&count).Error
		out = append(out, map[string]any{
			"id": row.ChapterID, "number": row.Number, "title": row.Title,
			"read": row.Flags&store.ChapterRead != 0, "skip": row.Flags&store.ChapterSkip != 0,
			"pages": count,
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
	rows := make([]components.ChapterRow, 0, end-start)
	for _, row := range stored[start:end] {
		var count int64
		_ = s.store.DB().WithContext(r.Context()).Model(&store.ChapterPage{}).
			Where("plugin = ? AND slug = ? AND chapter_id = ?", pluginID, slug, row.ChapterID).Count(&count).Error
		rows = append(rows, components.ChapterRow{
			Plugin: pluginID, Slug: slug, ID: row.ChapterID, Number: row.Number,
			Title: row.Title, Read: row.Flags&store.ChapterRead != 0,
			Skip: row.Flags&store.ChapterSkip != 0, Pages: int(count),
		})
	}
	pages := make([]components.Page, 0, len(s.Pages()))
	for _, p := range s.Pages() {
		pages = append(pages, components.Page{Path: p.Path, Title: p.Title})
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := components.Chapters(components.Page{Path: "/chapters", Title: "Chapters"}, pages, slug, rows, page, total).Render(r.Context(), w); err != nil {
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
