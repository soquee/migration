package migration

import (
	"context"
	"database/sql"
	"fmt"
)

// Setup creates a migrations table with the given name.
func Setup(ctx context.Context, migrationsTable string, tx *sql.Tx) error {
	migrationsTable = sanitize(migrationsTable)
	_, err := tx.ExecContext(ctx, fmt.Sprintf(setupQuery, migrationsTable))
	return err
}
