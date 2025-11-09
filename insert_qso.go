package database

import (
	"context"
	"github.com/Station-Manager/adapters"
	"github.com/Station-Manager/adapters/converters/postgres"
	"github.com/Station-Manager/adapters/converters/sqlite"
	pgmodels "github.com/Station-Manager/database/postgres/models"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// InsertQso inserts a QSO into the database. The returned QSO will have an ID set.
func (s *Service) InsertQso(qso types.Qso) (types.Qso, error) {
	const op errors.Op = "database.Service.InsertQso"
	if s == nil {
		return qso, errors.New(op).Msg(errMsgNilService)
	}
	if !s.isOpen.Load() {
		return qso, errors.New(op).Msg(errMsgNotOpen)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteInsertQso(qso)
	case PostgresDriver:
		return s.postgresInsertQso(qso)
	default:
		return qso, errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

func (s *Service) sqliteInsertQso(qso types.Qso) (types.Qso, error) {
	const op errors.Op = "database.Service.sqliteInsertQso"
	if err := checkService(op, s); err != nil {
		return qso, errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if h == nil || !isOpen {
		return qso, errors.New(op).Msg(errMsgNotOpen)
	}

	adapter := adapters.New()
	adapter.RegisterConverter("QsoDate", sqlite.TypeToModelDateConverter)
	adapter.RegisterConverter("TimeOn", sqlite.TypeToModelTimeConverter)
	adapter.RegisterConverter("TimeOff", sqlite.TypeToModelTimeConverter)
	adapter.RegisterConverter("Freq", sqlite.TypeToModelFreqConverter)
	adapter.RegisterConverter("Country", sqlite.TypeToModelCountryConverter)

	model := &sqmodels.Qso{}
	if err := adapter.Adapt(&qso, model); err != nil {
		return qso, errors.New(op).Err(err)
	}

	ctx, cancel := s.withDefaultTimeout(nil)
	defer cancel()
	if err := model.Insert(ctx, h, boil.Infer()); err != nil {
		return qso, errors.New(op).Err(err)
	}

	//	qso.ID = model.ID

	return qso, nil
}

func (s *Service) postgresInsertQso(qso types.Qso) (types.Qso, error) {
	const op errors.Op = "database.Service.postgresInsertQso"
	if err := checkService(op, s); err != nil {
		return qso, errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if h == nil || !isOpen {
		return qso, errors.New(op).Msg(errMsgNotOpen)
	}

	adapter := adapters.New()
	adapter.RegisterConverter("QsoDate", postgres.TypeToModelDateConverter)
	adapter.RegisterConverter("TimeOn", postgres.TypeToModelTimeConverter)
	adapter.RegisterConverter("Freq", postgres.TypeToModelFreqConverter)

	model := &pgmodels.Qso{}
	if err := adapter.Adapt(&qso, model); err != nil {
		return qso, errors.New(op).Err(err)
	}

	if err := model.Insert(context.Background(), h, boil.Infer()); err != nil {
		return qso, errors.New(op).Err(err)
	}

	qso.ID = model.ID

	return qso, nil
}
