package migrations

import "embed"

// FS contains SQL migrations used by the default CLI migration path.
//
//go:embed *.sql
var FS embed.FS
