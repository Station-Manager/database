package database

import (
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
	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if h == nil || !isOpen {
		return emptyRetVal, errors.New(op).Msg(errMsgNotOpen)
	}

	//ctx, cancel := s.withDefaultTimeout(nil)
	//defer cancel()
	//model, err := sqmodels.FindQso(ctx, h, id)
	//if err != nil && !stderr.Is(err, sql.ErrNoRows) {
	//	return emptyRetVal, errors.New(op).Err(err)
	//}
	//
	//adapter := adapters.New()
	//adapter.RegisterConverter("Freq", sqlite.ModelToTypeFreqConverter)
	//adapter.RegisterConverter("Country", sqlite.ModelToTypeCountryConverter)
	//
	typeQso := types.Qso{}
	//if err = adapter.Adapt(model, &typeQso); err != nil {
	//	return emptyRetVal, errors.New(op).Err(err)
	//}

	return typeQso, nil
}

func (s *Service) postgresFetchQso(id int64) (types.Qso, error) {
	return types.Qso{}, nil
}
