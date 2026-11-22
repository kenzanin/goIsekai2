// Package library owns saved-manga mutations and their domain events.
package library

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"goisekai2/internal/events"
	"goisekai2/internal/store"
)

var errMissingKey = errors.New("library: plugin and slug are required")

// Service owns library mutations over the typed store.
type Service struct {
	store *store.Store
	bus   *events.Bus
}

// New creates a library service.
func New(st *store.Store, bus *events.Bus) *Service {
	return &Service{store: st, bus: bus}
}

// Save upserts one saved manga without degrading a good stored name.
func (s *Service) Save(ctx context.Context, plugin, slug, name, cover string) error {
	if plugin == "" || slug == "" {
		return errMissingKey
	}
	var err error
	if err = s.store.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		var existing store.SavedManga
		found := tx.WithContext(txCtx).Where("plugin = ? AND slug = ?", plugin, slug).First(&existing).Error == nil
		row := existing
		if !found {
			row = store.SavedManga{Plugin: plugin, Slug: slug}
		}
		if name != "" && name != slug {
			row.Name = name
		}
		if cover != "" {
			row.Cover = cover
		}
		row.Plugin, row.Slug, row.AddedAt = plugin, slug, time.Now().Unix()
		if found && existing.AddedAt > 0 {
			row.AddedAt = existing.AddedAt
		}
		return tx.WithContext(txCtx).Save(&row).Error
	}); err != nil {
		return err
	}
	if s.bus != nil {
		s.bus.Publish(events.LibrarySaved, events.LibraryPayload{Plugin: plugin, Slug: slug})
	}
	return nil
}

// Remove deletes one saved manga and publishes after the commit.
func (s *Service) Remove(ctx context.Context, plugin, slug string) error {
	if err := s.store.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		var row store.SavedManga
		if err := tx.WithContext(txCtx).Where("plugin = ? AND slug = ?", plugin, slug).First(&row).Error; err != nil {
			return err
		}
		row.AddedAt = 0
		return tx.WithContext(txCtx).Save(&row).Error
	}); err != nil {
		return err
	}
	if s.bus != nil {
		s.bus.Publish(events.LibraryRemoved, events.LibraryPayload{Plugin: plugin, Slug: slug})
	}
	return nil
}

// Get returns one saved manga by composite key.
func (s *Service) Get(ctx context.Context, plugin, slug string) (*store.SavedManga, error) {
	var row store.SavedManga
	err := s.store.DB().WithContext(ctx).Where("plugin = ? AND slug = ?", plugin, slug).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// List returns every saved entry ordered by added_at descending.
func (s *Service) List(ctx context.Context) ([]store.SavedManga, error) {
	var rows []store.SavedManga
	err := s.store.DB().WithContext(ctx).Where("added_at > 0").Order("added_at DESC").Find(&rows).Error
	return rows, err
}

// SetFlag applies one bit operation to a saved manga's status flags.
func (s *Service) SetFlag(ctx context.Context, plugin, slug string, set, clear int) error {
	return s.store.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		var row store.SavedManga
		if err := tx.WithContext(txCtx).Where("plugin = ? AND slug = ?", plugin, slug).First(&row).Error; err != nil {
			return err
		}
		row.StatusFlags = (row.StatusFlags | set) &^ clear
		return tx.WithContext(txCtx).Save(&row).Error
	})
}

// ToggleDim flips only bit6.
func (s *Service) ToggleDim(ctx context.Context, plugin, slug string) error {
	return s.toggle(ctx, plugin, slug, store.StatusDim)
}

// TogglePin flips only the pinned column.
func (s *Service) TogglePin(ctx context.Context, plugin, slug string) error {
	return s.store.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		var row store.SavedManga
		if err := tx.WithContext(txCtx).Where("plugin = ? AND slug = ?", plugin, slug).First(&row).Error; err != nil {
			return err
		}
		if row.Pinned == 0 {
			row.Pinned = 1
		} else {
			row.Pinned = 0
		}
		return tx.WithContext(txCtx).Save(&row).Error
	})
}

func (s *Service) toggle(ctx context.Context, plugin, slug string, bit int) error {
	return s.store.Tx(ctx, func(txCtx context.Context, tx *gorm.DB) error {
		var row store.SavedManga
		if err := tx.WithContext(txCtx).Where("plugin = ? AND slug = ?", plugin, slug).First(&row).Error; err != nil {
			return err
		}
		if row.StatusFlags&bit != 0 {
			row.StatusFlags &^= bit
		} else {
			row.StatusFlags |= bit
		}
		return tx.WithContext(txCtx).Save(&row).Error
	})
}
