package store

import (
	"context"
	"time"

	sqlb "goisekai2/internal/store/sql"
)

// UpdateCandidate is one scheduler candidate identified by its composite key.
type UpdateCandidate = sqlb.UpdateCandidate

// UpdateCandidates returns barely-read titles eligible for a scheduled update.
func (s *Store) UpdateCandidates(ctx context.Context, now int64, limit int) ([]UpdateCandidate, error) {
	return sqlb.UpdateCandidates(s.db.WithContext(ctx), now, limit)
}

// PruneExpiredCache removes API cache rows whose TTL has elapsed.
func (s *Store) PruneExpiredCache(ctx context.Context, now int64) error {
	return sqlb.PruneExpired(s.db.WithContext(ctx), now)
}

// BackupInto writes a consistent database snapshot to dst.
func (s *Store) BackupInto(ctx context.Context, dst string) error {
	return sqlb.BackupInto(s.db.WithContext(ctx), dst)
}

// BackupWatermark returns the last successful backup time, or zero.
func (s *Store) BackupWatermark(ctx context.Context) (time.Time, error) {
	return sqlb.BackupWatermark(s.db.WithContext(ctx))
}
