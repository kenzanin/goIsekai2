package store

import "context"

// TitleMeta holds enrichment payloads for one title.
type TitleMeta struct {
	Plugin    string `gorm:"primaryKey"`
	Slug      string `gorm:"primaryKey"`
	Fetched   string
	Chosen    string
	FetchedAt int64
}

// TableName pins the frozen table name.
func (TitleMeta) TableName() string { return "title_meta" }

// UpdateAttempt holds scheduler backoff state for one title.
type UpdateAttempt struct {
	Plugin            string `gorm:"primaryKey"`
	Slug              string `gorm:"primaryKey"`
	LastAttemptAt     int64
	Outcome           string
	Error             string
	ConsecutiveErrors int
	BackoffUntil      int64
}

// TableName pins the frozen table name.
func (UpdateAttempt) TableName() string { return "update_attempts" }

// PutTitleMeta upserts one enrichment payload by composite key.
func (s *Store) PutTitleMeta(ctx context.Context, m *TitleMeta) error {
	return s.db.WithContext(ctx).Save(m).Error
}

// GetTitleMeta returns one enrichment payload.
func (s *Store) GetTitleMeta(ctx context.Context, plugin, slug string) (*TitleMeta, error) {
	var m TitleMeta
	err := s.db.WithContext(ctx).Where("plugin = ? AND slug = ?", plugin, slug).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// PutUpdateAttempt upserts scheduler backoff state.
func (s *Store) PutUpdateAttempt(ctx context.Context, a *UpdateAttempt) error {
	return s.db.WithContext(ctx).Save(a).Error
}

// GetUpdateAttempt returns scheduler backoff state.
func (s *Store) GetUpdateAttempt(ctx context.Context, plugin, slug string) (*UpdateAttempt, error) {
	var a UpdateAttempt
	err := s.db.WithContext(ctx).Where("plugin = ? AND slug = ?", plugin, slug).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}
