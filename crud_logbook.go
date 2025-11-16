package database

import (
	"context"
	stderr "errors"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
	"time"
)

// InsertLogbook creates and stores a new logbook entry in the database, returning the created logbook or an error.
func (s *Service) InsertLogbook(logbook types.Logbook) (types.Logbook, error) {
	return s.InsertLogbookContext(context.Background(), logbook)
}

// InsertLogbookContext inserts a new logbook into the database within the provided context and returns the created logbook or an error.
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

// InsertLogbookWithTxContext performs a logbook insert using the provided transaction and context.
// This is used by higher-level operations that need to coordinate multiple writes atomically.
func (s *Service) InsertLogbookWithTxContext(ctx context.Context, tx boil.ContextExecutor, logbook types.Logbook) (types.Logbook, error) {
	const op errors.Op = "database.Service.InsertLogbookWithTxContext"
	if err := checkService(op, s); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteInsertLogbookWithTxContext(ctx, tx, logbook)
	case PostgresDriver:
		return s.postgresInsertLogbookWithTxContext(ctx, tx, logbook)
	default:
		return logbook, errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

// sqliteInsertLogbookContext inserts a Logbook record into an SQLite database, ensuring service validity and context deadlines.
// It adapts the input Logbook to an internal SQLite model, performs the database operation, and returns the updated Logbook.
// Errors encountered during validation, adaptation, insertion, or service checks are wrapped with operation metadata and returned.
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

	// DEBUG: context deadlines & timing
	if dl, ok := ctx.Deadline(); ok {
		s.Logger.DebugWith().Str("component", "db").Str("driver", "sqlite").Str("op", "insert_logbook").Time("start", time.Now()).Time("deadline", dl).Msg("starting insert")
	} else {
		s.Logger.DebugWith().Str("component", "db").Str("driver", "sqlite").Str("op", "insert_logbook").Time("start", time.Now()).Msg("starting insert (no deadline)")
	}

	if err = model.Insert(ctx, h, boil.Infer()); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	logbook.ID = model.ID
	return logbook, nil
}

// sqliteInsertLogbookWithTxContext mirrors sqliteInsertLogbookContext but uses the
// provided transaction instead of the shared handle. It assumes the caller has already
// begun the transaction and applied any desired timeouts to ctx.
func (s *Service) sqliteInsertLogbookWithTxContext(ctx context.Context, tx boil.ContextExecutor, logbook types.Logbook) (types.Logbook, error) {
	const op errors.Op = "database.Service.sqliteInsertLogbookWithTxContext"
	if err := checkService(op, s); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	if tx == nil {
		return logbook, errors.New(op).Msg("transaction is nil")
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

	if err = model.Insert(ctx, tx, boil.Infer()); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	logbook.ID = model.ID
	return logbook, nil
}

// postgresInsertLogbookContext inserts a new logbook record into the database within a transactional context.
// It ensures appropriate context deadlines, adapts the logbook to the Postgres model, and commits the transaction.
// Returns the inserted logbook with its assigned ID or an error if the operation fails.
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

	// DEBUG: context deadlines & timing
	if dl, ok := ctx.Deadline(); ok {
		s.Logger.DebugWith().Str("component", "db").Str("driver", "postgres").Str("op", "insert_logbook").Time("start", time.Now()).Time("deadline", dl).Msg("starting insert")
	} else {
		s.Logger.DebugWith().Str("component", "db").Str("driver", "postgres").Str("op", "insert_logbook").Time("start", time.Now()).Msg("starting insert (no deadline)")
	}

	// For Postgres, perform the insert inside a transaction so BeginTxContext can set
	// a local statement_timeout (SET LOCAL) and bound server-side statement execution.
	tx, txCancel, err := s.BeginTxContext(ctx)
	if err != nil {
		return logbook, errors.New(op).Err(err)
	}
	defer txCancel()

	// Create a transaction-specific context for statement execution. If the caller hasn't
	// provided a deadline, use the configured TransactionContextTimeout so client-side
	// cancellation matches the server-side statement_timeout we set in BeginTxContext.
	txCtx := ctx
	var txLocalCancel context.CancelFunc = func() {}
	if _, ok := ctx.Deadline(); !ok {
		txCtx, txLocalCancel = context.WithTimeout(ctx, time.Duration(s.DatabaseConfig.TransactionContextTimeout)*time.Second)
		defer txLocalCancel()
	}

	if err = model.Insert(txCtx, tx, boil.Infer()); err != nil {
		_ = tx.Rollback()
		s.Logger.DebugWith().Str("component", "db").Str("driver", "postgres").Str("op", "insert_logbook").Time("at", time.Now()).Err(err).Interface("ctx_err", txCtx.Err()).Msg("insert error")
		// If the error is a deadline, capture current Postgres activity to help debug locks
		if txCtx.Err() != nil {
			s.logPostgresActivity()
		}
		return logbook, errors.New(op).Err(err)
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		s.Logger.DebugWith().Str("component", "db").Str("driver", "postgres").Str("op", "insert_logbook").Time("at", time.Now()).Err(err).Msg("commit error")
		if stderr.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			s.logPostgresActivity()
		}
		return logbook, errors.New(op).Err(err)
	}

	s.Logger.DebugWith().Str("component", "db").Str("driver", "postgres").Str("op", "insert_logbook").Time("at", time.Now()).Interface("ctx_err", txCtx.Err()).Msg("insert completed")

	logbook.ID = model.ID
	return logbook, nil
}

// postgresInsertLogbookWithTxContext mirrors postgresInsertLogbookContext but uses the
// provided transaction instead of creating a new one. This is intended for higher-level
// operations that already manage the transaction boundaries.
func (s *Service) postgresInsertLogbookWithTxContext(ctx context.Context, tx boil.ContextExecutor, logbook types.Logbook) (types.Logbook, error) {
	const op errors.Op = "database.Service.postgresInsertLogbookWithTxContext"
	if err := checkService(op, s); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	if tx == nil {
		return logbook, errors.New(op).Msg("transaction is nil")
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

	if err = model.Insert(ctx, tx, boil.Infer()); err != nil {
		return logbook, errors.New(op).Err(err)
	}

	logbook.ID = model.ID
	return logbook, nil
}
