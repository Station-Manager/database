package database

import (
	"context"
	pgmodels "github.com/Station-Manager/database/postgres/models"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

/*********************************************************************************************************************
Fetch QSO Methods
**********************************************************************************************************************/

// FetchQsoById retrieves a QSO record by its unique ID and returns it, defaulting to the background context for execution.
func (s *Service) FetchQsoById(id int64) (types.Qso, error) {
	return s.FetchQsoByIdContext(context.Background(), id)
}

// FetchQsoByIdContext fetches a QSO record by its unique ID within the provided context and returns it or an error.
func (s *Service) FetchQsoByIdContext(ctx context.Context, id int64) (types.Qso, error) {
	const op errors.Op = "database.Service.FetchQsoByIdContext"
	emptyRetVal := types.Qso{}
	if err := checkService(op, s); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}
	if id < 1 {
		return emptyRetVal, errors.New(op).Msg(errMsgInvalidId)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteFetchQsoContext(ctx, id)
	case PostgresDriver:
		return s.postgresFetchQsoContext(ctx, id)
	default:
		return emptyRetVal, errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

// sqliteFetchQsoContext retrieves a QSO record by its ID within the provided context using SQLite as the database driver.
// It initializes required adapters, checks the service state, and returns the QSO record or an error.
func (s *Service) sqliteFetchQsoContext(ctx context.Context, id int64) (types.Qso, error) {
	const op errors.Op = "database.Service.sqliteFetchQsoContext"
	emptyRetVal := types.Qso{}
	if err := checkService(op, s); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()
	if h == nil || !isOpen {
		return emptyRetVal, errors.New(op).Msg(errMsgNotOpen)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	model, err := sqmodels.FindQso(ctx, h, id)
	if err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	s.initAdapters()
	adapter := s.adapterFromModel
	out := types.Qso{}
	if err := adapter.Into(&out, model); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}
	return out, nil
}

// postgresFetchQsoContext retrieves a QSO record by its ID from a PostgreSQL database and returns a populated Qso object.
// It validates the service state, uses a database connection, applies a default timeout if none is set, and adapts the
// database result into the application's Qso type. Returns an error if any operation fails.
func (s *Service) postgresFetchQsoContext(ctx context.Context, id int64) (types.Qso, error) {
	const op errors.Op = "database.Service.postgresFetchQsoContext"
	emptyRetVal := types.Qso{}
	if err := checkService(op, s); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()
	if h == nil || !isOpen {
		return emptyRetVal, errors.New(op).Msg(errMsgNotOpen)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	model, err := pgmodels.FindQso(ctx, h, id)
	if err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	s.initAdapters()
	adapter := s.adapterFromModel
	out := types.Qso{}
	if err := adapter.Into(&out, model); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}
	return out, nil
}

/*********************************************************************************************************************
Insert QSO Methods
**********************************************************************************************************************/

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

/*********************************************************************************************************************
Update QSO Methods
**********************************************************************************************************************/

// UpdateQso delegates to UpdateQsoContext with a background context.
func (s *Service) UpdateQso(qso types.Qso) error {
	return s.UpdateQsoContext(context.Background(), qso)
}

// UpdateQsoContext updates an existing QSO with caller-provided context.
func (s *Service) UpdateQsoContext(ctx context.Context, qso types.Qso) error {
	const op errors.Op = "database.Service.UpdateQsoContext"
	if err := checkService(op, s); err != nil {
		return errors.New(op).Err(err)
	}
	if qso.ID < 1 {
		return errors.New(op).Msg(errMsgInvalidId)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteUpdateQsoContext(ctx, qso)
	case PostgresDriver:
		return s.postgresUpdateQsoContext(ctx, qso)
	default:
		return errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

// sqliteUpdateQsoContext updates a QSO record in the SQLite database within a transactional context.
// It ensures the service is operational, adapts a QSO object to the SQLite model, and commits changes if successful.
// Returns an error if the service is not ready, the transaction fails, or the update operation encounters issues.
func (s *Service) sqliteUpdateQsoContext(ctx context.Context, qso types.Qso) error {
	const op errors.Op = "database.Service.sqliteUpdateQsoContext"
	if err := checkService(op, s); err != nil {
		return errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if h == nil || !isOpen {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	tx, txCancel, err := s.BeginTxContext(ctx)
	if err != nil {
		return errors.New(op).Err(err)
	}
	defer txCancel()

	s.initAdapters()

	model, err := s.AdaptTypeToSqliteModelQso(qso)
	if err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}

	model.ID = qso.ID

	rows, err := model.Update(ctx, tx, boil.Infer())
	if err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}
	if rows == 0 {
		_ = tx.Rollback()
		return errors.New(op).Msg(errMsgInvalidId)
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}
	return nil
}

// postgresUpdateQsoContext updates the QSO record in the PostgreSQL database using the provided context and QSO data.
// Returns an error if the update fails, the database service is not open, or validation conditions are not met.
func (s *Service) postgresUpdateQsoContext(ctx context.Context, qso types.Qso) error {
	const op errors.Op = "database.Service.postgresUpdateQsoContext"
	if err := checkService(op, s); err != nil {
		return errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if h == nil || !isOpen {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	tx, txCancel, err := s.BeginTxContext(ctx)
	if err != nil {
		return errors.New(op).Err(err)
	}
	defer txCancel()

	s.initAdapters()

	model, err := s.AdaptTypeToPostgresModelQso(qso)
	if err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}

	model.ID = qso.ID

	rows, err := model.Update(ctx, tx, boil.Infer())
	if err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}
	if rows == 0 {
		_ = tx.Rollback()
		return errors.New(op).Msg(errMsgInvalidId)
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}
	return nil
}

/*********************************************************************************************************************
Delete QSO Methods
**********************************************************************************************************************/

// DeleteQso delegates to DeleteQsoContext with a background context.
func (s *Service) DeleteQso(id int64) error {
	return s.DeleteQsoContext(context.Background(), id)
}

// DeleteQsoContext deletes a QSO with a caller-provided context.
func (s *Service) DeleteQsoContext(ctx context.Context, id int64) error {
	const op errors.Op = "database.Service.DeleteQsoContext"
	if err := checkService(op, s); err != nil {
		return errors.New(op).Err(err)
	}
	if id < 1 {
		return errors.New(op).Msg(errMsgInvalidId)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()
	if h == nil || !isOpen {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	tx, txCancel, err := s.BeginTxContext(ctx)
	if err != nil {
		return errors.New(op).Err(err)
	}
	defer txCancel()

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		if rows, err := sqmodels.Qsos(sqmodels.QsoWhere.ID.EQ(id)).DeleteAll(ctx, tx, false); err != nil {
			_ = tx.Rollback()
			return errors.New(op).Err(err)
		} else if rows == 0 {
			_ = tx.Rollback()
			return errors.New(op).Msg(errMsgInvalidId)
		}
	case PostgresDriver:
		if rows, err := pgmodels.Qsos(pgmodels.QsoWhere.ID.EQ(id)).DeleteAll(ctx, tx); err != nil {
			_ = tx.Rollback()
			return errors.New(op).Err(err)
		} else if rows == 0 {
			_ = tx.Rollback()
			return errors.New(op).Msg(errMsgInvalidId)
		}
	default:
		_ = tx.Rollback()
		return errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}

	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}
	return nil
}
