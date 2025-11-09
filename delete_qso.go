package database

import (
	pgmodels "github.com/Station-Manager/database/postgres/models"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
)

func (s *Service) DeleteQso(id int64) error {
	const op errors.Op = "database.Service.DeleteQso"

	if err := checkService(op, s); err != nil {
		return errors.New(op).Err(err)
	}
	if id < 1 {
		return errors.New(op).Msg(errMsgInvalidId)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteDeleteQso(id)
	case PostgresDriver:
		return s.postgresDeleteQso(id)
	default:
		return errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

func (s *Service) sqliteDeleteQso(id int64) error {
	const op errors.Op = "database.Service.sqliteDeleteQso"

	if err := checkService(op, s); err != nil {
		return errors.New(op).Err(err)
	}

	s.mu.RLock()
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if !isOpen {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	ctx, cancel := s.withDefaultTimeout(nil)
	defer cancel()

	tx, txCancel, err := s.BeginTxContext(ctx)
	if err != nil {
		return errors.New(op).Err(err)
	}
	defer txCancel()

	model, err := sqmodels.FindQso(ctx, tx, id)
	if err != nil {
		return errors.New(op).Err(err)
	}
	if _, err = model.Delete(ctx, tx, false); err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}

	return nil
}

func (s *Service) postgresDeleteQso(id int64) error {
	const op errors.Op = "database.Service.postgresDeleteQso"

	if err := checkService(op, s); err != nil {
		return errors.New(op).Err(err)
	}

	s.mu.RLock()
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if !isOpen {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	ctx, cancel := s.withDefaultTimeout(nil)
	defer cancel()

	tx, txCancel, err := s.BeginTxContext(ctx)
	if err != nil {
		return errors.New(op).Err(err)
	}
	defer txCancel()

	model, err := pgmodels.FindQso(ctx, tx, id)
	if err != nil {
		return errors.New(op).Err(err)
	}
	if _, err = model.Delete(ctx, tx); err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}
	if err = tx.Commit(); err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}

	return nil
}
