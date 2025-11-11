package database

import (
	"context"
	pgmodels "github.com/Station-Manager/database/postgres/models"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

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
	adapter := s.adapterToModel

	model := &sqmodels.Qso{}
	if err := adapter.Adapt(&qso, model); err != nil {
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
	adapter := s.adapterToModel

	model := &pgmodels.Qso{}
	if err := adapter.Adapt(&qso, model); err != nil {
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
