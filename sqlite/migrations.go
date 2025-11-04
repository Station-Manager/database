package sqlite

import (
	"database/sql"
	"embed"
	"github.com/Station-Manager/errors"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func GetMigrationDrivers(handle *sql.DB) (source.Driver, database.Driver, error) {
	const op errors.Op = "database.sqlite.sourceDriver"
	srcDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return nil, nil, errors.New(op).Errorf("iofs.New: %w", err)
	}
	dbDriver, err := sqlite3.WithInstance(handle, &sqlite3.Config{})
	if err != nil {
		return nil, nil, errors.New(op).Errorf("sqlite3.WithInstance: %w", err)
	}
	return srcDriver, dbDriver, nil
}
