package database

import (
	"github.com/Station-Manager/adapters"
	"github.com/Station-Manager/adapters/converters/common"
	"github.com/Station-Manager/adapters/converters/postgres"
	"github.com/Station-Manager/adapters/converters/sqlite"
	pgmodels "github.com/Station-Manager/database/postgres/models"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// UpdateQso updates an existing QSO row. It requires qso.ID to be > 0.
func (s *Service) UpdateQso(qso types.Qso) error {
	const op errors.Op = "database.Service.UpdateQso"
	if err := checkService(op, s); err != nil {
		return errors.New(op).Err(err)
	}
	if qso.ID < 1 {
		return errors.New(op).Msg(errMsgInvalidId)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteUpdateQso(qso)
	case PostgresDriver:
		return s.postgresUpdateQso(qso)
	default:
		return errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

func (s *Service) sqliteUpdateQso(qso types.Qso) error {
	const op errors.Op = "database.Service.sqliteUpdateQso"
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

	ctx, cancel := s.withDefaultTimeout(nil)
	defer cancel()

	tx, txCancel, err := s.BeginTxContext(ctx)
	if err != nil {
		return errors.New(op).Err(err)
	}
	defer txCancel()

	adapter := adapters.New()
	adapter.RegisterConverter("QsoDate", sqlite.TypeToModelDateConverter)
	adapter.RegisterConverter("TimeOn", sqlite.TypeToModelTimeConverter)
	adapter.RegisterConverter("TimeOff", sqlite.TypeToModelTimeConverter)
	adapter.RegisterConverter("Freq", common.TypeToModelFreqConverter)
	adapter.RegisterConverter("Country", common.TypeToModelStringConverter)

	model := &sqmodels.Qso{}
	if err := adapter.Adapt(&qso, model); err != nil {
		_ = tx.Rollback()
		return errors.New(op).Err(err)
	}

	// Ensure primary key is set
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

func (s *Service) postgresUpdateQso(qso types.Qso) error {
	const op errors.Op = "database.Service.postgresUpdateQso"
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

	ctx, cancel := s.withDefaultTimeout(nil)
	defer cancel()

	tx, txCancel, err := s.BeginTxContext(ctx)
	if err != nil {
		return errors.New(op).Err(err)
	}
	defer txCancel()

	adapter := adapters.New()
	adapter.RegisterConverter("QsoDate", postgres.TypeToModelDateConverter)
	adapter.RegisterConverter("TimeOn", postgres.TypeToModelTimeConverter)
	adapter.RegisterConverter("TimeOff", postgres.TypeToModelTimeConverter)
	adapter.RegisterConverter("Freq", common.TypeToModelFreqConverter)
	adapter.RegisterConverter("Country", common.TypeToModelStringConverter)

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
