package database

import (
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestSuitePG struct {
	suite.Suite
	typeQso types.Qso
	service Service
}

func TestPsqlCrudSuite(t *testing.T) {
	suite.Run(t, new(TestSuitePG))
}

func (s *TestSuitePG) SetupSuite() {
	cfg := types.AppConfig{
		DatastoreConfig: types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "station_manager",
			User:                      "smuser",
			Password:                  "1q2w3e4r",
			SSLMode:                   "disable",
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
	err := cfgService.Initialize()
	require.NoError(s.T(), err)

	s.service = Service{
		ConfigService: cfgService,
	}
	err = s.service.Initialize()
	require.NoError(s.T(), err)

	err = s.service.Open()
	require.NoError(s.T(), err)

	s.typeQso = types.Qso{
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
}

func (s *TestSuitePG) TestPsqlInsertQso() {
	qso, err := s.service.InsertQso(s.typeQso)
	require.NoError(s.T(), err)
	assert.True(s.T(), qso.ID > 0)
}
