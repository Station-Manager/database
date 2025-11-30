package database

import (
	"context"
	"database/sql"
	stderr "errors"
	sqmodels "github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

/*********************************************************************************************************************
Insert Contacted Station Methods
**********************************************************************************************************************/

// InsertContactedStation inserts a contacted station record into the datastore using the provided station details.
// It returns the inserted types.ContactedStation object or an error if the operation fails.
// It supports only SQLite and will return an error for unsupported database drivers or uninitialized services.
func (s *Service) InsertContactedStation(station types.ContactedStation) (types.ContactedStation, error) {
	return s.InsertContactedStationContext(context.Background(), station)
}

// InsertContactedStationContext inserts a contacted station record into the database in a specific context.
// It supports only SQLite and will return an error for unsupported database drivers or uninitialized services.
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

// sqliteInsertContactedStationContext inserts a ContactedStation into the SQLite database within the given context.
// It initializes adapters, applies the default timeout if none is set, and converts the ContactedStation to its SQLite model.
// Returns the inserted ContactedStation or an error in case of failure.
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

	s.initAdapters()
	model, err := s.AdaptTypeToSqliteModelContactedStation(station)
	if err != nil {
		return station, errors.New(op).Err(err)
	}
	if err = model.Insert(ctx, h, boil.Infer()); err != nil {
		return station, errors.New(op).Err(err)
	}

	return station, nil
}

/*********************************************************************************************************************
Fetch by ID - Contacted Station Methods
**********************************************************************************************************************/

// FetchContactedStationById retrieves a contacted station's details by its unique identifier from persistent storage.
// It supports only SQLite and will return an error for unsupported database drivers or uninitialized services.
func (s *Service) FetchContactedStationById(id int64) (types.ContactedStation, error) {
	return s.FetchContactedStationByIdContext(context.Background(), id)
}

// FetchContactedStationByIdContext retrieves a contacted station by ID within the given context, implementing driver-specific logic.
// It supports only SQLite and will return an error for unsupported database drivers or uninitialized services.
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

	s.initAdapters()
	out, err := s.AdaptSqliteModelToTypeContactedStation(model)
	if err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	return out, nil
}

/*********************************************************************************************************************
Fetch by callsign - Contacted Station Methods
**********************************************************************************************************************/

// FetchContactedStationByCallsign retrieves a contacted station's details using the provided callsign.
// It returns a ContactedStation object or an error if the fetch operation fails.
// It supports only SQLite and will return an error for unsupported database drivers or uninitialized services.
func (s *Service) FetchContactedStationByCallsign(callsign string) (types.ContactedStation, error) {
	return s.FetchContactedStationByCallsignContext(context.Background(), callsign)
}

// FetchContactedStationByCallsignContext retrieves a contacted station's details by callsign within a given context.
// Returns a ContactedStation object or an error if the retrieval fails or the database driver is unsupported.
// It supports only SQLite and will return an error for unsupported database drivers or uninitialized services.
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
