package sqlite

import (
	"context"
	"github.com/Station-Manager/types"
)

/**********************************************************************************************************************
 * QSO Methods
 **********************************************************************************************************************/

func (s *Service) InsertQso(qso types.Qso) (int64, error) {
	return s.InsertQsoWithContext(context.Background(), qso)
}

func (s *Service) FetchQsoSliceBySessionID(id int64) ([]types.Qso, error) {
	return s.FetchQsoSliceBySessionIDWithContext(context.Background(), id)
}

func (s *Service) FetchQsoSliceByCallsign(callsign string) ([]types.ContactHistory, error) {
	return s.FetchQsoSliceByCallsignWithContext(context.Background(), callsign)
}

func (s *Service) FetchQsoCountByLogbookId(id int64) (int64, error) {
	return s.FetchQsoCountByLogbookIdWithContext(context.Background(), id)
}

/**********************************************************************************************************************
 * ContactedStation Methods
 **********************************************************************************************************************/

func (s *Service) FetchContactedStationByCallsign(callsign string) (types.ContactedStation, error) {
	return s.FetchContactedStationByCallsignWithContext(context.Background(), callsign)
}

func (s *Service) InsertContactedStation(station types.ContactedStation) (int64, error) {
	return s.InsertContactedStationWithContext(context.Background(), station)
}

func (s *Service) UpdateContactedStation(station types.ContactedStation) error {
	return s.UpdateContactedStationWithContext(context.Background(), station)
}

/**********************************************************************************************************************
 * Country Methods
 **********************************************************************************************************************/

func (s *Service) FetchCountryByCallsign(callsign string) (types.Country, error) {
	return s.FetchCountryByCallsignWithContext(context.Background(), callsign)
}

/**********************************************************************************************************************
 * Logbook Methods
 **********************************************************************************************************************/

func (s *Service) FetchLogbookByID(id int64) (types.Logbook, error) {
	return s.FetchLogbookByIDWithContext(context.Background(), id)
}

/**********************************************************************************************************************
 * Session Methods
 **********************************************************************************************************************/

func (s *Service) SoftDeleteSessionByID(id int64) error {
	return s.SoftDeleteSessionByIDWithContext(context.Background(), id)
}

func (s *Service) InsertSession() (int64, error) {
	return s.InsertSessionWithContext(context.Background())
}
