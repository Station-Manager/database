package adapters

import (
	"github.com/Station-Manager/database/sqlite/models"
	"github.com/Station-Manager/types"
	"github.com/aarondl/null/v8"
	"strconv"
)

func QsoTypeToSqliteModel(qso types.Qso) models.Qso {
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

	// Country conversion: types have string, model expects null.String
	var country null.String
	if qso.ContactedStation.Country != "" {
		country = null.StringFrom(qso.ContactedStation.Country)
	}

	additionalData := struct {
		AIndex      string `json:"a_index"`
		AntPath     string `json:"ant_path"` // ADIF, section II.B.1 - currently, we only use S and L
		BandRx      string `json:"band_rx"`  //in a split frequency QSO, the logging station's receiving band
		Comment     string `json:"comment"`
		ContestId   string `json:"contest_id"`
		Distance    string `json:"distance"` // km
		FreqRx      string `json:"freq_rx"`
		Submode     string `json:"submode"`
		Notes       string `json:"notes"` // information of interest to the logging station's operator
		QsoDate     string `json:"qso_date" validate:"required"`
		QsoDateOff  string `json:"qso_date_off"`
		QsoRandom   string `json:"qso_random"`
		QsoComplete string `json:"qso_complete"`
		RstRcvd     string `json:"rst_rcvd"`
		RstSent     string `json:"rst_sent"`
		RxPwr       string `json:"rx_pwr"` // the contacted station's transmitter power in Watts with a value greater than or equal to 0
		SRX         string `json:"srx"`    // contest QSO received serial number with a value greater than or equal to 0
		STX         string `json:"stx"`    // contest QSO transmitted serial number with a value greater than or equal to 0
		TimeOff     string `json:"time_off"`
		TimeOn      string `json:"time_on"`
		TxPwr       string `json:"tx_pwr"` // the logging station's power in Watts with a value greater than or equal to 0
	}{}

	return models.Qso{
		ID:        qso.ID,
		LogbookID: qso.LogbookID,
		SessionID: qso.SessionID,

		// Field name matches between types and models
		Call:    qso.ContactedStation.Call,
		Band:    qso.QsoDetails.Band,
		Mode:    qso.QsoDetails.Mode,
		Freq:    freqHz,
		QsoDate: qso.QsoDetails.QsoDate,
		TimeOn:  qso.QsoDetails.TimeOn,
		TimeOff: qso.QsoDetails.TimeOff,
		RstSent: qso.QsoDetails.RstSent,
		RstRcvd: qso.QsoDetails.RstRcvd,
		Country: country,

		// AdditionalData not present on types.Qso; leave zero value
	}

}
