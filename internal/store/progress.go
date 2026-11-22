package store

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

const readProgressPrefix = "host/read/"

func readProgressKey(plugin, slug, chapterID string) string {
	return readProgressPrefix + plugin + "/" + slug + "/" + chapterID
}

// PutReadProgress records the highest page reached for one chapter. The value
// lives in plugin_state under the reserved host id so the frozen schema needs no
// per-chapter progress table.
func (s *Store) PutReadProgress(ctx context.Context, plugin, slug, chapterID string, page int) error {
	return s.db.WithContext(ctx).Save(&PluginState{
		PluginID: "host",
		Key:      readProgressKey(plugin, slug, chapterID),
		Value:    strconv.Itoa(page),
	}).Error
}

// ReadProgress returns the stored page count for one chapter, or zero when no
// progress has been recorded.
func (s *Store) ReadProgress(ctx context.Context, plugin, slug, chapterID string) (int, error) {
	var row PluginState
	err := s.db.WithContext(ctx).
		Where("plugin_id = ? AND key = ?", "host", readProgressKey(plugin, slug, chapterID)).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	page, err := strconv.Atoi(row.Value)
	if err != nil {
		return 0, nil
	}
	return page, nil
}

// ReadProgresses returns stored progress for the given chapter ids. Ids without
// a record are omitted from the result.
func (s *Store) ReadProgresses(ctx context.Context, plugin, slug string, chapterIDs []string) (map[string]int, error) {
	out := map[string]int{}
	if len(chapterIDs) == 0 {
		return out, nil
	}
	keys := make([]string, 0, len(chapterIDs))
	for _, id := range chapterIDs {
		keys = append(keys, readProgressKey(plugin, slug, id))
	}
	var rows []PluginState
	if err := s.db.WithContext(ctx).
		Where("plugin_id = ? AND key IN ?", "host", keys).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		page, err := strconv.Atoi(row.Value)
		if err != nil {
			continue
		}
		id := strings.TrimPrefix(row.Key, readProgressPrefix+plugin+"/"+slug+"/")
		out[id] = page
	}
	return out, nil
}
