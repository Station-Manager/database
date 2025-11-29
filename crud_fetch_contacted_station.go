package database

import (
	"context"
	"database/sql"
	stderr "errors"
	"github.com/Station-Manager/adapters"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
)

// FetchContactedStationById retrieves a contacted station's details by its unique identifier from persistent storage.
func (s *Service) FetchContactedStationById(id int64) (types.ContactedStation, error) {
	return s.FetchContactedStationByIdContext(context.Background(), id)
}

// FetchContactedStationByIdContext retrieves a contacted station by ID within the given context, implementing driver-specific logic.
func (s *Service) FetchContactedStationByIdContext(ctx context.Context, id int64) (types.ContactedStation, error) {
	const op errors.Op = "database.Service.FetchContactedStationByIdContext"

	emptyRetVal := types.ContactedStation{}
	if err := checkService(op, s); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}
	if id < 1 {
		return emptyRetVal, errors.New(op).Msg(errMsgInvalidId)
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteFetchContactedStationByIdContext(ctx, id)
	case PostgresDriver:
		return emptyRetVal, errors.New(op).Msg("Not supported. Desktop application only.")
	default:
		return emptyRetVal, errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

// sqliteFetchContactedStationByIdContext retrieves a contacted station by its ID using SQLite as the database driver.
// It ensures the service is open and initialized, checks the context for cancellation, and fetches the record by ID.
// The method returns the fetched ContactedStation or an error if one occurs during the process.
func (s *Service) sqliteFetchContactedStationByIdContext(ctx context.Context, id int64) (types.ContactedStation, error) {
	const op errors.Op = "database.Service.sqliteFetchContactContext"
	emptyRetVal := types.ContactedStation{}
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

	model, err := sqmodels.FindContactedStation(ctx, h, id)
	if err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	adapter := adapters.New()
	out := types.ContactedStation{}
	if err = adapter.Into(&out, model); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	return out, nil
}

// FetchContactedStationByCallsign retrieves a contacted station's details using the provided callsign.
// It returns a ContactedStation object or an error if the fetch operation fails.
func (s *Service) FetchContactedStationByCallsign(callsign string) (types.ContactedStation, error) {
	return s.FetchContactedStationByCallsignContext(context.Background(), callsign)
}

// FetchContactedStationByCallsignContext retrieves a contacted station's details by callsign within a given context.
// Returns a ContactedStation object or an error if the retrieval fails or the database driver is unsupported.
func (s *Service) FetchContactedStationByCallsignContext(ctx context.Context, callsign string) (types.ContactedStation, error) {
	const op errors.Op = "database.Service.FetchContactedStationByCallsignContext"

	emptyRetVal := types.ContactedStation{
		Call: callsign,
	}
	if err := checkService(op, s); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}
	if callsign == "" {
		return emptyRetVal, errors.New(op).Msg("callsign cannot be empty")
	}

	switch s.DatabaseConfig.Driver {
	case SqliteDriver:
		return s.sqliteFetchContactedStationByCallsignContext(ctx, callsign)
	case PostgresDriver:
		return emptyRetVal, errors.New(op).Msg("Not supported. Desktop application only.")
	default:
		return emptyRetVal, errors.New(op).Errorf("Unsupported database driver: %s", s.DatabaseConfig.Driver)
	}
}

// sqliteFetchContactedStationByCallsignContext fetches a contacted station by callsign from the SQLite database within a given context.
// Returns a ContactedStation object and an error if any issue occurs during execution, such as database not open or record not found.
// It ensures proper locking of the service handle and sets a default timeout on the context if none is provided.
// Converts the SQLite result into the application-specific ContactedStation type before returning.
func (s *Service) sqliteFetchContactedStationByCallsignContext(ctx context.Context, callsign string) (types.ContactedStation, error) {
	const op errors.Op = "database.Service.sqliteFetchContactedStationByCallsign"
	emptyRetVal := types.ContactedStation{}
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

	model, err := sqmodels.ContactedStations(sqmodels.ContactedStationWhere.Callsign.EQ(callsign)).One(ctx, h)
	if err != nil && !stderr.Is(err, sql.ErrNoRows) {
		return emptyRetVal, errors.New(op).Err(err)
	}

	if model == nil || err != nil {
		return emptyRetVal, errors.New(op).Err(sql.ErrNoRows).Msg("Contacted station not found")
	}

	s.initAdapters()
	contactedStationType, err := s.AdaptSqliteModelToTypeContactedStation(model)
	if err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	return contactedStationType, nil
}
