package sqlite

import (
	"context"

	"github.com/Station-Manager/enums/upload"
	"github.com/Station-Manager/types"
)

/**********************************************************************************************************************
 * QSO Methods
 **********************************************************************************************************************/

func (s *Service) InsertQso(qso types.Qso) (int64, error) {
	return s.InsertQsoWithContext(context.Background(), qso)
}

func (s *Service) UpdateQso(qso types.Qso) error {
	return s.UpdateQsoWithContext(context.Background(), qso)
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

func (s *Service) InsertQsoUpload(qsoId int64, service upload.OnlineService) error {
	return s.InsertQsoUploadWithContext(context.Background(), qsoId, service)
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

func (s *Service) FetchCountryByName(name string) (types.Country, error) {
	return s.FetchCountryByNameWithContext(context.Background(), name)
}

func (s *Service) InsertCountry(country types.Country) (int64, error) {
	return s.InsertCountryWithContext(context.Background(), country)
}

func (s *Service) UpdateCountry(country types.Country) error {
	return s.UpdateCountryWithContext(context.Background(), country)
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

/**********************************************************************************************************************
 * Contest Related Methods
 **********************************************************************************************************************/

func (s *Service) IsContestDuplicateByLogbookID(id int64, callsign, band string) (bool, error) {
	return s.IsContestDuplicatByLogbookIDWithContext(context.Background(), id, callsign, band)
}

/**********************************************************************************************************************
 * Upload Methods
 **********************************************************************************************************************/

func (s *Service) FetchPendingUploads() ([]types.QsoUpload, error) {
	return s.FetchPendingUploadsWithContext(context.Background())
}

func (s *Service) UpdateQsoUploadStatus(id int64, status string, attempts int64, lastError string) error {
	return s.UpdateQsoUploadStatusWithContext(context.Background(), id, status, attempts, lastError)
}
