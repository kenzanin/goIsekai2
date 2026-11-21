package sql

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "m.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestMigrateFreshAndRejectNewer(t *testing.T) {
	db := openTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	version, err := CurrentVersion(db)
	if err != nil || version != len(Migrations) {
		t.Fatalf("version = %d, %v", version, err)
	}
	if err := db.Exec(`UPDATE schema_version SET version = ?`, 999).Error; err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); !errors.Is(err, ErrNewerDatabase) {
		t.Fatalf("newer db error = %v", err)
	}
}

func TestMigrateFailureKeepsPreviousVersion(t *testing.T) {
	db := openTestDB(t)
	original := Migrations
	t.Cleanup(func() { Migrations = original })
	Migrations = []Migration{
		original[0],
		{Version: 2, SQL: `CREATE TABLE ok_table (id INTEGER PRIMARY KEY)`},
		{Version: 3, SQL: `THIS IS NOT SQL`},
	}
	if err := db.Exec(original[0].SQL).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT OR IGNORE INTO schema_version (version) VALUES (0)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err == nil {
		t.Fatal("expected migration failure")
	}
	version, err := CurrentVersion(db)
	if err != nil || version != 2 {
		t.Fatalf("version after failure = %d, %v", version, err)
	}
}
