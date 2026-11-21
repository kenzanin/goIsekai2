// Package sql is the single boundary where goIsekai2 keeps raw SQL.
package sql

// Migration is one forward-only schema step.
type Migration struct {
	Version int
	SQL     string
}

// Migrations is the frozen schema in application order. Adding a column is a
// new numbered step; existing steps are never edited.
var Migrations = []Migration{
	{1, `CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY NOT NULL)`},
	{2, `CREATE TABLE IF NOT EXISTS plugin (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		plugin_id TEXT NOT NULL DEFAULT '',
		version TEXT NOT NULL DEFAULT '',
		capabilities TEXT NOT NULL DEFAULT '',
		enabled INTEGER NOT NULL DEFAULT 1,
		runtime TEXT NOT NULL DEFAULT '')`},
	{3, `CREATE TABLE IF NOT EXISTS title (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		plugin_id INTEGER NOT NULL,
		name TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (plugin_id) REFERENCES plugin(id))`},
	{4, `CREATE TABLE IF NOT EXISTS plugin_state (
		plugin_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL DEFAULT '',
		PRIMARY KEY (plugin_id, key))`},
	{5, `CREATE TABLE IF NOT EXISTS reader_positions (
		plugin TEXT NOT NULL,
		slug TEXT NOT NULL,
		chapter_id TEXT NOT NULL DEFAULT '',
		chapter_number REAL NOT NULL DEFAULT 0,
		page INTEGER NOT NULL DEFAULT 0,
		mode TEXT NOT NULL DEFAULT '',
		direction TEXT NOT NULL DEFAULT '',
		zoom TEXT NOT NULL DEFAULT '',
		updated_at INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (plugin, slug))`},
	{6, `CREATE TABLE IF NOT EXISTS saved_manga (
		plugin TEXT NOT NULL,
		slug TEXT NOT NULL,
		name TEXT NOT NULL DEFAULT '',
		cover TEXT NOT NULL DEFAULT '',
		added_at INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'ongoing',
		total_ch INTEGER NOT NULL DEFAULT 0,
		read_ch INTEGER NOT NULL DEFAULT 0,
		last_read_at INTEGER NOT NULL DEFAULT 0,
		updated_at INTEGER NOT NULL DEFAULT 0,
		synopsis TEXT NOT NULL DEFAULT '',
		alt_titles TEXT NOT NULL DEFAULT '',
		dup_group INTEGER NOT NULL DEFAULT 0,
		status_flags INTEGER NOT NULL DEFAULT 0,
		authors TEXT NOT NULL DEFAULT '',
		artists TEXT NOT NULL DEFAULT '',
		tags TEXT NOT NULL DEFAULT '',
		synopsis_display TEXT NOT NULL DEFAULT '',
		pinned INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (plugin, slug))`},
	{7, `CREATE TABLE IF NOT EXISTS title_meta (
		plugin TEXT NOT NULL,
		slug TEXT NOT NULL,
		fetched TEXT NOT NULL DEFAULT '',
		chosen TEXT NOT NULL DEFAULT '',
		fetched_at INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (plugin, slug))`},
	{8, `CREATE TABLE IF NOT EXISTS update_attempts (
		plugin TEXT NOT NULL,
		slug TEXT NOT NULL,
		last_attempt_at INTEGER NOT NULL DEFAULT 0,
		outcome TEXT NOT NULL DEFAULT '',
		error TEXT NOT NULL DEFAULT '',
		consecutive_errors INTEGER NOT NULL DEFAULT 0,
		backoff_until INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (plugin, slug))`},
	{9, `CREATE TABLE IF NOT EXISTS chapters (
		plugin TEXT NOT NULL,
		slug TEXT NOT NULL,
		chapter_id TEXT NOT NULL,
		number REAL NOT NULL DEFAULT 0,
		title TEXT NOT NULL DEFAULT '',
		url TEXT NOT NULL DEFAULT '',
		flags INTEGER NOT NULL DEFAULT 0,
		updated_at INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (plugin, slug, chapter_id))`},
	{10, `CREATE TABLE IF NOT EXISTS chapter_pages (
		plugin TEXT NOT NULL,
		slug TEXT NOT NULL,
		chapter_id TEXT NOT NULL,
		idx INTEGER NOT NULL,
		url TEXT NOT NULL DEFAULT '',
		PRIMARY KEY (plugin, slug, chapter_id, idx))`},
	{11, `CREATE TABLE IF NOT EXISTS api_cache (
		plugin_id TEXT NOT NULL,
		query_hash TEXT NOT NULL,
		params_json TEXT NOT NULL DEFAULT '',
		fetched_at INTEGER NOT NULL DEFAULT 0,
		ttl_s INTEGER NOT NULL DEFAULT 0,
		payload TEXT NOT NULL DEFAULT '',
		PRIMARY KEY (plugin_id, query_hash))`},
}

// FrozenTables lists every non-virtual table the schema must contain.
var FrozenTables = []string{
	"schema_version", "plugin", "title", "plugin_state", "reader_positions",
	"saved_manga", "title_meta", "update_attempts", "chapters", "chapter_pages", "api_cache",
}
