//go:build sqlite3

package migration

import (
	"strings"
)

const setupQuery = `
CREATE TABLE IF NOT EXISTS %s (
	version TEXT PRIMARY KEY NOT NULL,
	run_on  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`

var replacer = strings.NewReplacer(
	string([]byte{0}), "",
	`"`, `""`,
)

func sanitize(ident string) string {
	return `"` + replacer.Replace(ident) + `"`
}
