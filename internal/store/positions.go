package store

import "context"

// ReaderPosition is one per-manga reading state.
type ReaderPosition struct {
	Plugin        string `gorm:"primaryKey"`
	Slug          string `gorm:"primaryKey"`
	ChapterID     string
	ChapterNumber float64
	Page          int
	Mode          string
	Direction     string
	Zoom          string
	UpdatedAt     int64
}

// TableName pins the frozen table name.
func (ReaderPosition) TableName() string { return "reader_positions" }

// PluginState is one host bookkeeping value for a plugin folder.
type PluginState struct {
	PluginID string `gorm:"primaryKey"`
	Key      string `gorm:"primaryKey"`
	Value    string
}

// TableName pins the frozen table name.
func (PluginState) TableName() string { return "plugin_state" }

// PutPosition upserts a reader position by composite key.
func (s *Store) PutPosition(ctx context.Context, p *ReaderPosition) error {
	return s.db.WithContext(ctx).Save(p).Error
}

// GetPosition returns one reader position.
func (s *Store) GetPosition(ctx context.Context, plugin, slug string) (*ReaderPosition, error) {
	var p ReaderPosition
	err := s.db.WithContext(ctx).Where("plugin = ? AND slug = ?", plugin, slug).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// PutPluginState upserts one host bookkeeping value.
func (s *Store) PutPluginState(ctx context.Context, pluginID, key, value string) error {
	return s.db.WithContext(ctx).Save(&PluginState{PluginID: pluginID, Key: key, Value: value}).Error
}

// GetPluginState returns one host bookkeeping value.
func (s *Store) GetPluginState(ctx context.Context, pluginID, key string) (string, error) {
	var p PluginState
	err := s.db.WithContext(ctx).Where("plugin_id = ? AND key = ?", pluginID, key).First(&p).Error
	return p.Value, err
}
