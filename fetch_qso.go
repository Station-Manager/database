package database

import (
	pgmodels "github.com/Station-Manager/database/postgres/models"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
)

func (s *Service) FetchQsoById(id int64) (types.Qso, error) {
	const op errors.Op = "database.Service.FetchQsoById"
	emptyRetVal := types.Qso{}
	if err := checkService(op, s); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}
	if id < 1 {
		return emptyRetVal, errors.New(op).Msg(errMsgInvalidId)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteFetchQso(id)
	case PostgresDriver:
		return s.postgresFetchQso(id)
	default:
		return emptyRetVal, errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

func (s *Service) sqliteFetchQso(id int64) (types.Qso, error) {
	const op errors.Op = "database.Service.sqliteFetchQso"

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

	ctx, cancel := s.withDefaultTimeout(nil)
	defer cancel()
	model, err := sqmodels.FindQso(ctx, h, id)
	if err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	// Use cached adapter for model -> type conversion
	s.initAdapters()
	adapter := s.adapterFromModel

	typeQso := types.Qso{}
	if err = adapter.Adapt(model, &typeQso); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	return typeQso, nil
}

func (s *Service) postgresFetchQso(id int64) (types.Qso, error) {
	const op errors.Op = "database.Service.postgresFetchQso"
	emptyRetVal := types.Qso{}
	if err := checkService(op, s); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}
	if id < 1 {
		return emptyRetVal, errors.New(op).Msg(errMsgInvalidId)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if h == nil || !isOpen {
		return emptyRetVal, errors.New(op).Msg(errMsgNotOpen)
	}

	ctx, cancel := s.withDefaultTimeout(nil)
	defer cancel()

	model, err := pgmodels.FindQso(ctx, h, id)
	if err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	// Cached adapter
	s.initAdapters()
	adapter := s.adapterFromModel

	typeQso := types.Qso{}
	if err = adapter.Adapt(model, &typeQso); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	return typeQso, nil
}
