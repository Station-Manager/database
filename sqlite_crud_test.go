package database

import (
	"encoding/json"
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestSqliteCRUD(t *testing.T) {
	t.Run("SQLite InserQSO test", func(t *testing.T) {
		qso := types.Qso{
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

		fp, err := filepath.Abs("../build/db/data.db")
		if err != nil {
			panic(err)
		}
		cfg := types.AppConfig{
			DatastoreConfig: types.DatastoreConfig{
				Driver:                    SqliteDriver,
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
		require.NoError(t, err)

		dbService := Service{
			ConfigService: cfgService,
		}
		err = dbService.Initialize()
		require.NoError(t, err)

		err = dbService.Open()
		require.NoError(t, err)

		err = dbService.Migrate()
		require.NoError(t, err)

		qso, err = dbService.InsertQso(qso)
		require.NoError(t, err)
		assert.True(t, qso.ID > 0)

		// Query the database to verify AdditionalData
		ctx, cancel := dbService.withDefaultTimeout(nil)
		defer cancel()

		var additionalData string
		query := "SELECT additional_data FROM qso WHERE id = ?"
		err = dbService.handle.QueryRowContext(ctx, query, qso.ID).Scan(&additionalData)
		require.NoError(t, err)

		//		t.Logf("AdditionalData from DB: %s", additionalData)

		// Parse and verify the additional data contains MyCountry
		var additionalFields map[string]interface{}
		err = json.Unmarshal([]byte(additionalData), &additionalFields)
		require.NoError(t, err)

		//		t.Logf("Parsed AdditionalData: %+v", additionalFields)
		require.Contains(t, additionalFields, "MyCountry", "MyCountry should be in AdditionalData")
		require.Equal(t, "Mzuzu", additionalFields["MyCountry"], "MyCountry value should be 'Mzuzu'")

		err = dbService.Close()
		require.NoError(t, err)
	})
}
