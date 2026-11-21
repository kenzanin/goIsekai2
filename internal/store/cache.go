package store

import (
	"context"

	sqlb "goisekai2/internal/store/sql"
)

// ApiCache is one cached plugin response.
type ApiCache struct {
	PluginID   string `gorm:"primaryKey"`
	QueryHash  string `gorm:"primaryKey"`
	ParamsJSON string
	FetchedAt  int64
	TTL        int64 `gorm:"column:ttl_s"`
	Payload    string
}

// TableName pins the frozen table name.
func (ApiCache) TableName() string { return "api_cache" }

// GetCached returns a payload only while its TTL has not expired.
func (s *Store) GetCached(ctx context.Context, pluginID, hash string, now int64) (string, bool) {
	return sqlb.CacheHit(s.db.WithContext(ctx), pluginID, hash, now)
}

// PutCached writes a payload through to the cache.
func (s *Store) PutCached(ctx context.Context, c *ApiCache) error {
	return s.db.WithContext(ctx).Save(c).Error
}

// InvalidateCached deletes one query or all queries of a plugin.
func (s *Store) InvalidateCached(ctx context.Context, pluginID, hash string, all bool) error {
	return sqlb.InvalidateCache(s.db.WithContext(ctx), pluginID, hash, all)
}
