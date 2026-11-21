package store

import (
	"context"

	"gorm.io/gorm"

	sqlb "goisekai2/internal/store/sql"
)

const (
	// ChapterRead marks a chapter as read.
	ChapterRead = 1 << 0
	// ChapterSkip marks a chapter as skipped.
	ChapterSkip = 1 << 1
)

// Chapter is one stable chapter identity; number is display/sort only.
type Chapter struct {
	Plugin    string `gorm:"primaryKey"`
	Slug      string `gorm:"primaryKey"`
	ChapterID string `gorm:"primaryKey"`
	Number    float64
	Title     string
	URL       string
	Flags     int
	UpdatedAt int64
}

// TableName pins the frozen table name.
func (Chapter) TableName() string { return "chapters" }

// ChapterPage is one source image URL in reading order.
type ChapterPage struct {
	Plugin    string `gorm:"primaryKey"`
	Slug      string `gorm:"primaryKey"`
	ChapterID string `gorm:"primaryKey"`
	Idx       int    `gorm:"primaryKey"`
	URL       string
}

// TableName pins the frozen table name.
func (ChapterPage) TableName() string { return "chapter_pages" }

// UpsertChapters replaces chapter metadata in place, preserving existing flag bits.
func (s *Store) UpsertChapters(ctx context.Context, plugin, slug string, chapters []Chapter) error {
	return s.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		for _, c := range chapters {
			var existing Chapter
			err := tx.WithContext(txCtx).Where("plugin = ? AND slug = ? AND chapter_id = ?", plugin, slug, c.ChapterID).First(&existing).Error
			if err == nil {
				c.Flags = existing.Flags
			}
			c.Plugin, c.Slug = plugin, slug
			if err := tx.WithContext(txCtx).Save(&c).Error; err != nil {
				return err
			}
		}
		return sqlb.RollupSaved(tx.WithContext(txCtx), plugin, slug)
	})
}

// ReplacePages replaces the page list for one chapter.
func (s *Store) ReplacePages(ctx context.Context, plugin, slug, chapterID string, pages []ChapterPage) error {
	return s.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		if err := tx.WithContext(txCtx).Where("plugin = ? AND slug = ? AND chapter_id = ?", plugin, slug, chapterID).Delete(&ChapterPage{}).Error; err != nil {
			return err
		}
		for i := range pages {
			pages[i].Plugin, pages[i].Slug, pages[i].ChapterID = plugin, slug, chapterID
			pages[i].Idx = i
			if err := tx.WithContext(txCtx).Create(&pages[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// MarkChapterRead sets the read bit while preserving skip and reserved bits.
func (s *Store) MarkChapterRead(ctx context.Context, plugin, slug, chapterID string) error {
	return s.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		if err := sqlb.SetChapterFlag(tx.WithContext(txCtx), plugin, slug, chapterID, ChapterRead, 0); err != nil {
			return err
		}
		return sqlb.RollupSaved(tx.WithContext(txCtx), plugin, slug)
	})
}

// HighestReadNumber returns the largest chapter number with the read bit set.
func (s *Store) HighestReadNumber(ctx context.Context, plugin, slug string) (float64, error) {
	var chapters []Chapter
	if err := s.db.WithContext(ctx).Where("plugin = ? AND slug = ? AND (flags & 1) <> 0", plugin, slug).Find(&chapters).Error; err != nil {
		return 0, err
	}
	var max float64
	for _, c := range chapters {
		if c.Number > max {
			max = c.Number
		}
	}
	return max, nil
}

// CarryReadToTarget marks target chapters up to bound as read, OR-ing flags.
func (s *Store) CarryReadToTarget(ctx context.Context, plugin, slug string, bound float64) error {
	return s.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		if err := sqlb.CarryReadBoundary(tx.WithContext(txCtx), plugin, slug, bound); err != nil {
			return err
		}
		return sqlb.RollupSaved(tx.WithContext(txCtx), plugin, slug)
	})
}

// TargetSaved reports whether a migration target already exists in the library.
func (s *Store) TargetSaved(ctx context.Context, plugin, slug string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&SavedManga{}).Where("plugin = ? AND slug = ?", plugin, slug).Count(&count).Error
	return count > 0, err
}
