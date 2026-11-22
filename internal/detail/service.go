// Package detail resolves manga metadata and chapters from saved rows or plugins.
package detail

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"goisekai2/internal/plugin"
	"goisekai2/internal/store"
	"goisekai2/internal/updates"
)

// ErrUnknownTitle reports a slug the host cannot resolve.
var ErrUnknownTitle = errors.New("unknown title")

// Service owns detail resolution, stale-serve and refresh.
type Service struct {
	store      *store.Store
	host       *plugin.Host
	enrichment []*EnrichmentSource
	attempt    func(context.Context, *store.Store, string, string, bool, string, int64)
}

// New creates a detail service.
func New(st *store.Store, host *plugin.Host) *Service {
	return &Service{store: st, host: host, attempt: updates.StampAttempt}
}

// SetEnrichment attaches the separately loaded enrichment registry.
func (s *Service) SetEnrichment(sources []*EnrichmentSource) { s.enrichment = sources }

func cacheHash(pluginID, method, slug string) string {
	sum := sha256.Sum256([]byte(pluginID + "\x00" + method + "\x00" + slug))
	return hex.EncodeToString(sum[:])[:16]
}

func envelope(raw []byte) (map[string]any, error) {
	var env map[string]any
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	if okValue, present := env["ok"]; present {
		if ok, _ := okValue.(bool); !ok {
			return nil, fmt.Errorf("plugin envelope: ok=false")
		}
	}
	var data map[string]any
	switch raw := env["data"].(type) {
	case map[string]any:
		data = raw
	case []any:
		data = map[string]any{"chapters": raw}
	default:
		return nil, fmt.Errorf("plugin envelope: missing data")
	}
	data["cached"] = false
	data["saved"] = false
	return data, nil
}

func (s *Service) fetch(ctx context.Context, pluginID, method, slug string) (map[string]any, error) {
	params, _ := json.Marshal(map[string]string{"slug": slug, "id": slug})
	var err error
	if s.host != nil {
		var raw []byte
		raw, err = s.host.Call(ctx, pluginID, method, params)
		if err == nil {
			preview := raw
			if len(preview) > 300 {
				preview = preview[:300]
			}
			slog.Debug("plugin envelope preview", "plugin", pluginID, "method", method, "preview", string(preview))
			data, parseErr := envelope(raw)
			if parseErr == nil {
				if method == "detail" {
					canonicalizeDetail(data)
				}
				_ = s.store.PutCached(ctx, &store.ApiCache{
					PluginID: pluginID, QueryHash: cacheHash(pluginID, method, slug),
					Payload: string(raw), FetchedAt: time.Now().Unix(), TTL: 86400,
				})
				if method == "detail" {
					s.materialize(ctx, pluginID, slug, data)
				}
				return data, nil
			}
			err = parseErr
		}
	} else {
		err = fmt.Errorf("plugin %s unavailable", pluginID)
	}
	if cached, ok := s.store.GetCached(ctx, pluginID, cacheHash(pluginID, method, slug), time.Now().Unix()); ok {
		data, parseErr := envelope([]byte(cached))
		if parseErr == nil {
			data["cached"] = true
			return data, nil
		}
	}
	return nil, err
}

// ApplyCandidate stores a user-picked metadata candidate. Plugin fields are
// never overwritten automatically.
func (s *Service) ApplyCandidate(ctx context.Context, pluginID, slug string, candidate map[string]any) error {
	meta, err := s.store.GetTitleMeta(ctx, pluginID, slug)
	if err != nil {
		meta = &store.TitleMeta{Plugin: pluginID, Slug: slug}
	}
	raw, err := json.Marshal(candidate)
	if err != nil {
		return err
	}
	meta.Fetched = string(raw)
	if cover, ok := candidate["cover"].(string); ok && cover != "" {
		meta.Chosen = cover
	}
	meta.FetchedAt = time.Now().Unix()
	return s.store.PutTitleMeta(ctx, meta)
}

// ResetDetail clears stored enrichment metadata without touching the cover.
func (s *Service) ResetDetail(ctx context.Context, pluginID, slug string) error {
	meta, err := s.store.GetTitleMeta(ctx, pluginID, slug)
	if err != nil {
		return nil
	}
	meta.Fetched, meta.FetchedAt = "", 0
	return s.store.PutTitleMeta(ctx, meta)
}

// ResetCover clears a user-chosen cover without touching metadata.
func (s *Service) ResetCover(ctx context.Context, pluginID, slug string) error {
	meta, err := s.store.GetTitleMeta(ctx, pluginID, slug)
	if err != nil {
		return nil
	}
	meta.Chosen = ""
	return s.store.PutTitleMeta(ctx, meta)
}

// Related returns saved titles sharing at least one stored tag.
func (s *Service) Related(ctx context.Context, pluginID, slug string) []map[string]any {
	out := make([]map[string]any, 0)
	var rows []store.SavedManga
	if err := s.store.DB().WithContext(ctx).Find(&rows).Error; err != nil {
		return out
	}
	wanted := tagSet(s.pluginTag(ctx, pluginID, slug))
	if len(wanted) == 0 {
		return out
	}
	for _, row := range rows {
		if row.Plugin == pluginID && row.Slug == slug {
			continue
		}
		if intersects(wanted, tagSet(row.Tags)) {
			out = append(out, map[string]any{"plugin": row.Plugin, "slug": row.Slug, "name": row.Name})
		}
		if len(out) >= 10 {
			break
		}
	}
	return out
}

func (s *Service) pluginTag(ctx context.Context, pluginID, slug string) string {
	row, err := s.store.GetSaved(ctx, pluginID, slug)
	if err != nil {
		return ""
	}
	return row.Tags
}

func tagSet(raw string) map[string]bool {
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		tags = strings.Fields(raw)
	}
	out := make(map[string]bool, len(tags))
	for _, tag := range tags {
		out[strings.ToLower(strings.TrimSpace(tag))] = true
	}
	return out
}

func intersects(a, b map[string]bool) bool {
	for key := range a {
		if b[key] {
			return true
		}
	}
	return false
}

// Candidates asks each enrichment source for metadata candidates. Failures
// are skipped so one dead source does not break the detail page.
func (s *Service) Candidates(ctx context.Context, title, kind string) []map[string]any {
	out := make([]map[string]any, 0)
	for _, source := range s.enrichment {
		data, err := source.Enrich(ctx, map[string]string{"title": title, "kind": kind})
		if err != nil {
			continue
		}
		out = append(out, data)
	}
	return out
}

// Detail returns a saved row without calling the plugin, or resolves through
// the plugin with stale fallback for unsaved titles.
func (s *Service) Detail(ctx context.Context, pluginID, slug string) (map[string]any, error) {
	row, err := s.store.GetSaved(ctx, pluginID, slug)
	if err == nil {
		if row.AddedAt == 0 && (row.Synopsis == "" || row.Authors == "" || row.Tags == "") {
			if _, fetchErr := s.fetch(ctx, pluginID, "detail", slug); fetchErr == nil {
				if fresh, getErr := s.store.GetSaved(ctx, pluginID, slug); getErr == nil {
					row = fresh
				}
			}
		}
		out := map[string]any{
			"plugin": row.Plugin, "slug": row.Slug, "name": row.Name, "cover": row.Cover,
			"status": row.Status, "synopsis": row.Synopsis, "alt_titles": row.AltTitles,
			"authors": row.Authors, "artists": row.Artists, "tags": row.Tags,
			"cached": true, "saved": row.AddedAt > 0,
			"enrichment": map[string]any{}, "chosen_cover": row.Cover,
		}
		if meta, err := s.store.GetTitleMeta(ctx, pluginID, slug); err == nil {
			out["enrichment"] = meta.Fetched
		}
		return out, nil
	}
	data, err := s.fetch(ctx, pluginID, "detail", slug)
	if err != nil {
		return nil, fmt.Errorf("%w: %s/%s: %v", ErrUnknownTitle, pluginID, slug, err)
	}
	return data, nil
}

// canonicalizeDetail maps plugin-specific field names to the frozen
// detail vocabulary for every runtime, without hardcoding site strings.
func canonicalizeDetail(data map[string]any) {
	data["synopsis"] = detailString(data, "synopsis", "description")
	data["authors"] = detailString(data, "authors", "author")
	data["artists"] = detailString(data, "artists", "artist")
	data["tags"] = detailString(data, "tags", "genres", "genre")
	data["alt_titles"] = detailString(data, "alt_titles", "altTitles", "other_names")
}

func detailString(data map[string]any, keys ...string) string {
	for _, key := range keys {
		switch v := data[key].(type) {
		case string:
			if v != "" {
				return v
			}
		case []any:
			parts := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok && s != "" {
					parts = append(parts, s)
				}
			}
			if len(parts) > 0 {
				return strings.Join(parts, ", ")
			}
		}
	}
	return ""
}

// materialize keeps saved_manga as the single master row for title metadata.
func (s *Service) materialize(ctx context.Context, pluginID, slug string, data map[string]any) {
	row, err := s.store.GetSaved(ctx, pluginID, slug)
	if err != nil {
		row = &store.SavedManga{Plugin: pluginID, Slug: slug, AddedAt: 0}
	}
	get := func(keys ...string) string {
		for _, key := range keys {
			if v, ok := data[key].(string); ok && v != "" {
				return v
			}
		}
		return ""
	}
	if v := get("name", "title"); v != "" {
		row.Name = v
	}
	if v := get("cover", "cover_url"); v != "" {
		row.Cover = v
	}
	if v := get("status"); v != "" {
		row.Status = v
	}
	if v := get("synopsis", "description"); v != "" {
		row.Synopsis = v
	}
	if v := get("alt_titles"); v != "" {
		row.AltTitles = v
	}
	if v := detailString(data, "authors", "author"); v != "" {
		row.Authors = v
	}
	if v := detailString(data, "artists", "artist"); v != "" {
		row.Artists = v
	}
	if v := detailString(data, "tags", "genres", "genre"); v != "" {
		row.Tags = v
	}
	if n, ok := data["total_ch"].(float64); ok && int(n) > row.TotalCh {
		row.TotalCh = int(n)
	}
	if err := s.store.UpsertSaved(ctx, row); err != nil {
		slog.Debug("detail materialize failed", "plugin", pluginID, "slug", slug, "error", err)
	}
}

// Chapters resolves and persists the chapter list for saved or unsaved titles.
func (s *Service) Chapters(ctx context.Context, pluginID, slug string) ([]store.Chapter, error) {
	data, err := s.fetch(ctx, pluginID, "chapters", slug)
	if err != nil {
		return nil, err
	}
	return s.storeChapters(ctx, pluginID, slug, data)
}

func (s *Service) storeChapters(ctx context.Context, pluginID, slug string, data map[string]any) ([]store.Chapter, error) {
	raw, err := json.Marshal(data["chapters"])
	if err != nil {
		return nil, err
	}
	var parsed []struct {
		ID        string  `json:"id"`
		ChapterID string  `json:"chapter_id"`
		Number    float64 `json:"number"`
		Title     string  `json:"title"`
		URL       string  `json:"url"`
		ReleaseAt int64   `json:"release_at"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	rows := make([]store.Chapter, 0, len(parsed))
	for _, item := range parsed {
		id := item.ChapterID
		if id == "" {
			id = item.ID
		}
		rows = append(rows, store.Chapter{
			Plugin: pluginID, Slug: slug, ChapterID: id,
			Number: item.Number, Title: item.Title, URL: item.URL, UpdatedAt: item.ReleaseAt,
		})
	}
	if len(rows) > 0 {
		if err := s.store.UpsertChapters(ctx, pluginID, slug, rows); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

// Refresh resolves detail and chapters, then stores both only when both pass
// validation. The failure is named so callers can return a precise 502.
func (s *Service) Refresh(ctx context.Context, pluginID, slug string) error {
	detailData, err := s.fetch(ctx, pluginID, "detail", slug)
	if err != nil {
		s.stamp(ctx, pluginID, slug, false, err.Error())
		return fmt.Errorf("detail: %w", err)
	}
	chapterData, err := s.fetch(ctx, pluginID, "chapters", slug)
	if err != nil {
		s.stamp(ctx, pluginID, slug, false, err.Error())
		return fmt.Errorf("chapters: %w", err)
	}
	for method, data := range map[string]map[string]any{"detail": detailData, "chapters": chapterData} {
		raw, _ := json.Marshal(data)
		if err := s.store.PutCached(ctx, &store.ApiCache{
			PluginID: pluginID, QueryHash: cacheHash(pluginID, method, slug),
			Payload: string(raw), FetchedAt: time.Now().Unix(), TTL: 86400,
		}); err != nil {
			s.stamp(ctx, pluginID, slug, false, err.Error())
			return err
		}
		if method == "chapters" {
			if _, err := s.storeChapters(ctx, pluginID, slug, data); err != nil {
				s.stamp(ctx, pluginID, slug, false, err.Error())
				return err
			}
		}
	}
	s.stamp(ctx, pluginID, slug, true, "")
	return nil
}

func (s *Service) stamp(ctx context.Context, pluginID, slug string, ok bool, errText string) {
	s.attempt(ctx, s.store, pluginID, slug, ok, errText, time.Now().Unix())
}
