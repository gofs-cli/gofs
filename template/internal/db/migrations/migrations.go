package migrations

import (
	"embed"
)

//go:embed *.sql
var MigrationDir embed.FS
