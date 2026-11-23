package store

import (
	"context"

	sqlb "goisekai2/internal/store/sql"
)

// Stage is a process-lifetime staged search result held in a SQLite TEMP table.
type Stage struct {
	Plugin       string
	Slug         string
	DetailJSON   string
	ChaptersJSON string
	FetchedAt    int64
}

// EnsureStage creates the TEMP staging table on the active connection.
func (s *Store) EnsureStage(ctx context.Context) error {
	return sqlb.EnsureStageTable(s.db.WithContext(ctx))
}

// PutStage upserts a staged detail and chapters envelope pair.
func (s *Store) PutStage(ctx context.Context, plugin, slug, detail, chapters string, at int64) error {
	if err := s.EnsureStage(ctx); err != nil {
		return err
	}
	return sqlb.PutStage(s.db.WithContext(ctx), plugin, slug, detail, chapters, at)
}

// GetStage returns the staged envelope pair for one slug, or nil when absent.
func (s *Store) GetStage(ctx context.Context, plugin, slug string) (*Stage, error) {
	if err := s.EnsureStage(ctx); err != nil {
		return nil, err
	}
	row, err := sqlb.ReadStage(s.db.WithContext(ctx), plugin, slug)
	if err != nil || row == nil {
		return nil, err
	}
	return &Stage{
		Plugin: row.Plugin, Slug: row.Slug, DetailJSON: row.DetailJSON,
		ChaptersJSON: row.ChaptersJSON, FetchedAt: row.FetchedAt,
	}, nil
}

// DeleteStage removes a staged pair after it has been promoted.
func (s *Store) DeleteStage(ctx context.Context, plugin, slug string) error {
	if err := s.EnsureStage(ctx); err != nil {
		return err
	}
	return sqlb.RemoveStage(s.db.WithContext(ctx), plugin, slug)
}
