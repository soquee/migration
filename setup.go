package migration

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Setup creates a migrations table with the given name.
func Setup(ctx context.Context, migrationsTable string, tx *sql.Tx) error {
	tableIdent := pgx.Identifier{migrationsTable}
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	version CHARACTER VARYING(50)       PRIMARY KEY,
	run_on  TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
)`, tableIdent.Sanitize()))
	return err
}
