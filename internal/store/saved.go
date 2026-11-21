package store

import (
	"context"
	"strings"

	"gorm.io/gorm"

	sqlb "goisekai2/internal/store/sql"
)

const (
	// StatusNew marks a newly added title.
	StatusNew = 1 << 0
	// StatusDone marks a fully read title.
	StatusDone = 1 << 1
	// StatusOngoing marks an ongoing source status.
	StatusOngoing = 1 << 2
	// StatusHiatus marks a paused source status.
	StatusHiatus = 1 << 3
	// StatusCanceled marks a canceled source status.
	StatusCanceled = 1 << 4
	// StatusCompleted marks a completed source status.
	StatusCompleted = 1 << 5
	// StatusDim marks a dimmed title.
	StatusDim = 1 << 6
)

// SavedManga is the host's persisted view of one source title.
type SavedManga struct {
	Plugin          string `gorm:"primaryKey"`
	Slug            string `gorm:"primaryKey"`
	Name            string
	Cover           string
	AddedAt         int64
	Status          string
	TotalCh         int
	ReadCh          int
	LastReadAt      int64
	UpdatedAt       int64
	Synopsis        string
	AltTitles       string
	DupGroup        int
	StatusFlags     int
	Authors         string
	Artists         string
	Tags            string
	SynopsisDisplay string
	Pinned          int
}

// TableName pins the frozen table name.
func (SavedManga) TableName() string { return "saved_manga" }

// FoldStatus maps a source status to its category bits while leaving the
// verbatim string untouched. Unknown strings keep the previous category.
func FoldStatus(status string, prev int) int {
	category := StatusOngoing | StatusHiatus | StatusCanceled | StatusCompleted
	base := prev &^ category
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "ongoing", "ongoin", " ongoing ":
		return base | StatusOngoing
	case "hiatus", "onhaiatus", "on-haiatus":
		return base | StatusHiatus
	case "canceled", "cancelled":
		return base | StatusCanceled
	case "complete", "completed", "done":
		return base | StatusCompleted
	default:
		return prev
	}
}

// UpsertSaved inserts or updates one saved title without touching its flags.
func (s *Store) UpsertSaved(ctx context.Context, m *SavedManga) error {
	return s.db.WithContext(ctx).Save(m).Error
}

// GetSaved returns one saved title by composite key.
func (s *Store) GetSaved(ctx context.Context, plugin, slug string) (*SavedManga, error) {
	var m SavedManga
	err := s.db.WithContext(ctx).Where("plugin = ? AND slug = ?", plugin, slug).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// SetSavedStatus stores the verbatim status and folds its category bits.
func (s *Store) SetSavedStatus(ctx context.Context, plugin, slug, status string) error {
	m, err := s.GetSaved(ctx, plugin, slug)
	if err != nil {
		return err
	}
	m.Status = status
	m.StatusFlags = FoldStatus(status, m.StatusFlags)
	return s.db.WithContext(ctx).Save(m).Error
}

// Rollup recomputes total_ch, read_ch and the is_done bit from store rows.
func Rollup(ctx context.Context, tx *gorm.DB, plugin, slug string) error {
	return sqlb.RollupSaved(tx.WithContext(ctx), plugin, slug)
}
