package database

import (
	"context"
	pgmodels "github.com/Station-Manager/database/postgres/models"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
)

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
