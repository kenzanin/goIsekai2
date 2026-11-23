package sql

import "gorm.io/gorm"

// CreateCacheView exposes staged JSON payloads as queryable columns without
// duplicating their storage. It is a V2 convenience view, not a frozen table.
const CreateCacheView = `CREATE VIEW IF NOT EXISTS api_cache_view AS
SELECT
	plugin_id,
	query_hash,
	json_extract(params_json, '$.query') AS query_text,
	json_extract(payload, '$.ok') AS ok,
	json_extract(payload, '$.data.total') AS total,
	fetched_at,
	ttl_s,
	length(payload) AS bytes
FROM api_cache`

// EnsureCacheView creates the JSON1 view used to inspect cached payloads.
func EnsureCacheView(db *gorm.DB) error { return db.Exec(CreateCacheView).Error }

// FindView reports whether a named view exists in sqlite_master.
func FindView(db *gorm.DB, name string) (string, error) {
	var found string
	err := db.Raw(`SELECT name FROM sqlite_master WHERE type = 'view' AND name = ?`, name).Scan(&found).Error
	return found, err
}
