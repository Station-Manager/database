package postgres

import (
	"database/sql"
	"embed"
	stderr "errors"
	"github.com/Station-Manager/errors"
	"github.com/golang-migrate/migrate/v4"
	pg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var schema embed.FS

func Migrations(handle *sql.DB) error {
	const op errors.Op = "database.postgres.Migrations"
	if handle == nil {
		return errors.New(op).Msg("database handle is nil")
	}

	// Prepare iofs source from embedded files
	sourceDriver, err := iofs.New(schema, "migrations")
	if err != nil {
		return errors.New(op).Errorf("iofs.New: %w", err)
	}

	dbDriver, err := pg.WithInstance(handle, &pg.Config{})
	if err != nil {
		return errors.New(op).Errorf("pg.WithInstance: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return errors.New(op).Errorf("migrate.NewWithInstance: %w", err)
	}
	defer func(m *migrate.Migrate) {
		if closeErr, _ := m.Close(); closeErr != nil && err == nil {
			// Only return close error if no other error occurred
			err = errors.New(op).Errorf("m.Close: %w", closeErr)
		}
	}(m)

	err = m.Up()
	if err != nil && !stderr.Is(err, migrate.ErrNoChange) {
		return errors.New(op).Errorf("m.Up: %w", err)
	}

	// Returns err (which could be set by deferred close)
	return err
}
