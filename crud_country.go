package database

import (
	"context"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"strings"
)

func (s *Service) CountryExistsByName(countryName string) (bool, error) {
	return s.CountryExistsByNameContext(context.Background(), countryName)
}

func (s *Service) CountryExistsByNameContext(ctx context.Context, countryName string) (bool, error) {
	const op errors.Op = "database.Service.CountryExistsByName"
	if err := checkService(op, s); err != nil {
		return false, errors.New(op).Err(err)
	}
	countryName = strings.TrimSpace(countryName)
	if countryName == "" {
		return false, errors.New(op).Msg("country name is empty")
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteCountryExistsByNameContext(ctx, countryName)
	case PostgresDriver:
		return false, errors.New(op).Msg("Not supported. Desktop application only.")
	default:
		return false, errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

func (s *Service) sqliteCountryExistsByNameContext(ctx context.Context, countryName string) (bool, error) {
	const op errors.Op = "database.Service.sqliteCountryExistsByNameContext"
	if err := checkService(op, s); err != nil {
		return false, errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()
	if h == nil || !isOpen {
		return false, errors.New(op).Msg(errMsgNotOpen)
	}

	// Apply default timeout if caller did not set one
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	exists, err := sqmodels.Countries(sqmodels.CountryWhere.Name.EQ(countryName)).Exists(ctx, h)
	if err != nil {
		return false, errors.New(op).Err(err)
	}
	return exists, nil
}
