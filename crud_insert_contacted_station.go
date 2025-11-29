package database

import (
	"context"
	"github.com/Station-Manager/adapters"
	"github.com/Station-Manager/adapters/converters/common"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func (s *Service) InsertContactedStation(station types.ContactedStation) (types.ContactedStation, error) {
	return s.InsertContactedStationContext(context.Background(), station)
}

func (s *Service) InsertContactedStationContext(ctx context.Context, station types.ContactedStation) (types.ContactedStation, error) {
	const op errors.Op = "database.Service.InsertContactedStationContext"
	if err := checkService(op, s); err != nil {
		return station, errors.New(op).Err(err)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteInsertContactedStationContext(ctx, station)
	case PostgresDriver:
		return station, errors.New(op).Msg("Not supported. Desktop application only.")
	default:
		return station, errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

func (s *Service) sqliteInsertContactedStationContext(ctx context.Context, station types.ContactedStation) (types.ContactedStation, error) {
	const op errors.Op = "database.Service.sqliteInsertContactedStationContext"
	if err := checkService(op, s); err != nil {
		return station, errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()
	if h == nil || !isOpen {
		return station, errors.New(op).Msg(errMsgNotOpen)
	}

	// Apply default timeout if caller did not set one
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	adapter := adapters.New()
	adapter.RegisterConverter("country", common.TypeToModelStringConverter)
	model := &sqmodels.ContactedStation{}
	if err := adapter.Into(&model, station); err != nil {
		return station, errors.New(op).Err(err)
	}

	if err := model.Insert(ctx, h, boil.Infer()); err != nil {
		return station, errors.New(op).Err(err)
	}

	return station, nil
}
