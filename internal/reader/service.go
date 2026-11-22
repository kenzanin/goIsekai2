package reader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"goisekai2/internal/plugin"
	"goisekai2/internal/store"
)

// ErrUnknownChapter reports a pages request that matched no stored chapter.
var ErrUnknownChapter = errors.New("reader: unknown chapter")

// Service resolves reader pages and positions on top of the plugin host and
// the typed store.
type Service struct {
	Store *store.Store
	Host  *plugin.Host
}

// New creates a reader service.
func New(st *store.Store, host *plugin.Host) *Service {
	return &Service{Store: st, Host: host}
}

// Pages resolves one chapter entry (exact chapterID first, else the first
// stored chapter with that number) and returns its number, page URLs and
// positional neighbours from the number-ordered chapter list, so duplicated
// numbers stay individually reachable. A failed or empty host reply falls
// back to the last stored page list.
func (s *Service) Pages(ctx context.Context, plugin, slug string, number float64, chapterID string) (map[string]any, error) {
	chs, err := s.chapters(ctx, plugin, slug)
	if err != nil {
		return nil, err
	}
	at := locate(chs, number, chapterID)
	if at < 0 {
		return nil, fmt.Errorf("%w: %s/%s number %v id %q", ErrUnknownChapter, plugin, slug, number, chapterID)
	}
	urls, err := s.pageList(ctx, plugin, slug, chs[at])
	if err != nil {
		return nil, err
	}
	out := map[string]any{"chapter": chs[at].Number, "pages": urls}
	if at > 0 {
		out["prev_ch"], out["prev_ch_id"] = chs[at-1].Number, chs[at-1].ChapterID
	}
	if at+1 < len(chs) {
		out["next_ch"], out["next_ch_id"] = chs[at+1].Number, chs[at+1].ChapterID
	}
	return out, nil
}

// Window returns up to limit page URLs strictly after the stored position,
// walking forward through the number-ordered chapter list and skipping
// chapters with no stored pages. Only the store is consulted.
func (s *Service) Window(ctx context.Context, plugin, slug string, pos store.ReaderPosition, limit int) []string {
	chs, err := s.chapters(ctx, plugin, slug)
	if err != nil {
		return nil
	}
	at := locate(chs, pos.ChapterNumber, pos.ChapterID)
	if at < 0 {
		return nil
	}
	var out []string
	for i := at; i < len(chs) && len(out) < limit; i++ {
		urls := s.storedPages(ctx, plugin, slug, chs[i].ChapterID)
		begin := 0
		if i == at {
			begin = max(0, min(pos.Page, len(urls)))
		}
		if begin >= len(urls) {
			continue
		}
		end := min(len(urls), begin+limit-len(out))
		out = append(out, urls[begin:end]...)
	}
	return out
}

// PutPosition upserts the reading position, keeping stored values for absent
// fields. Recognised keys: chapter, chapter_id, page, mode, direction, zoom.
// ponytail: unknown keys are ignored; tighten when the API contract freezes.
func (s *Service) PutPosition(ctx context.Context, plugin, slug string, fields map[string]any) error {
	pos, err := s.Store.GetPosition(ctx, plugin, slug)
	if err != nil {
		pos = &store.ReaderPosition{Plugin: plugin, Slug: slug}
	}
	for k, v := range fields {
		switch k {
		case "chapter":
			if f, ok := num(v); ok {
				pos.ChapterNumber = f
			}
		case "chapter_id":
			if str, ok := v.(string); ok {
				pos.ChapterID = str
			}
		case "page":
			if f, ok := num(v); ok {
				pos.Page = int(f)
			}
		case "mode":
			if str, ok := v.(string); ok {
				pos.Mode = str
			}
		case "direction":
			if str, ok := v.(string); ok {
				pos.Direction = str
			}
		case "zoom":
			if str, ok := v.(string); ok {
				pos.Zoom = str
			}
		}
	}
	pos.UpdatedAt = time.Now().Unix()
	return s.Store.PutPosition(ctx, pos)
}

// Prefetch refreshes the page list of the next and then the previous chapter
// by calling Pages directly; empty ids are skipped.
func (s *Service) Prefetch(ctx context.Context, plugin, slug, next, prev string) error {
	var errs []error
	for _, id := range []string{next, prev} {
		if id == "" {
			continue
		}
		if _, err := s.Pages(ctx, plugin, slug, 0, id); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *Service) chapters(ctx context.Context, plugin, slug string) ([]store.Chapter, error) {
	var chs []store.Chapter
	if err := s.Store.DB().WithContext(ctx).Where("plugin = ? AND slug = ?", plugin, slug).Find(&chs).Error; err != nil {
		return nil, err
	}
	sort.Slice(chs, func(i, j int) bool {
		if chs[i].Number != chs[j].Number {
			return chs[i].Number < chs[j].Number
		}
		return chs[i].ChapterID < chs[j].ChapterID
	})
	return chs, nil
}

// locate returns the index of the exact chapterID entry, else the first
// entry with that number, else -1.
func locate(chs []store.Chapter, number float64, chapterID string) int {
	if chapterID != "" {
		for i, c := range chs {
			if c.ChapterID == chapterID {
				return i
			}
		}
	}
	for i, c := range chs {
		if c.Number == number {
			return i
		}
	}
	return -1
}

// pageList fetches the page list through the host and stores it, falling back
// to the last stored list when the host fails or returns nothing usable.
func (s *Service) pageList(ctx context.Context, plugin, slug string, ch store.Chapter) ([]string, error) {
	params, _ := json.Marshal(map[string]string{"id": ch.ChapterID})
	if out, err := s.Host.Call(ctx, plugin, "pages", params); err == nil {
		var urls []string
		if json.Unmarshal(out, &urls) == nil && len(urls) > 0 {
			pages := make([]store.ChapterPage, len(urls))
			for i, u := range urls {
				pages[i] = store.ChapterPage{URL: u}
			}
			return urls, s.Store.ReplacePages(ctx, plugin, slug, ch.ChapterID, pages)
		}
	}
	return s.storedPages(ctx, plugin, slug, ch.ChapterID), nil
}

func (s *Service) storedPages(ctx context.Context, plugin, slug, chapterID string) []string {
	var rows []store.ChapterPage
	_ = s.Store.DB().WithContext(ctx).Where("plugin = ? AND slug = ? AND chapter_id = ?", plugin, slug, chapterID).Order("idx").Find(&rows).Error
	urls := make([]string, len(rows))
	for i, r := range rows {
		urls[i] = r.URL
	}
	return urls
}

func num(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}
