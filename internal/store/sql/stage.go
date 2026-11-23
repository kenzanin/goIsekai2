package sql

import "gorm.io/gorm"

// StageRow is one process-lifetime staged search result.
type StageRow struct {
	Plugin       string `gorm:"column:plugin_id"`
	Slug         string `gorm:"column:slug"`
	DetailJSON   string `gorm:"column:detail_json"`
	ChaptersJSON string `gorm:"column:chapters_json"`
	FetchedAt    int64  `gorm:"column:fetched_at"`
}

// EnsureStage creates the per-connection TEMP table. With MaxOpenConns(1) its
// lifetime is the daemon process; a restart starts empty.
const EnsureStage = `CREATE TEMP TABLE IF NOT EXISTS stage_detail (
	plugin_id TEXT NOT NULL,
	slug TEXT NOT NULL,
	detail_json TEXT NOT NULL DEFAULT '',
	chapters_json TEXT NOT NULL DEFAULT '',
	fetched_at INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (plugin_id, slug)
)`

// UpsertStage stores the raw detail and chapters envelopes for one slug.
const UpsertStage = `INSERT INTO stage_detail
	(plugin_id, slug, detail_json, chapters_json, fetched_at)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(plugin_id, slug) DO UPDATE SET
	detail_json = excluded.detail_json,
	chapters_json = excluded.chapters_json,
	fetched_at = excluded.fetched_at`

// GetStage reads one staged row.
const GetStage = `SELECT plugin_id, slug, detail_json, chapters_json, fetched_at
	FROM stage_detail WHERE plugin_id = ? AND slug = ?`

// DeleteStage removes a staged row after promotion.
const DeleteStage = `DELETE FROM stage_detail WHERE plugin_id = ? AND slug = ?`

// EnsureStageTable creates the TEMP staging table on the current connection.
func EnsureStageTable(db *gorm.DB) error { return db.Exec(EnsureStage).Error }

// PutStage upserts a staged search result.
func PutStage(db *gorm.DB, plugin, slug, detail, chapters string, at int64) error {
	return db.Exec(UpsertStage, plugin, slug, detail, chapters, at).Error
}

// ReadStage returns one staged row, or nil when nothing is staged.
func ReadStage(db *gorm.DB, plugin, slug string) (*StageRow, error) {
	var row StageRow
	err := db.Raw(GetStage, plugin, slug).Scan(&row).Error
	if err != nil || row.Slug == "" {
		return nil, err
	}
	return &row, nil
}

// RemoveStage deletes a staged row.
func RemoveStage(db *gorm.DB, plugin, slug string) error {
	return db.Exec(DeleteStage, plugin, slug).Error
}
