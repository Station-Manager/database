package database

import (
	"github.com/Station-Manager/adapters"
	pgmodels "github.com/Station-Manager/database/postgres/models"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func (s *Service) InsertLogbook(logbook types.Logbook) (types.Logbook, error) {
	const op errors.Op = "database.Service.InsertLogbook"
	if err := checkService(op, s); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteInsertLogbook(logbook)
	case PostgresDriver:
		return s.postgresInsertLogbook(logbook)
	default:
		return logbook, errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

func (s *Service) sqliteInsertLogbook(logbook types.Logbook) (types.Logbook, error) {
	const op errors.Op = "database.Service.sqliteInsertLogbook"
	if err := checkService(op, s); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if h == nil || !isOpen {
		return logbook, errors.New(op).Msg(errMsgNotOpen)
	}

	adapter := adapters.New()

	model := &sqmodels.Logbook{}
	if err := adapter.Adapt(&logbook, model); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	ctx, cancel := s.withDefaultTimeout(nil)
	defer cancel()
	if err := model.Insert(ctx, h, boil.Infer()); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	// Update the returned types.Qso with the ID from the database
	logbook.ID = model.ID

	return logbook, nil
}

func (s *Service) postgresInsertLogbook(logbook types.Logbook) (types.Logbook, error) {
	const op errors.Op = "database.Service.postgresInsertLogbook"
	if err := checkService(op, s); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if h == nil || !isOpen {
		return logbook, errors.New(op).Msg(errMsgNotOpen)
	}

	adapter := adapters.New()

	model := &pgmodels.Logbook{}
	if err := adapter.Adapt(&logbook, model); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	ctx, cancel := s.withDefaultTimeout(nil)
	defer cancel()
	if err := model.Insert(ctx, h, boil.Infer()); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	logbook.ID = model.ID

	return logbook, nil
}
