package database_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Station-Manager/config"
	"github.com/Station-Manager/database"
	"github.com/Station-Manager/logging"
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
			ContextTimeout:            30, // increased for integration test
			TransactionContextTimeout: 30, // increased for integration test
		},
		LoggingConfig: types.LoggingConfig{
			Level:          "info",
			WithTimestamp:  false,
			ConsoleLogging: true,
			FileLogging:    false,
			RelLogFileDir:  "logs",
		},
	}
	cfgService := &config.Service{
		WorkingDir: "",
		AppConfig:  cfg,
	}
	err := cfgService.Initialize()
	require.NoError(s.T(), err)

	// Initialize logging service for database
	logSvc := &logging.Service{ConfigService: cfgService, WorkingDir: s.T().TempDir()}
	require.NoError(s.T(), logSvc.Initialize())

	s.service = database.Service{
		ConfigService: cfgService,
		Logger:        logSvc,
	}
	err = s.service.Initialize()
	require.NoError(s.T(), err)

	err = s.service.Open()
	if err != nil { // Skip if database unavailable
		s.T().Skip("Postgres not available: skipping suite")
	}

	// Run migrations to ensure tables exist
	err = s.service.Migrate()
	if err != nil {
		s.T().Skip("Migrations failed; skipping suite")
	}

	// Post-migration ping to verify connection usability
	if pingErr := s.service.Ping(); pingErr != nil {
		// Attempt one reopen cycle
		_ = s.service.Close()
		if reopenErr := s.service.Open(); reopenErr != nil {
			s.T().Skip("Post-migration reopen failed: " + reopenErr.Error())
		}
		if pingErr2 := s.service.Ping(); pingErr2 != nil {
			s.T().Skip("Post-migration ping still failing: " + pingErr2.Error())
		}
	}

	// Simple probe to detect long-running locks
	probeCtx, probeCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer probeCancel()
	_, probeErr := s.service.QueryContext(probeCtx, "SELECT 1")
	if probeErr != nil {
		s.T().Skip("Probe SELECT 1 failed: " + probeErr.Error())
	}

	// Create a test logbook for FK usage. Use a unique name to avoid conflicts.
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	name := "test_logbook_it"
	callsign := "SMTEST"
	desc := "integration test logbook"
	apiKey := "PGKEY-" + time.Now().Format("20060102-150405.000")
	_, err = s.service.ExecContext(ctx, "INSERT INTO logbook (name, callsign, api_key, description) VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO NOTHING", name, callsign, apiKey, desc)
	if err != nil {
		// Fail early rather than silently skipping so underlying issue is visible
		s.T().Fatal("Postgres insert failed: " + err.Error())
	}

	// Query its ID
	rows, err := s.service.QueryContext(ctx, "SELECT id FROM logbook WHERE name = $1", name)
	if err != nil {
		s.T().Fatal("Postgres query failed: " + err.Error())
	}
	defer func(rows *sql.Rows) { _ = rows.Close() }(rows)
	if rows.Next() {
		require.NoError(s.T(), rows.Scan(&s.logbookID))
	} else {
		s.T().Fatal("failed to ensure logbook for tests")
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
	// cleanup
	_ = s.service.DeleteQso(qso.ID)
}

func (s *TestSuitePG) TestPsqlFetchQso() {
	if testing.Short() {
		s.T().Skip("Skipping PostgreSQL integration test in short mode")
	}
	// Insert a QSO to fetch
	qso, err := s.service.InsertQso(s.typeQso)
	require.NoError(s.T(), err)
	defer func() { _ = s.service.DeleteQso(qso.ID) }()

	tq, err := s.service.FetchQsoById(qso.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), qso.ID, tq.ID)
	assert.Equal(s.T(), "M0CMC", tq.ContactedStation.Call)
}

func (s *TestSuitePG) TestPsqlUpdateQso() {
	if testing.Short() {
		s.T().Skip("Skipping PostgreSQL integration test in short mode")
	}
	// Insert baseline
	qso, err := s.service.InsertQso(s.typeQso)
	require.NoError(s.T(), err)
	defer func() { _ = s.service.DeleteQso(qso.ID) }()

	tq, err := s.service.FetchQsoById(qso.ID)
	require.NoError(s.T(), err)
	// change call
	tq.ContactedStation.Call = "M1PG"
	require.NoError(s.T(), s.service.UpdateQso(tq))
	// re-fetch
	req, err := s.service.FetchQsoById(qso.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "M1PG", req.ContactedStation.Call)
}

func (s *TestSuitePG) TestPsqlDeleteQso() {
	if testing.Short() {
		s.T().Skip("Skipping PostgreSQL integration test in short mode")
	}
	// Insert then delete
	qso, err := s.service.InsertQso(s.typeQso)
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.service.DeleteQso(qso.ID))
	_, err = s.service.FetchQsoById(qso.ID)
	require.Error(s.T(), err)
}
