package database

import (
	"context"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func (s *Service) InsertLogbook(logbook types.Logbook) (types.Logbook, error) {
	return s.InsertLogbookContext(context.Background(), logbook)
}

func (s *Service) InsertLogbookContext(ctx context.Context, logbook types.Logbook) (types.Logbook, error) {
	const op errors.Op = "database.Service.InsertLogbookContext"
	if err := checkService(op, s); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteInsertLogbookContext(ctx, logbook)
	case PostgresDriver:
		return s.postgresInsertLogbookContext(ctx, logbook)
	default:
		return logbook, errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

func (s *Service) sqliteInsertLogbookContext(ctx context.Context, logbook types.Logbook) (types.Logbook, error) {
	const op errors.Op = "database.Service.sqliteInsertLogbookContext"
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

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	s.initAdapters()
	model, err := s.AdaptTypeToSqliteModelLogbook(logbook)
	if err != nil {
		return logbook, errors.New(op).Err(err)
	}

	if err := model.Insert(ctx, h, boil.Infer()); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	logbook.ID = model.ID
	return logbook, nil
}

func (s *Service) postgresInsertLogbookContext(ctx context.Context, logbook types.Logbook) (types.Logbook, error) {
	const op errors.Op = "database.Service.postgresInsertLogbookContext"
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

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	s.initAdapters()
	model, err := s.AdaptTypeToPostgresModelLogbook(logbook)
	if err != nil {
		return logbook, errors.New(op).Err(err)
	}

	if err := model.Insert(ctx, h, boil.Infer()); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	logbook.ID = model.ID
	return logbook, nil
}
