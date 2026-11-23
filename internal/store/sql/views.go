package sql

import "gorm.io/gorm"

// CreateCacheView exposes staged JSON payloads as queryable columns without
// duplicating their storage. It is a V2 convenience view, not a frozen table.
const DropCacheView = `DROP VIEW IF EXISTS api_cache_view`

// CreateCacheView exposes staged JSON payloads as queryable columns without
// duplicating their storage. It is a V2 convenience view, not a frozen table.
const CreateCacheView = `CREATE VIEW api_cache_view AS
SELECT
	plugin_id,
	query_hash,
	CASE WHEN json_valid(params_json) THEN json_extract(params_json, '$.query') END AS query_text,
	CASE WHEN json_valid(payload) THEN json_extract(payload, '$.ok') END AS ok,
	CASE WHEN json_valid(payload) THEN json_extract(payload, '$.data.total') END AS total,
	fetched_at,
	ttl_s,
	length(payload) AS bytes
FROM api_cache`

// EnsureCacheView recreates the JSON1 view used to inspect cached payloads.
func EnsureCacheView(db *gorm.DB) error {
	if err := db.Exec(DropCacheView).Error; err != nil {
		return err
	}
	return db.Exec(CreateCacheView).Error
}

// FindView reports whether a named view exists in sqlite_master.
func FindView(db *gorm.DB, name string) (string, error) {
	var found string
	err := db.Raw(`SELECT name FROM sqlite_master WHERE type = 'view' AND name = ?`, name).Scan(&found).Error
	return found, err
}
