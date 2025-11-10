package database_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Station-Manager/config"
	"github.com/Station-Manager/database"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuitePG struct {
	suite.Suite
	typeQso   types.Qso
	service   database.Service
	logbookID int64
	qsoID     int64
}

func TestPsqlCrudSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL integration test in short mode")
	}
	suite.Run(t, new(TestSuitePG))
}

func (s *TestSuitePG) SetupSuite() {
	// Load password from .env if present
	pwd := "1q2w3e4r" // default fallback
	dotenvPath := filepath.Join("database", "postgres", ".env")
	if b, err := os.ReadFile(dotenvPath); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if key == "DB_PASSWORD" {
				pwd = val
				break
			}
		}
	}

	cfg := types.AppConfig{
		DatastoreConfig: types.DatastoreConfig{
			Driver:                    database.PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "station_manager",
			User:                      "smuser",
			Password:                  pwd,
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

	s.service = database.Service{
		ConfigService: cfgService,
	}
	err = s.service.Initialize()
	require.NoError(s.T(), err)

	err = s.service.Open()
	require.NoError(s.T(), err)

	// Run migrations to ensure tables exist
	err = s.service.Migrate()
	require.NoError(s.T(), err)

	// Create a test logbook for FK usage. Use a unique name to avoid conflicts.
	ctx := context.Background()
	name := "test_logbook_for_integration"
	callsign := "SMTEST"
	desc := "integration test logbook"
	_, err = s.service.ExecContext(ctx, "INSERT INTO logbook (name, callsign, description) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING", name, callsign, desc)
	require.NoError(s.T(), err)

	// Query its ID
	rows, err := s.service.QueryContext(ctx, "SELECT id FROM logbook WHERE name = $1", name)
	require.NoError(s.T(), err)
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	if rows.Next() {
		require.NoError(s.T(), rows.Scan(&s.logbookID))
	} else {
		require.FailNow(s.T(), "failed to ensure logbook for tests")
	}

	// Prepare a base QSO for insertion tests
	s.typeQso = types.Qso{
		QsoDetails: types.QsoDetails{
			Band:    "20m",
			Freq:    "14.320",
			Mode:    "SSB",
			QsoDate: "20251108",
			TimeOn:  "1140",
			TimeOff: "1146",
			RstRcvd: "59",
			RstSent: "56",
		},
		ContactedStation: types.ContactedStation{Call: "M0CMC", Country: "England"},
		LoggingStation:   types.LoggingStation{StationCallsign: "7Q5MLV", MyCountry: "Mzuzu", MyAntenna: "VHQ Hex Beam"},
		LogbookID:        s.logbookID,
	}
}

func (s *TestSuitePG) TearDownSuite() {
	_ = s.service.Close()
}

func (s *TestSuitePG) TestPsqlInsertQso() {
	if testing.Short() {
		s.T().Skip("Skipping PostgreSQL integration test in short mode")
	}
	qso, err := s.service.InsertQso(s.typeQso)
	require.NoError(s.T(), err)
	assert.True(s.T(), qso.ID > 0)
	// store for later tests
	s.qsoID = qso.ID
}

func (s *TestSuitePG) TestPsqlFetchQso() {
	if testing.Short() {
		s.T().Skip("Skipping PostgreSQL integration test in short mode")
	}
	// ensure insert ran
	if s.qsoID == 0 {
		s.T().Run("insert", func(t *testing.T) { s.TestPsqlInsertQso() })
	}
	tq, err := s.service.FetchQsoById(s.qsoID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.qsoID, tq.ID)
	assert.Equal(s.T(), "M0CMC", tq.ContactedStation.Call)
}

func (s *TestSuitePG) TestPsqlUpdateQso() {
	if testing.Short() {
		s.T().Skip("Skipping PostgreSQL integration test in short mode")
	}
	if s.qsoID == 0 {
		s.T().Run("insert", func(t *testing.T) { s.TestPsqlInsertQso() })
	}
	tq, err := s.service.FetchQsoById(s.qsoID)
	require.NoError(s.T(), err)
	// change call
	tq.ContactedStation.Call = "M1PG"
	require.NoError(s.T(), s.service.UpdateQso(tq))
	// re-fetch
	req, err := s.service.FetchQsoById(s.qsoID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "M1PG", req.ContactedStation.Call)
}

func (s *TestSuitePG) TestPsqlDeleteQso() {
	if testing.Short() {
		s.T().Skip("Skipping PostgreSQL integration test in short mode")
	}
	if s.qsoID == 0 {
		s.T().Run("insert", func(t *testing.T) { s.TestPsqlInsertQso() })
	}
	require.NoError(s.T(), s.service.DeleteQso(s.qsoID))
	_, err := s.service.FetchQsoById(s.qsoID)
	require.Error(s.T(), err)
}
