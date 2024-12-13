//go:build pgx5 || !sqlite3

package migration

import (
	"github.com/jackc/pgx/v5"
)

const setupQuery = `
CREATE TABLE IF NOT EXISTS %s (
	version CHARACTER VARYING(50)       PRIMARY KEY,
	run_on  TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
)`

func sanitize(ident string) string {
	return pgx.Identifier([]string{ident}).Sanitize()
}
