// Package detail resolves manga metadata and chapters from saved rows or plugins.
package detail

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"goisekai2/internal/plugin"
	"goisekai2/internal/store"
)

// ErrUnknownTitle reports a slug the host cannot resolve.
var ErrUnknownTitle = errors.New("unknown title")

// Service owns detail resolution, stale-serve and refresh.
type Service struct {
	store *store.Store
	host  *plugin.Host
}

// New creates a detail service.
func New(st *store.Store, host *plugin.Host) *Service {
	return &Service{store: st, host: host}
}

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
