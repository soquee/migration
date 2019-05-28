package migration

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

// Setup creates a migrations table with the given name.
func Setup(ctx context.Context, migrationsTable string, tx *sql.Tx) error {
	escTable := pq.QuoteIdentifier(migrationsTable)
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	version CHARACTER VARYING(50)       PRIMARY KEY,
	run_on  TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
)`, escTable))
	return err
}
