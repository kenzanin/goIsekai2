package sql

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ErrNewerDatabase reports a database written by a newer binary.
var ErrNewerDatabase = errors.New("database is newer than this binary")

// Migrate applies every migration newer than the recorded version. Each
// migration runs in its own transaction; a failure leaves the version intact.
func Migrate(db *gorm.DB) error {
	var current int
	var name string
	err := db.Raw(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'schema_version' LIMIT 1`).Scan(&name).Error
	if err == nil && name == "schema_version" {
		err = db.Raw(`SELECT version FROM schema_version LIMIT 1`).Scan(&current).Error
	}
	if err != nil || name != "schema_version" {
		if execErr := db.Exec(Migrations[0].SQL).Error; execErr != nil {
			return fmt.Errorf("create schema_version: %w", execErr)
		}
		if execErr := db.Exec(`INSERT OR IGNORE INTO schema_version (version) VALUES (0)`).Error; execErr != nil {
			return fmt.Errorf("seed schema_version: %w", execErr)
		}
		current = 0
	}
	if current > len(Migrations) {
		return fmt.Errorf("%w: database %d, binary %d", ErrNewerDatabase, current, len(Migrations))
	}
	for _, m := range Migrations {
		if m.Version <= current {
			continue
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(m.SQL).Error; err != nil {
				return fmt.Errorf("migration %d: %w", m.Version, err)
			}
			return tx.Exec(`UPDATE schema_version SET version = ? WHERE version < ?`, m.Version, m.Version).Error
		}); err != nil {
			return err
		}
	}
	return nil
}

// CurrentVersion returns the recorded schema version.
func CurrentVersion(db *gorm.DB) (int, error) {
	var version int
	err := db.Raw(`SELECT version FROM schema_version LIMIT 1`).Scan(&version).Error
	return version, err
}
