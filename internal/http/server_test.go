package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	coderwebsocket "github.com/coder/websocket"

	imagecache "goisekai2/internal/cache"
	"goisekai2/internal/config"
	"goisekai2/internal/detail"
	"goisekai2/internal/events"
	"goisekai2/internal/library"
	"goisekai2/internal/plugin"
	"goisekai2/internal/reader"
	"goisekai2/internal/store"
	"goisekai2/internal/stream"
)

func testIni(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.ini")
	if err := os.WriteFile(path, []byte("[paths]\ndata_dir = "+dir+"\nlibrary = "+filepath.Join(dir, "lib.db")+"\n\n[plugin]\nplugins = "+filepath.Join(dir, "plugins")+"\n\n[server]\n; listen port\nport = 8090\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRoutesOnlyIncludeImplementedPages(t *testing.T) {
	s := New(config.Knobs(), testIni(t), nil, nil, nil, nil, nil)
	if got := s.Pages(); len(got) != 5 || got[0].Path != "/settings" || got[1].Path != "/log" || got[2].Path != "/about" || got[3].Path != "/library" || got[4].Path != "/search" {
		t.Fatalf("pages = %v, want settings/log/about/library/search", got)
	}

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/settings", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `href="/settings" aria-current="page"`) {
		t.Fatalf("active Settings link missing: %s", body)
	}
	for _, dead := range []string{`href="/plugins"`, `href="/reader"`} {
		if strings.Contains(body, dead) {
			t.Fatalf("dead nav link %q rendered", dead)
		}
	}
}

func TestGetConfigReflectsFileEdits(t *testing.T) {
	path := testIni(t)
	s := New(config.Knobs(), path, nil, nil, nil, nil, nil)
	get := func() []config.Projection {
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/config", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d", rec.Code)
		}
		var out []config.Projection
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	first := get()
	if len(first) == 0 || first[0].Section != "server" || first[0].Value != "8090" || first[0].Source != "ini" {
		t.Fatalf("projection = %+v", first)
	}
	if err := os.WriteFile(path, []byte("[server]\nport = 7777\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	second := get()
	if second[0].Value != "7777" {
		t.Fatalf("stale projection: %+v", second[0])
	}
}

func TestPostConfigMergeSaveAndReject(t *testing.T) {
	path := testIni(t)
	s := New(config.Knobs(), path, nil, nil, nil, nil, nil)
	post := func(body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/config", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		s.Handler().ServeHTTP(rec, req)
		return rec
	}

	rec := post(`{"server.port":"9000"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"written":1`) {
		t.Fatalf("success = %d %s", rec.Code, rec.Body.String())
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "; listen port") || !strings.Contains(string(after), "port = 9000") {
		t.Fatalf("comment or value lost: %q", after)
	}

	before, _ := os.ReadFile(path)
	rec = post(`{"server.port":"http"}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "server.port") {
		t.Fatalf("reject = %d %s", rec.Code, rec.Body.String())
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(before, got) {
		t.Fatalf("rejected save changed the file: %q -> %q", before, got)
	}
}

func TestServeAndShutdown(t *testing.T) {
	s := New(config.Knobs(), testIni(t), nil, nil, nil, nil, nil)
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- s.Serve(l) }()

	resp, err := http.Get("http://" + l.Addr().String() + "/settings")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil && err != http.ErrServerClosed {
		t.Fatalf("serve: %v", err)
	}
}

func TestPostConfigPublishesAfterWriteAndAppliesLive(t *testing.T) {
	config.ResetAppliers()
	t.Cleanup(config.ResetAppliers)
	levelVar := new(slog.LevelVar)
	config.RegisterLiveApplier("log.level", func(v string) error {
		var level slog.Level
		if err := level.UnmarshalText([]byte(v)); err != nil {
			return err
		}
		levelVar.Set(level)
		return nil
	})

	path := filepath.Join(t.TempDir(), "cfg.ini")
	if err := os.WriteFile(path, []byte("[log]\nlevel = info\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	bus := events.New()
	t.Cleanup(bus.Close)
	rec := bus.Record(events.ConfigChanged)
	s := New(config.Knobs(), path, bus, nil, nil, nil, nil)

	post := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/config", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, req)
		return rec
	}

	recHTTP := post(`{"log.level":"debug"}`)
	if recHTTP.Code != http.StatusOK {
		t.Fatalf("status = %d %s", recHTTP.Code, recHTTP.Body.String())
	}
	onDisk, _ := os.ReadFile(path)
	if !strings.Contains(string(onDisk), "level = debug") {
		t.Fatalf("file not written before event: %q", onDisk)
	}
	items := rec.Items()
	if len(items) != 1 {
		t.Fatalf("events = %d, want 1", len(items))
	}
	payload, ok := items[0].(events.ConfigChangedPayload)
	if !ok || payload.Key != "log.level" || payload.Old != "info" || payload.New != "debug" {
		t.Fatalf("payload = %+v", items[0])
	}
	if levelVar.Level() != slog.LevelDebug {
		t.Fatalf("level = %v, want debug", levelVar.Level())
	}
	if strings.Contains(recHTTP.Body.String(), `"log.level"`) {
		t.Fatalf("applied key listed as restart-required: %s", recHTTP.Body.String())
	}

	post(`{"log.level":"loud"}`)
	if n := len(rec.Items()); n != 1 {
		t.Fatalf("rejected save published %d events, want 0 more", n-1)
	}
}

func TestPluginsPageAndActions(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "plugins")
	folder := filepath.Join(pluginDir, "p")
	if err := os.MkdirAll(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := `{"name":"Fixture","version":"1","runtime":"lua","capabilities":["search"],"api":1,"entry":"plugin.lua"}`
	if err := os.WriteFile(filepath.Join(folder, "main.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "plugin.lua"), []byte(`function handle(req) return '{"ok":true}' end`), 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(filepath.Join(dir, "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	host := plugin.NewHost(st, nil, pluginDir, nil)
	ctx := context.Background()
	if err := host.Boot(ctx); err != nil {
		t.Fatal(err)
	}
	if err := host.SetEnabled(ctx, "p", true); err != nil {
		t.Fatal(err)
	}
	s := New(config.Knobs(), testIni(t), nil, host, nil, nil, nil)

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/plugins", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Fixture") ||
		!strings.Contains(rec.Body.String(), `href="/library?plugin=p"`) {
		t.Fatalf("plugins page: %d %s", rec.Code, rec.Body.String())
	}

	for _, path := range []string{"/plugins/p/toggle", "/plugins/p/reload"} {
		action := httptest.NewRecorder()
		s.Handler().ServeHTTP(action, httptest.NewRequest(http.MethodPost, path, nil))
		if action.Code != http.StatusOK {
			t.Fatalf("%s = %d", path, action.Code)
		}
	}
}

func TestLibraryRoutes(t *testing.T) {
	path := testIni(t)
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	s := New(config.Knobs(), path, nil, nil, library.New(st, nil), nil, st)

	post := func(url, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, req)
		return rec
	}
	rec := post("/api/library", `{"plugin":"p","slug":"s","name":"Name","cover":"c"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"saved":true`) {
		t.Fatalf("save = %d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/library", nil))
	var rows []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0]["name"] != "Name" || rows[0]["total_ch"] != float64(0) {
		t.Fatalf("listing = %v", rows)
	}
	if _, exists := rows[0]["has_update"]; exists {
		t.Fatal("has_update leaked into listing")
	}

	page := httptest.NewRecorder()
	s.Handler().ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/library", nil))
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), `data-status="`) {
		t.Fatalf("grid = %d %s", page.Code, page.Body.String())
	}

	rec = post("/api/library/remove", `{"plugin":"p","slug":"s"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"saved":false`) {
		t.Fatalf("remove = %d %s", rec.Code, rec.Body.String())
	}
}

func TestAssetsAreServedLocally(t *testing.T) {
	s := New(config.Knobs(), testIni(t), nil, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/settings", nil))
	for _, attr := range []string{`src="http`, `href="http`, `src="//`, `href="//`} {
		if strings.Contains(rec.Body.String(), attr) {
			t.Fatalf("page references an external asset (%s): %s", attr, rec.Body.String())
		}
	}
	for _, path := range []string{"/assets/vendor/htmx.min.js", "/assets/vendor/alpine.min.js"} {
		assetRec := httptest.NewRecorder()
		s.Handler().ServeHTTP(assetRec, httptest.NewRequest(http.MethodGet, path, nil))
		if assetRec.Code != http.StatusOK || assetRec.Body.Len() == 0 {
			t.Fatalf("asset %s = %d (%d bytes)", path, assetRec.Code, assetRec.Body.Len())
		}
	}
}

func TestPostConfigHTMXFragment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.ini")
	if err := os.WriteFile(path, []byte("[log]\nlevel = info\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := New(config.Knobs(), path, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/config", strings.NewReader("log.level=debug"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Saved 1 change") {
		t.Fatalf("status fragment missing: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"written"`) {
		t.Fatalf("htmx response should be HTML, got JSON: %s", rec.Body.String())
	}
	onDisk, _ := os.ReadFile(path)
	if !strings.Contains(string(onDisk), "level = debug") {
		t.Fatalf("form values not saved: %q", onDisk)
	}
}

func TestSearchPageUsesEnabledPlugins(t *testing.T) {
	s := New(config.Knobs(), testIni(t), nil, nil, nil, nil, nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/search?q=manga", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `name="q"`) || !strings.Contains(rec.Body.String(), `name="plugin"`) {
		t.Fatalf("search page = %d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `href="/library?plugin=demo"`) {
		t.Fatal("unexpected plugin option")
	}
}

func TestReaderPagesAndPositionRoutes(t *testing.T) {
	path := testIni(t)
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if err := st.UpsertChapters(ctx, "p", "s", []store.Chapter{
		{ChapterID: "a", Number: 1}, {ChapterID: "b", Number: 2},
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.ReplacePages(ctx, "p", "s", "b", []store.ChapterPage{{URL: "one"}, {URL: "two"}}); err != nil {
		t.Fatal(err)
	}
	s := New(config.Knobs(), path, nil, nil, nil, reader.New(st, nil), st)

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/pages/p/s/2", nil))
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "/cache/img?") || !strings.Contains(body, "url=one") ||
		!strings.Contains(body, "ch=2") || !strings.Contains(body, "page=1") {
		t.Fatalf("pages = %d %s", rec.Code, body)
	}
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/pages/p/s/9", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown chapter = %d", rec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/reader/position", strings.NewReader(`{"plugin":"p","slug":"s","chapter":2,"chapter_id":"b","page":1}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("position = %d %s", rec.Code, rec.Body.String())
	}
	pos, err := st.GetPosition(ctx, "p", "s")
	if err != nil || pos.ChapterNumber != 2 || pos.Page != 1 {
		t.Fatalf("position = %+v, %v", pos, err)
	}

	rec = httptest.NewRecorder()
	flagReq := httptest.NewRequest(http.MethodPost, "/chapter/flags", strings.NewReader(`{"plugin":"p","slug":"s","chapter_id":"b","set":2,"clear":0}`))
	flagReq.Header.Set("Content-Type", "application/json")
	s.Handler().ServeHTTP(rec, flagReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("flags = %d %s", rec.Code, rec.Body.String())
	}
	var chapter store.Chapter
	if err := st.DB().Where("plugin = ? AND slug = ? AND chapter_id = ?", "p", "s", "b").First(&chapter).Error; err != nil || chapter.Flags != store.ChapterSkip {
		t.Fatalf("chapter flags = %d, %v", chapter.Flags, err)
	}

	page := httptest.NewRecorder()
	s.Handler().ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/read/p/s?ch=2", nil))
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), `data-mode="`) ||
		!strings.Contains(page.Body.String(), "ch=2") || !strings.Contains(page.Body.String(), "page=1") ||
		!strings.Contains(page.Body.String(), `aria-label="Chapter"`) || !strings.Contains(page.Body.String(), `aria-label="Page"`) ||
		!strings.Contains(page.Body.String(), "prevRef") || !strings.Contains(page.Body.String(), "nextRef") ||
		!strings.Contains(page.Body.String(), "h-auto w-full max-w-none") ||
		!strings.Contains(page.Body.String(), "items-start") {
		t.Fatalf("reader page = %d %s", page.Code, page.Body.String())
	}

	chapters := httptest.NewRecorder()
	s.Handler().ServeHTTP(chapters, httptest.NewRequest(http.MethodGet, "/api/chapters/p/s", nil))
	if chapters.Code != http.StatusOK || !strings.Contains(chapters.Body.String(), `"pages":2`) ||
		!strings.Contains(chapters.Body.String(), `"read_pages":1`) {
		t.Fatalf("chapters = %d %s", chapters.Code, chapters.Body.String())
	}

	list := httptest.NewRecorder()
	s.Handler().ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/chapters/p/s", nil))
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `/read/p/s?ch=`) ||
		!strings.Contains(list.Body.String(), "Mark all read") || !strings.Contains(list.Body.String(), "Read progress") ||
		!strings.Contains(list.Body.String(), "chapter-boot") {
		t.Fatalf("chapter page = %d %s", list.Code, list.Body.String())
	}
}

func TestChapterFlagsBulkForm(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if err := st.UpsertChapters(ctx, "p", "s", []store.Chapter{
		{ChapterID: "a", Number: 1}, {ChapterID: "b", Number: 2}, {ChapterID: "c", Number: 3},
	}); err != nil {
		t.Fatal(err)
	}
	path := testIni(t)
	s := New(config.Knobs(), path, nil, nil, nil, nil, st)
	body := "plugin=p&slug=s&ch_id=a&ch_id=b&set=1"
	req := httptest.NewRequest(http.MethodPost, "/chapter/flags", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("flags = %d %s", rec.Code, rec.Body.String())
	}
	for _, id := range []string{"a", "b"} {
		var chapter store.Chapter
		if err := st.DB().Where("plugin = ? AND slug = ? AND chapter_id = ?", "p", "s", id).First(&chapter).Error; err != nil {
			t.Fatal(err)
		}
		if chapter.Flags&store.ChapterRead == 0 {
			t.Errorf("%s read bit not set: %d", id, chapter.Flags)
		}
	}
	var last store.Chapter
	if err := st.DB().Where("plugin = ? AND slug = ? AND chapter_id = ?", "p", "s", "c").First(&last).Error; err != nil {
		t.Fatal(err)
	}
	if last.Flags&store.ChapterRead != 0 {
		t.Errorf("c unexpectedly marked read: %d", last.Flags)
	}
}

func TestDetailRoutes(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "lib.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.UpsertSaved(context.Background(), &store.SavedManga{
		Plugin: "p", Slug: "s", Name: "Name", Cover: "cover",
	}); err != nil {
		t.Fatal(err)
	}
	s := New(config.Knobs(), testIni(t), nil, nil, nil, nil, st)
	s.SetDetail(detail.New(st, nil))

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/detail/p/s", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"cached":true`) {
		t.Fatalf("detail = %d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/view/manga/p/s", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Name") {
		t.Fatalf("page = %d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/refresh/p/s", nil))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("refresh = %d %s", rec.Code, rec.Body.String())
	}
}

func TestCacheImageRoute(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for x := 0; x < 4; x++ {
		for y := 0; y < 4; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 40), G: uint8(y * 40), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write(buf.Bytes())
	}))
	defer upstream.Close()

	root := t.TempDir()
	pool := imagecache.NewPool(root, 1<<20)
	s := New(config.Knobs(), testIni(t), nil, nil, nil, nil, nil)
	s.SetCache(imagecache.NewStore(root, "webp", 800, pool))

	bad := httptest.NewRecorder()
	s.Handler().ServeHTTP(bad, httptest.NewRequest(http.MethodGet, "/cache/img?url=x", nil))
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("missing context = %d", bad.Code)
	}

	target := "/cache/img?plugin=p&slug=s&ch=1&ch_id=abcd&page=1&url=" + upstream.URL + "/a.png"
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/webp" {
			t.Fatalf("hit %d = %d %s", i, rec.Code, rec.Header().Get("Content-Type"))
		}
	}
	if calls != 1 {
		t.Fatalf("upstream calls = %d, want 1", calls)
	}
}

func TestPluginLogoRouteAlias(t *testing.T) {
	s := New(config.Knobs(), testIni(t), nil, nil, nil, nil, nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/plugin/mangadex/logo", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("logo alias = %d, want 404 without a plugin host", rec.Code)
	}
}

func TestAboutAndLogRoutes(t *testing.T) {
	s := New(config.Knobs(), testIni(t), nil, nil, nil, nil, nil)
	s.SetStream(stream.New(20))
	for _, path := range []string{"/about", "/log"} {
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s = %d", path, rec.Code)
		}
	}
	found := false
	for _, p := range s.Pages() {
		if p.Path == "/about" {
			found = true
		}
	}
	if !found {
		t.Fatal("about not in navigation")
	}
}

func TestWebSocketLogReplay(t *testing.T) {
	hub := stream.New(20)
	hub.Publish("HELLO_LOG")
	s := New(config.Knobs(), testIni(t), nil, nil, nil, nil, nil)
	s.SetStream(hub)
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/log"
	conn, _, err := coderwebsocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	_, data, err := conn.Read(ctx)
	if err != nil || !strings.Contains(string(data), "HELLO_LOG") {
		t.Fatalf("frame = %q, %v", data, err)
	}
}
