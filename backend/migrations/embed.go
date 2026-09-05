// Package migrations embeds the SQL schema files applied at boot.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
