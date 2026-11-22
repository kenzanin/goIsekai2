// Package detail resolves manga metadata and chapters from saved rows or plugins.
package detail

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"goisekai2/internal/plugin"
	"goisekai2/internal/store"
)

// ErrUnknownTitle reports a slug the host cannot resolve.
var ErrUnknownTitle = errors.New("unknown title")

// Service owns detail resolution, stale-serve and refresh.
type Service struct {
	store      *store.Store
	host       *plugin.Host
	enrichment []*EnrichmentSource
}

// New creates a detail service.
func New(st *store.Store, host *plugin.Host) *Service {
	return &Service{store: st, host: host}
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
	if ok, _ := env["ok"].(bool); !ok {
		return nil, fmt.Errorf("plugin envelope: ok=false")
	}
	data, ok := env["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("plugin envelope: missing data")
	}
	data["cached"] = false
	return data, nil
}

func (s *Service) fetch(ctx context.Context, pluginID, method, slug string) (map[string]any, error) {
	params, _ := json.Marshal(map[string]string{"slug": slug, "id": slug})
	var err error
	if s.host != nil {
		var raw []byte
		raw, err = s.host.Call(ctx, pluginID, method, params)
		if err == nil {
			data, parseErr := envelope(raw)
			if parseErr == nil {
				_ = s.store.PutCached(ctx, &store.ApiCache{
					PluginID: pluginID, QueryHash: cacheHash(pluginID, method, slug),
					Payload: string(raw), FetchedAt: time.Now().Unix(), TTL: 86400,
				})
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
		out := map[string]any{
			"plugin": row.Plugin, "slug": row.Slug, "name": row.Name, "cover": row.Cover,
			"status": row.Status, "synopsis": row.Synopsis, "alt_titles": row.AltTitles,
			"authors": row.Authors, "artists": row.Artists, "tags": row.Tags,
			"cached": true, "enrichment": map[string]any{}, "chosen_cover": row.Cover,
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

// Refresh resolves detail and chapters, then stores both only when both pass
// validation. The failure is named so callers can return a precise 502.
func (s *Service) Refresh(ctx context.Context, pluginID, slug string) error {
	detailData, err := s.fetch(ctx, pluginID, "detail", slug)
	if err != nil {
		s.stamp(ctx, pluginID, slug, "error")
		return fmt.Errorf("detail: %w", err)
	}
	chapterData, err := s.fetch(ctx, pluginID, "chapters", slug)
	if err != nil {
		s.stamp(ctx, pluginID, slug, "error")
		return fmt.Errorf("chapters: %w", err)
	}
	for method, data := range map[string]map[string]any{"detail": detailData, "chapters": chapterData} {
		raw, _ := json.Marshal(data)
		if err := s.store.PutCached(ctx, &store.ApiCache{
			PluginID: pluginID, QueryHash: cacheHash(pluginID, method, slug),
			Payload: string(raw), FetchedAt: time.Now().Unix(), TTL: 86400,
		}); err != nil {
			s.stamp(ctx, pluginID, slug, "error")
			return err
		}
	}
	s.stamp(ctx, pluginID, slug, "ok")
	return nil
}

func (s *Service) stamp(ctx context.Context, pluginID, slug, status string) {
	_ = s.store.PutUpdateAttempt(ctx, &store.UpdateAttempt{
		Plugin: pluginID, Slug: slug, Outcome: status, LastAttemptAt: time.Now().Unix(),
	})
}
