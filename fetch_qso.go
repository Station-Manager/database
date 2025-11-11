package database

import (
	"context"
	pgmodels "github.com/Station-Manager/database/postgres/models"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
)

func (s *Service) FetchQsoById(id int64) (types.Qso, error) {
	return s.FetchQsoByIdContext(context.Background(), id)
}

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
	if err := adapter.Adapt(model, &out); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}
	return out, nil
}

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
	if err := adapter.Adapt(model, &out); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}
	return out, nil
}
