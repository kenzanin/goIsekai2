package sql

import "gorm.io/gorm"

// EnsureFTS creates the virtual table and reports whether FTS5 is available.

// CreateFTSTable is the baseline FTS5 index for local title search and
// BM25-ranked dedup candidates. It is applied after migrations so a missing
// FTS5 module only disables search/dedup, never boot.
const CreateFTSTable = `CREATE VIRTUAL TABLE IF NOT EXISTS library_fts USING fts5(
	plugin UNINDEXED, slug UNINDEXED, name, alt_titles, synopsis)`

// FTSTableName is the virtual table name.
const FTSTableName = "library_fts"

// EnsureFTS creates the FTS5 table; false means callers must fall back to LIKE.
func EnsureFTS(db *gorm.DB) bool {
	return db.Exec(CreateFTSTable).Error == nil
}

// InsertFTSRow replaces the indexed row for one (plugin, slug).
const InsertFTSRow = `INSERT INTO library_fts (plugin, slug, name, alt_titles, synopsis)
	VALUES (?, ?, ?, ?, ?)`

// DeleteFTSRow removes the indexed row for one (plugin, slug).
const DeleteFTSRow = `DELETE FROM library_fts WHERE plugin = ? AND slug = ?`

// SearchFTS ranks prefix matches by BM25 score.
const SearchFTS = `SELECT saved_manga.plugin, saved_manga.slug, saved_manga.name FROM library_fts
	JOIN saved_manga ON saved_manga.plugin = library_fts.plugin AND saved_manga.slug = library_fts.slug
	WHERE library_fts MATCH ? AND saved_manga.added_at > 0 ORDER BY bm25(library_fts) LIMIT 20`

// SearchLike is the fallback when FTS5 is unavailable.
const SearchLike = `SELECT plugin, slug, name FROM saved_manga
	WHERE (name LIKE ? OR alt_titles LIKE ? OR synopsis LIKE ?) AND added_at > 0 LIMIT 20`
