package postgres

import (
	"database/sql"
	"embed"
)

//go:embed migrations/*.sql
var schema embed.FS

func Migrations(handle *sql.DB) error {
	return nil
}
