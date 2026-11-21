package store

import (
	"context"
	"fmt"
	"time"

	sqlb "goisekai2/internal/store/sql"
)

// BackupTo writes a consistent snapshot with the baseline filename pattern.
func (s *Store) BackupTo(ctx context.Context, dataDir string) (string, error) {
	name := fmt.Sprintf("library-%s.db", time.Now().Format("20060102-150405"))
	dest := dataDir + "/" + name
	if err := sqlb.BackupTo(s.db.WithContext(ctx), dest); err != nil {
		return "", err
	}
	return dest, nil
}
