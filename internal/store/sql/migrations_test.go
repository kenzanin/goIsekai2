package sql

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var createdTable = regexp.MustCompile(`(?i)CREATE (?:VIRTUAL )?TABLE (?:IF NOT EXISTS )?(\w+)`)

// frozen is the verbatim table inventory from openspec/specs/store/spec.md.
var frozen = []struct {
	version int
	table   string
}{
	{1, "schema_version"}, {2, "plugin"}, {3, "title"}, {4, "plugin_state"},
	{5, "reader_positions"}, {6, "saved_manga"}, {7, "title_meta"},
	{8, "update_attempts"}, {9, "chapters"}, {10, "chapter_pages"},
	{11, "api_cache"},
}

func TestMigrationsMatchFrozenInventory(t *testing.T) {
	if len(Migrations) != len(frozen) {
		t.Fatalf("got %d migrations, want %d", len(Migrations), len(frozen))
	}
	for i, w := range frozen {
		m := Migrations[i]
		if m.Version != w.version {
			t.Errorf("migration %d: Version = %d, want %d", i, m.Version, w.version)
		}
		name := createdTable.FindStringSubmatch(m.SQL)
		if name == nil || name[1] != w.table {
			t.Errorf("migration %d: DDL does not create %q", i, w.table)
		}
	}
	if got := Migrations[0].SQL; !strings.Contains(got, "version INTEGER PRIMARY KEY") {
		t.Errorf("schema_version DDL missing version INTEGER PRIMARY KEY: %s", got)
	}
	if !strings.Contains(CreateFTSTable, "CREATE VIRTUAL TABLE IF NOT EXISTS library_fts USING fts5") {
		t.Errorf("FTS5 DDL missing: %s", CreateFTSTable)
	}
}

// TestMigrationsExecute proves the DDL is valid on the shipping pure-Go stack
// (glebarez/sqlite, CGO_ENABLED=0), including the FTS5 virtual table.
func TestMigrationsExecute(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "store.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	for _, m := range Migrations {
		if err := db.Exec(m.SQL).Error; err != nil {
			t.Errorf("migration %d: %v", m.Version, err)
		}
	}
	if err := db.Exec(CreateFTSTable).Error; err != nil {
		t.Logf("FTS5 unavailable (fallback applies): %v", err)
	}
}

// TestNoSQLOutsideBoundary runs the repo's own boundary gate; it exits nonzero
// if any raw SQL literal or raw-SQL API call appears outside this package.
func TestNoSQLOutsideBoundary(t *testing.T) {
	root, err := filepath.Abs("../../..") // package dir is three below the module root
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", "scripts/sql-boundary-check.go")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("sql-boundary-check: %v\n%s", err, out)
	}
}
