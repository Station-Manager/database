package sqlite

import (
	"context"
	"github.com/Station-Manager/types"
)

func (s *Service) InsertQso(qso types.Qso) (int64, error) {
	return s.InsertQsoWithContext(context.Background(), qso)
}

func (s *Service) FetchContactedStationByCallsign(callsign string) (types.ContactedStation, error) {
	return s.FetchContactedStationByCallsignWithContext(context.Background(), callsign)
}

func (s *Service) FetchCountryByCallsign(callsign string) (types.Country, error) {
	return s.FetchCountryByCallsignWithContext(context.Background(), callsign)
}

func (s *Service) FetchContactHistory(callsign string) ([]types.ContactHistory, error) {
	return s.FetchContactHistoryWithContext(context.Background(), callsign)
}
