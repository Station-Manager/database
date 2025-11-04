package database

import (
	stderr "errors"
	"github.com/Station-Manager/database/postgres"
	"github.com/Station-Manager/database/sqlite"
	"github.com/Station-Manager/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source"
)

func (s *Service) doMigrations() error {
	const op errors.Op = "database.Service.doMigrations"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	var srcDriver source.Driver
	var dbDriver database.Driver
	var err error

	switch s.config.Driver {
	case PostgresDriver:
		srcDriver, dbDriver, err = postgres.GetMigrationDrivers(s.handle)
	case SqliteDriver:
		srcDriver, dbDriver, err = sqlite.GetMigrationDrivers(s.handle)
	default:
		return errors.New(op).Msg("Driver not supported.")
	}

	m, err := migrate.NewWithInstance("iofs", srcDriver, s.config.Driver, dbDriver)
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

	if stderr.Is(err, migrate.ErrNoChange) {
		return nil
	}

	return err
}
