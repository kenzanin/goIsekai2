package sql

import (
	"time"

	"gorm.io/gorm"
)

// UpdateCandidate is one scheduler candidate row.
type UpdateCandidate struct {
	Plugin string
	Slug   string
}

// UpdateCandidates selects barely-read titles whose plugin is enabled and whose
// backoff has elapsed, never-attempted first then oldest attempt.
func UpdateCandidates(db *gorm.DB, now int64, limit int) ([]UpdateCandidate, error) {
	var out []UpdateCandidate
	err := db.Table("saved_manga").
		Select("saved_manga.plugin, saved_manga.slug").
		Joins("JOIN plugin ON plugin.plugin_id = saved_manga.plugin AND plugin.enabled = 1").
		Where("saved_manga.read_ch <= 1").
		Where(`NOT EXISTS (
			SELECT 1 FROM update_attempts
			WHERE update_attempts.plugin = saved_manga.plugin
			  AND update_attempts.slug = saved_manga.slug
			  AND update_attempts.backoff_until > ?
		)`, now).
		Order(`CASE WHEN EXISTS (
			SELECT 1 FROM update_attempts
			WHERE update_attempts.plugin = saved_manga.plugin
			  AND update_attempts.slug = saved_manga.slug
		) THEN 1 ELSE 0 END ASC, (
			SELECT last_attempt_at FROM update_attempts
			WHERE update_attempts.plugin = saved_manga.plugin
			  AND update_attempts.slug = saved_manga.slug
		) ASC`).
		Limit(limit).Scan(&out).Error
	return out, err
}

// PruneExpired deletes API cache rows whose TTL has elapsed.
func PruneExpired(db *gorm.DB, now int64) error {
	return db.Exec("DELETE FROM api_cache WHERE fetched_at + ttl_s < ?", now).Error
}

// BackupInto writes a consistent snapshot with SQLite's VACUUM INTO.
func BackupInto(db *gorm.DB, dst string) error {
	return db.Exec("VACUUM INTO ?", dst).Error
}

// BackupWatermark returns the stored backup watermark.
func BackupWatermark(db *gorm.DB) (time.Time, error) {
	var raw string
	err := db.Raw("SELECT value FROM plugin_state WHERE plugin_id = ? AND key = ?", "backup", "last_run").Scan(&raw).Error
	if err != nil || raw == "" {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, raw)
}
