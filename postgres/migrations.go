package postgres

import (
	"database/sql"
	"embed"
	"github.com/Station-Manager/errors"
	"github.com/golang-migrate/migrate/v4/database"
	pg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func GetMigrationDrivers(handle *sql.DB) (source.Driver, database.Driver, error) {
	const op errors.Op = "database.postgres.sourceDriver"
	srcDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return nil, nil, errors.New(op).Errorf("iofs.New: %w", err)
	}
	dbDriver, err := pg.WithInstance(handle, &pg.Config{})
	if err != nil {
		return nil, nil, errors.New(op).Errorf("pg.WithInstance: %w", err)
	}
	return srcDriver, dbDriver, nil
}
