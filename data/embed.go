package data

import "embed"

//go:embed repository/mysql/schema
//go:embed jobqueue/mysql/schema
var EFS embed.FS
