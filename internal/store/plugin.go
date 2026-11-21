package store

import (
	"context"

	"gorm.io/gorm"
)

// Title is the legacy title row keyed to a plugin.id foreign key.
type Title struct {
	ID       int64 `gorm:"primaryKey"`
	PluginID int64
	Name     string
}

// TableName pins the frozen table name.
func (Title) TableName() string { return "title" }

// PluginRow is one registered plugin row from the frozen schema.
type PluginRow struct {
	ID           int64 `gorm:"primaryKey"`
	Name         string
	PluginID     string
	Version      string
	Capabilities string
	Enabled      int
	Runtime      string
}

// TableName pins the frozen table name.
func (PluginRow) TableName() string { return "plugin" }

// UpsertPlugin inserts or updates one plugin row by folder id.
func (s *Store) UpsertPlugin(ctx context.Context, p *PluginRow) error {
	var existing PluginRow
	err := s.db.WithContext(ctx).Where("plugin_id = ?", p.PluginID).First(&existing).Error
	if err == nil {
		p.ID = existing.ID
	}
	return s.db.WithContext(ctx).Save(p).Error
}

// GetPlugin returns one plugin row by folder id.
func (s *Store) GetPlugin(ctx context.Context, folder string) (*PluginRow, error) {
	var p PluginRow
	err := s.db.WithContext(ctx).Where("plugin_id = ?", folder).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// AllPlugins returns every registered plugin row ordered by folder id.
func (s *Store) AllPlugins(ctx context.Context) ([]PluginRow, error) {
	var rows []PluginRow
	err := s.db.WithContext(ctx).Order("plugin_id").Find(&rows).Error
	return rows, err
}

// SetPluginEnabled updates the parity boolean used by the baseline API.
func (s *Store) SetPluginEnabled(ctx context.Context, folder string, enabled bool) error {
	value := 0
	if enabled {
		value = 1
	}
	return s.db.WithContext(ctx).Model(&PluginRow{}).Where("plugin_id = ?", folder).Update("enabled", value).Error
}

// CountSavedByPlugin returns the saved-title count for one plugin folder.
func (s *Store) CountSavedByPlugin(ctx context.Context, folder string) (int, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&SavedManga{}).Where("plugin = ?", folder).Count(&count).Error
	return int(count), err
}

// DeletePluginRows purges every data row of a plugin in one transaction.
func (s *Store) DeletePluginRows(ctx context.Context, folder string) error {
	row, err := s.GetPlugin(ctx, folder)
	if err != nil {
		return err
	}
	return s.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		if err := tx.WithContext(txCtx).Where("plugin_id = ?", row.ID).Delete(&Title{}).Error; err != nil {
			return err
		}
		typed := []struct {
			model  any
			column string
		}{
			{&PluginState{}, "plugin_id"},
			{&ReaderPosition{}, "plugin"},
			{&SavedManga{}, "plugin"},
			{&TitleMeta{}, "plugin"},
			{&UpdateAttempt{}, "plugin"},
			{&ApiCache{}, "plugin_id"},
		}
		for _, item := range typed {
			if err := tx.WithContext(txCtx).Where(item.column+" = ?", folder).Delete(item.model).Error; err != nil {
				return err
			}
		}
		return tx.WithContext(txCtx).Where("plugin_id = ?", folder).Delete(&PluginRow{}).Error
	})
}
