package adapters

import (
	"github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/types"
	"github.com/goccy/go-json"
	"strconv"
	"strings"
)

func QsoTypeToModel(qso types.Qso) (models.Qso, error) {
	// Parse frequency from types (string) to int64 expected by models.
	var freqHz int64
	if qso.QsoDetails.Freq != "" {
		// Attempt to parse as integer first (already Hz). If that fails, try parsing as float (e.g., MHz)
		if v, err := strconv.ParseInt(qso.QsoDetails.Freq, 10, 64); err == nil {
			freqHz = v
		} else if f, err2 := strconv.ParseFloat(qso.QsoDetails.Freq, 64); err2 == nil {
			// Assume MHz and convert to Hz
			freqHz = int64(f * 1_000_000)
		}
	}

	date := qso.QsoDetails.QsoDate
	if strings.Contains(date, "-") {
		date = strings.ReplaceAll(date, "-", "")
	}

	timeOn := qso.QsoDetails.TimeOn
	if strings.Contains(timeOn, ":") {
		timeOn = strings.ReplaceAll(timeOn, ":", "")
	}

	timeOff := qso.QsoDetails.TimeOff
	if strings.Contains(timeOff, ":") {
		timeOff = strings.ReplaceAll(timeOff, ":", "")
	}

	additionalData := struct {
		// QsoDetails
		AIndex      string `json:"a_index,omitempty"`
		AntPath     string `json:"ant_path,omitempty"` // ADIF, section II.B.1 - currently, we only use S and L
		BandRx      string `json:"band_rx,omitempty"`  //in a split frequency QSO, the logging station's receiving band
		Comment     string `json:"comment,omitempty"`
		ContestId   string `json:"contest_id,omitempty"`
		Distance    string `json:"distance,omitempty"` // km
		FreqRx      string `json:"freq_rx,omitempty"`
		Submode     string `json:"submode,omitempty"`
		Notes       string `json:"notes,omitempty"` // information of interest to the logging station's operator
		QsoDateOff  string `json:"qso_date_off,omitempty"`
		QsoRandom   string `json:"qso_random,omitempty"`
		QsoComplete string `json:"qso_complete,omitempty"`
		RxPwr       string `json:"rx_pwr,omitempty"` // the contacted station's transmitter power in Watts with a value greater than or equal to 0
		SRX         string `json:"srx,omitempty"`    // contest QSO received serial number with a value greater than or equal to 0
		STX         string `json:"stx,omitempty"`    // contest QSO transmitted serial number with a value greater than or equal to 0
		TxPwr       string `json:"tx_pwr,omitempty"` // the logging station's power in Watts with a value greater than or equal to 0
		// ContactedStation
		Address      string `json:"address,omitempty"`
		Age          string `json:"age,omitempty"`
		Altitude     string `json:"altitude,omitempty"`
		Cont         string `json:"cont,omitempty"` // the contacted station's Continent
		ContactedOp  string `json:"contacted_op,omitempty"`
		CQZ          string `json:"cqz,omitempty"`
		DXCC         string `json:"dxcc,omitempty"`
		Email        string `json:"email,omitempty"`
		EqCall       string `json:"eq_call,omitempty"` // the contacted station's owner's callsign (if different from call)
		Gridsquare   string `json:"gridsquare,omitempty"`
		Iota         string `json:"iota,omitempty"`
		IotaIslandId string `json:"iota_island_id,omitempty"`
		ITUZ         string `json:"ituz,omitempty"`
		Lat          string `json:"lat,omitempty"`
		Lon          string `json:"lon,omitempty"`
		Name         string `json:"name,omitempty"`
		QTH          string `json:"qth,omitempty"`
		Rig          string `json:"rig,omitempty"`
		Sig          string `json:"sig,omitempty"`      // the name of the contacted station's special activity or interest group
		SigInfo      string `json:"sig_info,omitempty"` // information associated with the contacted station's activity or interest group
		Web          string `json:"web,omitempty"`
		WwffRef      string `json:"wwff_ref,omitempty"`
		// LoggingStation
		AntennaAzimuth  string `json:"ant_az,omitempty"` // the bearing from the logging station to the contacted station
		MyAltitude      string `json:"my_altitude,omitempty"`
		MyAntenna       string `json:"my_antenna,omitempty"`
		MyCity          string `json:"my_city,omitempty"`
		MyCountry       string `json:"my_country,omitempty"`
		MyCqZone        string `json:"my_cq_zone,omitempty"`
		MyDXCC          string `json:"my_dxcc,omitempty"`
		MyGridsquare    string `json:"my_gridsquare,omitempty"`
		MyIota          string `json:"my_iota,omitempty"`
		MyIotaIslandID  string `json:"my_iota_island_id,omitempty"`
		MyITUZone       string `json:"my_itu_zone,omitempty"`
		MyLat           string `json:"my_lat,omitempty"`
		MyLon           string `json:"my_lon,omitempty"`
		MyMorseKeyInfo  string `json:"my_morse_key_info,omitempty"`
		MyMorseKeyType  string `json:"my_morse_key_type,omitempty"`
		MyName          string `json:"my_name,omitempty"`
		MyPostalCode    string `json:"my_postal_code,omitempty"`
		MyRig           string `json:"my_rig,omitempty"`
		MySig           string `json:"my_sig,omitempty"`
		MySigInfo       string `json:"my_sig_info,omitempty"`
		MyStreet        string `json:"my_street,omitempty"`
		MyWwffRef       string `json:"my_wwff_ref,omitempty"`
		Operator        string `json:"operator,omitempty"` // the logging operator's callsign if STATION_CALLSIGN is absent, OPERATOR shall be treated as both the logging station's callsign and the logging operator's callsign
		OwnerCallsign   string `json:"owner_callsign,omitempty"`
		StationCallsign string `json:"station_callsign"`
		// QSL (TBD)
	}{
		AIndex:      qso.QsoDetails.AIndex,
		AntPath:     qso.QsoDetails.AntPath,
		BandRx:      qso.QsoDetails.BandRx,
		Comment:     qso.QsoDetails.Comment,
		ContestId:   qso.QsoDetails.ContestId,
		Distance:    qso.QsoDetails.Distance,
		FreqRx:      qso.QsoDetails.FreqRx,
		Submode:     qso.QsoDetails.Submode,
		Notes:       qso.QsoDetails.Notes,
		QsoDateOff:  qso.QsoDetails.QsoDateOff,
		QsoRandom:   qso.QsoDetails.QsoRandom,
		QsoComplete: qso.QsoDetails.QsoComplete,
		RxPwr:       qso.QsoDetails.RxPwr,
		SRX:         qso.QsoDetails.SRX,
		STX:         qso.QsoDetails.STX,
		TxPwr:       qso.QsoDetails.TxPwr,

		// ContactedStation
		Address:      qso.ContactedStation.Address,
		Age:          qso.ContactedStation.Age,
		Altitude:     qso.ContactedStation.Altitude,
		Cont:         qso.ContactedStation.Cont,
		ContactedOp:  qso.ContactedStation.ContactedOp,
		CQZ:          qso.ContactedStation.CQZ,
		DXCC:         qso.ContactedStation.DXCC,
		Email:        qso.ContactedStation.Email,
		EqCall:       qso.ContactedStation.EqCall,
		Gridsquare:   qso.ContactedStation.Gridsquare,
		Iota:         qso.ContactedStation.Iota,
		IotaIslandId: qso.ContactedStation.IotaIslandId,
		ITUZ:         qso.ContactedStation.ITUZ,
		Lat:          qso.ContactedStation.Lat,
		Lon:          qso.ContactedStation.Lon,
		Name:         qso.ContactedStation.Name,
		QTH:          qso.ContactedStation.QTH,
		Rig:          qso.ContactedStation.Rig,
		Sig:          qso.ContactedStation.Sig,
		SigInfo:      qso.ContactedStation.SigInfo,
		Web:          qso.ContactedStation.Web,
		WwffRef:      qso.ContactedStation.WwffRef,

		// LoggingStation
		AntennaAzimuth:  qso.LoggingStation.AntennaAzimuth,
		MyAltitude:      qso.LoggingStation.MyAltitude,
		MyAntenna:       qso.LoggingStation.MyAntenna,
		MyCity:          qso.LoggingStation.MyCity,
		MyCountry:       qso.LoggingStation.MyCountry,
		MyCqZone:        qso.LoggingStation.MyCqZone,
		MyDXCC:          qso.LoggingStation.MyDXCC,
		MyGridsquare:    qso.LoggingStation.MyGridsquare,
		MyIota:          qso.LoggingStation.MyIota,
		MyIotaIslandID:  qso.LoggingStation.MyIotaIslandID,
		MyITUZone:       qso.LoggingStation.MyITUZone,
		MyLat:           qso.LoggingStation.MyLat,
		MyLon:           qso.LoggingStation.MyLon,
		MyMorseKeyInfo:  qso.LoggingStation.MyMorseKeyInfo,
		MyMorseKeyType:  qso.LoggingStation.MyMorseKeyType,
		MyName:          qso.LoggingStation.MyName,
		MyPostalCode:    qso.LoggingStation.MyPostalCode,
		MyRig:           qso.LoggingStation.MyRig,
		MySig:           qso.LoggingStation.MySig,
		MySigInfo:       qso.LoggingStation.MySigInfo,
		MyStreet:        qso.LoggingStation.MyStreet,
		MyWwffRef:       qso.LoggingStation.MyWwffRef,
		Operator:        qso.LoggingStation.Operator,
		OwnerCallsign:   qso.LoggingStation.OwnerCallsign,
		StationCallsign: qso.LoggingStation.StationCallsign,
		// QSL (TBD)
	}

	jsonData, err := json.Marshal(additionalData)
	if err != nil {
		return models.Qso{}, err
	}

	if len(jsonData) == 0 {
		jsonData = []byte("{}")
	}

	return models.Qso{
		ID:        qso.ID,
		LogbookID: qso.LogbookID,
		SessionID: qso.SessionID,

		// Field name matches between types and models
		Call:           qso.ContactedStation.Call,
		Band:           qso.QsoDetails.Band,
		Mode:           qso.QsoDetails.Mode,
		Freq:           freqHz,
		QsoDate:        date,
		TimeOn:         timeOn,
		TimeOff:        timeOff,
		RstSent:        qso.QsoDetails.RstSent,
		RstRcvd:        qso.QsoDetails.RstRcvd,
		Country:        qso.ContactedStation.Country,
		AdditionalData: jsonData,
	}, nil
}

func ContactedStationTypeToModel(station types.ContactedStation) (models.ContactedStation, error) {
	additionalData := struct {
		Address      string `json:"address,omitempty"`
		Age          string `json:"age,omitempty"`
		Altitude     string `json:"altitude,omitempty"`
		Cont         string `json:"cont,omitempty"` // the contacted station's Continent
		ContactedOp  string `json:"contacted_op,omitempty"`
		CQZ          string `json:"cqz,omitempty"`
		DXCC         string `json:"dxcc,omitempty"`
		Email        string `json:"email,omitempty"`
		EqCall       string `json:"eq_call,omitempty"` // the contacted station's owner's callsign (if different from call)
		Gridsquare   string `json:"gridsquare,omitempty"`
		Iota         string `json:"iota,omitempty"`
		IotaIslandId string `json:"iota_island_id,omitempty"`
		ITUZ         string `json:"ituz,omitempty"`
		Lat          string `json:"lat,omitempty"`
		Lon          string `json:"lon,omitempty"`
		//		Name         string `json:"name,omitempty"`
		QTH     string `json:"qth,omitempty"`
		Rig     string `json:"rig,omitempty"`
		Sig     string `json:"sig,omitempty"`      // the name of the contacted station's special activity or interest group
		SigInfo string `json:"sig_info,omitempty"` // information associated with the contacted station's activity or interest group
		Web     string `json:"web,omitempty"`
		WwffRef string `json:"wwff_ref,omitempty"`
	}{
		Address:      station.Address,
		Age:          station.Age,
		Altitude:     station.Altitude,
		Cont:         station.Cont,
		ContactedOp:  station.ContactedOp,
		CQZ:          station.CQZ,
		DXCC:         station.DXCC,
		Email:        station.Email,
		EqCall:       station.EqCall,
		Gridsquare:   station.Gridsquare,
		Iota:         station.Iota,
		IotaIslandId: station.IotaIslandId,
		ITUZ:         station.ITUZ,
		Lat:          station.Lat,
		Lon:          station.Lon,
		//		Name:         station.Name,
		QTH:     station.QTH,
		Rig:     station.Rig,
		Sig:     station.Sig,
		SigInfo: station.SigInfo,
		Web:     station.Web,
		WwffRef: station.WwffRef,
	}

	jsonData, err := json.Marshal(additionalData)
	if err != nil {
		return models.ContactedStation{}, err
	}

	if len(jsonData) == 0 {
		jsonData = []byte("{}")
	}

	return models.ContactedStation{
		ID:             station.ID,
		Call:           station.Call,
		Country:        station.Country,
		Name:           station.Name,
		AdditionalData: jsonData,
	}, nil
}

func CountryTypeToModel(country types.Country) (models.Country, error) {
	return models.Country{
		ID:         country.ID,
		Name:       country.Name,
		CQZone:     country.CQZone,
		ItuZone:    country.ITUZone,
		Continent:  country.Continent,
		Prefix:     country.Prefix,
		Ccode:      country.Ccode,
		DXCCPrefix: country.DXCCPrefix,
		TimeOffset: country.TimeOffset,
	}, nil
}
