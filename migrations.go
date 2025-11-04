package database

import (
	"embed"
	stderr "errors"
	"github.com/Station-Manager/database/postgres"
	"github.com/Station-Manager/database/sqlite"
	"github.com/Station-Manager/errors"
	"github.com/golang-migrate/migrate/v4"
	pg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func (s *Service) doMigrations() error {
	const op errors.Op = "database.Service.doMigrations"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	// Hold the lock for the duration of the migration so that Open() or Close() cannot be called.
	//s.mu.Lock()
	//defer s.mu.Unlock()
	// The call should do this

	var fs embed.FS
	switch s.config.Driver {
	case PostgresDriver:
		fs = postgres.MigrationFiles
	case SqliteDriver:
		fs = sqlite.MigrationFiles
	default:
		return errors.New(op).Msg("Driver not supported.")
	}

	// Prepare iofs source from embedded files
	sourceDriver, err := iofs.New(fs, "migrations")
	if err != nil {
		return errors.New(op).Errorf("iofs.New: %w", err)
	}

	dbDriver, err := pg.WithInstance(s.handle, &pg.Config{})
	if err != nil {
		return errors.New(op).Errorf("pg.WithInstance: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return errors.New(op).Errorf("migrate.NewWithInstance: %w", err)
	}
	defer func(m *migrate.Migrate) {
		if closeErr, _ := m.Close(); closeErr != nil && err == nil {
			// Only return a close error if no other error occurred
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
