package migration

import (
	"context"
	"database/sql"
	"fmt"
)

// Setup creates a migrations table with the given name.
func Setup(ctx context.Context, migrationsTable string, tx *sql.Tx) error {
	migrationsTable = sanitize(migrationsTable)
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	version TEXT PRIMARY KEY NOT NULL,
	run_on  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`, migrationsTable))
	return err
}
