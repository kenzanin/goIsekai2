package sql

import (
	"gorm.io/gorm"
)

// SearchHit is one local-search result.
type SearchHit struct {
	Plugin string `gorm:"column:plugin"`
	Slug   string `gorm:"column:slug"`
	Name   string `gorm:"column:name"`
}

// IndexTitle replaces the FTS row for one title. Callers check FTS availability first.
func IndexTitle(db *gorm.DB, plugin, slug, name, alt, synopsis string) error {
	if err := db.Exec(DeleteFTSRow, plugin, slug).Error; err != nil {
		return err
	}
	return db.Exec(InsertFTSRow, plugin, slug, name, alt, synopsis).Error
}

// DeleteTitle removes the FTS row for one title.
func DeleteTitle(db *gorm.DB, plugin, slug string) error {
	return db.Exec(DeleteFTSRow, plugin, slug).Error
}

// Search ranks prefix matches by BM25, or falls back to LIKE when FTS is off.
func Search(db *gorm.DB, term string, ftsOK bool) ([]SearchHit, error) {
	var hits []SearchHit
	if !ftsOK {
		like := "%" + term + "%"
		err := db.Raw(SearchLike, like, like, like).Scan(&hits).Error
		return hits, err
	}
	err := db.Raw(SearchFTS, term+"*").Scan(&hits).Error
	return hits, err
}

// RollupSaved recomputes the saved-stats fields from the store's own rows.
func RollupSaved(db *gorm.DB, plugin, slug string) error {
	const stmt = `UPDATE saved_manga SET
		total_ch = (SELECT COUNT(*) FROM chapters WHERE plugin = ? AND slug = ?),
		read_ch = (SELECT COUNT(*) FROM chapters WHERE plugin = ? AND slug = ? AND (flags & 1) <> 0),
		status_flags = CASE WHEN (SELECT COUNT(*) FROM chapters WHERE plugin = ? AND slug = ?) > 0
			AND (SELECT COUNT(*) FROM chapters WHERE plugin = ? AND slug = ? AND (flags & 1) <> 0)
				= (SELECT COUNT(*) FROM chapters WHERE plugin = ? AND slug = ?)
			THEN status_flags | 2 ELSE status_flags & ~2 END
		WHERE plugin = ? AND slug = ?`
	return db.Exec(stmt, plugin, slug, plugin, slug, plugin, slug, plugin, slug, plugin, slug, plugin, slug).Error
}

// SetChapterFlag applies a bit operation; it never rewrites flags wholesale.
func SetChapterFlag(db *gorm.DB, plugin, slug, chapterID string, set, clear int) error {
	const stmt = `UPDATE chapters SET flags = (flags | ?) & ~? WHERE plugin = ? AND slug = ? AND chapter_id = ?`
	return db.Exec(stmt, set, clear, plugin, slug, chapterID).Error
}

// CarryReadBoundary marks target chapters at or below bound as read, OR-ing
// existing flags instead of overwriting them.
func CarryReadBoundary(db *gorm.DB, plugin, slug string, bound float64) error {
	const stmt = `UPDATE chapters SET flags = flags | 1 WHERE plugin = ? AND slug = ? AND number < ? + 1`
	return db.Exec(stmt, plugin, slug, bound).Error
}

// CacheHit returns a payload only while fetched_at + ttl_s is in the future.
func CacheHit(db *gorm.DB, pluginID, hash string, now int64) (string, bool) {
	var payload string
	err := db.Raw(`SELECT payload FROM api_cache WHERE plugin_id = ? AND query_hash = ?
		AND fetched_at + ttl_s > ? LIMIT 1`, pluginID, hash, now).Scan(&payload).Error
	return payload, err == nil && payload != ""
}

// InvalidateCache removes cached rows for one query or one plugin.
func InvalidateCache(db *gorm.DB, pluginID, hash string, all bool) error {
	if all {
		return db.Exec(`DELETE FROM api_cache WHERE plugin_id = ?`, pluginID).Error
	}
	return db.Exec(`DELETE FROM api_cache WHERE plugin_id = ? AND query_hash = ?`, pluginID, hash).Error
}

// BackupTo writes a consistent snapshot via VACUUM INTO.
func BackupTo(db *gorm.DB, dest string) error {
	return db.Exec(`VACUUM INTO ?`, dest).Error
}
