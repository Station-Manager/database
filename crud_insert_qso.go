package database

import (
	"context"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// InsertQso inserts a QSO using a background context with default timeout semantics.
func (s *Service) InsertQso(qso types.Qso) (types.Qso, error) {
	ctx := context.Background()
	return s.InsertQsoContext(ctx, qso)
}

// InsertQsoContext inserts a QSO with caller-provided context.
// If the context has no deadline, a default timeout is applied.
func (s *Service) InsertQsoContext(ctx context.Context, qso types.Qso) (types.Qso, error) {
	const op errors.Op = "database.Service.InsertQsoContext"
	if err := checkService(op, s); err != nil {
		return qso, errors.New(op).Err(err)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteInsertQsoContext(ctx, qso)
	case PostgresDriver:
		return s.postgresInsertQsoContext(ctx, qso)
	default:
		return qso, errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

func (s *Service) sqliteInsertQsoContext(ctx context.Context, qso types.Qso) (types.Qso, error) {
	const op errors.Op = "database.Service.sqliteInsertQsoContext"
	if err := checkService(op, s); err != nil {
		return qso, errors.New(op).Err(err)
	}
	if qso.LogbookID < 1 {
		return qso, errors.New(op).Msg("LogbookID is required")
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()
	if h == nil || !isOpen {
		return qso, errors.New(op).Msg(errMsgNotOpen)
	}

	// Apply default timeout if caller did not set one
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	s.initAdapters()
	model, err := s.AdaptTypeToSqliteModelQso(qso)
	if err != nil {
		return qso, errors.New(op).Err(err)
	}

	if err := model.Insert(ctx, h, boil.Infer()); err != nil {
		return qso, errors.New(op).Err(err)
	}
	qso.ID = model.ID
	return qso, nil
}

func (s *Service) postgresInsertQsoContext(ctx context.Context, qso types.Qso) (types.Qso, error) {
	const op errors.Op = "database.Service.postgresInsertQsoContext"
	if err := checkService(op, s); err != nil {
		return qso, errors.New(op).Err(err)
	}
	if qso.LogbookID < 1 {
		return qso, errors.New(op).Msg("LogbookID is required")
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()
	if h == nil || !isOpen {
		return qso, errors.New(op).Msg(errMsgNotOpen)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	s.initAdapters()
	model, err := s.AdaptTypeToPostgresModelQso(qso)
	if err != nil {
		return qso, errors.New(op).Err(err)
	}

	if err = model.Insert(ctx, h, boil.Infer()); err != nil {
		return qso, errors.New(op).Err(err)
	}
	qso.ID = model.ID
	return qso, nil
}
