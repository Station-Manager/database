package database

import (
	"context"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// InsertQso inserts a QSO using a background context with default timeout semantics.
func (s *Service) InsertQso(qso types.Qso) (types.Qso, error) {
	return s.InsertQsoContext(context.Background(), qso)
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

// sqliteInsertQsoContext inserts a QSO entry into the SQLite database within the given context and returns the updated QSO.
func (s *Service) sqliteInsertQsoContext(ctx context.Context, qso types.Qso) (types.Qso, error) {
	const op errors.Op = "database.Service.sqliteInsertQsoContext"
	if err := checkService(op, s); err != nil {
		return qso, errors.New(op).Err(err)
	}
	if qso.LogbookID < 1 {
		return qso, errors.New(op).Msg("LogbookID is required")
	}
	// SessionID is required for SQLite, but not for PostgreSQL. This is because SQLite is only used for desktop apps
	// where sessions are used.
	if qso.SessionID < 1 {
		return qso, errors.New(op).Msg("SessionID is required")
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

	if err = model.Insert(ctx, h, boil.Infer()); err != nil {
		return qso, errors.New(op).Err(err)
	}
	qso.ID = model.ID
	return qso, nil
}

// postgresInsertQsoContext inserts a QSO record into the PostgreSQL database and returns the updated QSO object or an error.
// It ensures the database service is initialized and open, and applies a default timeout if none exists in the context.
// The method adapts the QSO type to the PostgreSQL model before performing the insert operation.
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
