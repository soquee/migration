package migration

import (
	"strings"
)

var replacer = strings.NewReplacer(
	string([]byte{0}), "",
	`"`, `""`,
)

func sanitize(ident string) string {
	return `"` + replacer.Replace(ident) + `"`
}
