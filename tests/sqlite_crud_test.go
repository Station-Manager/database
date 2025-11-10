package database

import (
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/database"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"path/filepath"
	"testing"
)

type TestSuite struct {
	suite.Suite
	//	typeQso types.Qso
	service database.Service
	qsoID   int64
}

func TestSqliteCrudSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
	fp, err := filepath.Abs("../../build/db/data.db")
	require.NoError(s.T(), err)

	cfg := types.AppConfig{
		DatastoreConfig: types.DatastoreConfig{
			Driver:                    database.SqliteDriver,
			Path:                      fp,
			MaxOpenConns:              1,
			MaxIdleConns:              1,
			ConnMaxLifetime:           1,
			ConnMaxIdleTime:           1,
			ContextTimeout:            5,
			TransactionContextTimeout: 5,
		},
	}
	cfgService := &config.Service{
		WorkingDir: "",
		AppConfig:  cfg,
	}
	err = cfgService.Initialize()
	require.NoError(s.T(), err)

	s.service = database.Service{
		ConfigService: cfgService,
	}
	err = s.service.Initialize()
	require.NoError(s.T(), err)

	err = s.service.Open()
	require.NoError(s.T(), err)
}

func (s *TestSuite) TestSqliteInsertQso() {
	typeQso := types.Qso{
		QsoDetails: types.QsoDetails{
			AIndex:      "",
			AntPath:     "",
			Band:        "20m",
			BandRx:      "",
			Comment:     "",
			ContestId:   "",
			Distance:    "",
			Freq:        "14.320",
			FreqRx:      "",
			Mode:        "SSB",
			Submode:     "USB",
			MySig:       "",
			MySigInfo:   "",
			Notes:       "",
			QsoDate:     "2025-11-08",
			QsoDateOff:  "2025-11-08",
			QsoRandom:   "",
			QsoComplete: "",
			RstRcvd:     "59",
			RstSent:     "56",
			RxPwr:       "",
			Sig:         "",
			SigInfo:     "",
			SRX:         "",
			STX:         "",
			TimeOff:     "11:46",
			TimeOn:      "11:40",
			TxPwr:       "500w",
		},
		ContactedStation: types.ContactedStation{
			Call:    "M0CMC",
			Country: "England",
		},
		LoggingStation: types.LoggingStation{
			StationCallsign: "7Q5MLV",
			MyCountry:       "Mzuzu",
			MyAntenna:       "VHQ Hex Beam",
		},
	}

	qso, err := s.service.InsertQso(typeQso)
	require.NoError(s.T(), err)
	assert.True(s.T(), qso.ID > 0)

	s.qsoID = qso.ID
	// Verify AdditionalData contains fields not in the model
	//ctx, cancel := s.service.withDefaultTimeout(nil)
	//defer cancel()
	//
	//var additionalData string
	//query := "SELECT additional_data FROM qso WHERE id = ?"
	//err = s.service.handle.QueryRowContext(ctx, query, qso.ID).Scan(&additionalData)
	//require.NoError(s.T(), err)
	//
	//// Parse the additional data
	//var additionalFields map[string]interface{}
	//err = json.Unmarshal([]byte(additionalData), &additionalFields)
	//require.NoError(s.T(), err)
	//
	//// Verify that fields not in the database model are stored in AdditionalData
	//assert.Contains(s.T(), additionalFields, "MyCountry", "MyCountry should be in AdditionalData")
	//assert.Equal(s.T(), "Mzuzu", additionalFields["MyCountry"], "MyCountry value should be 'Mzuzu'")
	//assert.Contains(s.T(), additionalFields, "MyAntenna", "MyAntenna should be in AdditionalData")
	//assert.Equal(s.T(), "VHQ Hex Beam", additionalFields["MyAntenna"], "MyAntenna value should be 'VHQ Hex Beam'")
}

func (s *TestSuite) TestSqliteFetchQso() {
	if s.qsoID == 0 {
		s.TestSqliteInsertQso()
	}

	typeQso, err := s.service.FetchQsoById(s.qsoID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.qsoID, typeQso.ID)
	assert.Equal(s.T(), "M0CMC", typeQso.ContactedStation.Call)
	assert.Equal(s.T(), "7Q5MLV", typeQso.LoggingStation.StationCallsign)
	assert.Equal(s.T(), "Mzuzu", typeQso.LoggingStation.MyCountry)
	assert.Equal(s.T(), "VHQ Hex Beam", typeQso.LoggingStation.MyAntenna)
}
