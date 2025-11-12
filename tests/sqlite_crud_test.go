package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/Station-Manager/config"
	"github.com/Station-Manager/database"
	"github.com/Station-Manager/logging"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
)

func setupSvcAndDefaultLogbook(t *testing.T) (*database.Service, int64) {
	// Temporary file-backed database (works like an in-memory DB for tests but is file-backed)
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	app := types.AppConfig{DatastoreConfig: types.DatastoreConfig{
		Driver:                    database.SqliteDriver,
		Path:                      dbPath,
		Options:                   map[string]string{},
		MaxOpenConns:              1,
		MaxIdleConns:              1,
		ConnMaxLifetime:           1,
		ConnMaxIdleTime:           1,
		ContextTimeout:            5,
		TransactionContextTimeout: 5,
	}, LoggingConfig: types.LoggingConfig{Level: "info", ConsoleLogging: true, FileLogging: false, RelLogFileDir: "logs"}}
	b, _ := json.MarshalIndent(app, "", "  ")
	_ = os.WriteFile(filepath.Join(tmpDir, "config.json"), b, 0o640)

	cfgSvc := &config.Service{WorkingDir: tmpDir, AppConfig: app}
	require.NoError(t, cfgSvc.Initialize())

	logSvc := &logging.Service{ConfigService: cfgSvc, WorkingDir: tmpDir}
	require.NoError(t, logSvc.Initialize())

	svc := &database.Service{ConfigService: cfgSvc, Logger: logSvc}
	require.NoError(t, svc.Initialize())
	require.NoError(t, svc.Open())
	// Ensure closed on test cleanup
	t.Cleanup(func() { _ = svc.Close() })

	// Run migrations to create tables
	require.NoError(t, svc.Migrate())

	ctx := context.Background()
	// Try find existing default logbook
	if rows, err := svc.QueryContext(ctx, "SELECT id FROM logbook WHERE name = ?", "default"); err == nil {
		defer func(rows *sql.Rows) {
			_ = rows.Close()
		}(rows)
		if rows.Next() {
			var id int64
			require.NoError(t, rows.Scan(&id))
			return svc, id
		}
	}
	// Insert a default logbook manually with required api_key in case the migration seed didn't include api_key
	_, err := svc.ExecContext(ctx, "INSERT INTO logbook (name, callsign, api_key, description) VALUES (?,?,?,?)", "default", "NOCALL", "DEFAULTKEY", "Default logbook created by test setup")
	require.NoError(t, err)

	rows2, err := svc.QueryContext(ctx, "SELECT id FROM logbook WHERE name = ?", "default")
	require.NoError(t, err)
	defer func(rows2 *sql.Rows) {
		_ = rows2.Close()
	}(rows2)
	var lbID int64
	require.True(t, rows2.Next(), "default logbook should exist")
	require.NoError(t, rows2.Scan(&lbID))
	return svc, lbID
}

// TestSqliteCRUDSequence creates a temporary SQLite database, runs migrations, then
// performs a sequence: insert QSO -> fetch QSO -> update QSO -> delete QSO. This is deliberately
// sequential to ensure tests that depend on prior operations run in order.
func TestSqliteCRUDSequence(t *testing.T) {
	svc, lbID := setupSvcAndDefaultLogbook(t)

	// Prepare a QSO using the default seeded logbook
	qsoIn := types.Qso{
		QsoDetails: types.QsoDetails{
			Band:    "20m",
			Freq:    "14.320",
			Mode:    "SSB",
			QsoDate: "20251108", // YYYYMMDD format expected by schema
			TimeOn:  "1140",
			TimeOff: "1146",
			RstRcvd: "59",
			RstSent: "56",
		},
		ContactedStation: types.ContactedStation{Call: "M0CMC", Country: "England"},
		LoggingStation:   types.LoggingStation{StationCallsign: "7Q5MLV", MyCountry: "Mzuzu", MyAntenna: "VHQ Hex Beam"},
		LogbookID:        lbID,
	}

	// Insert
	qsoOut, err := svc.InsertQso(qsoIn)
	require.NoError(t, err)
	assert.Greater(t, qsoOut.ID, int64(0))

	// Fetch
	typQso, err := svc.FetchQsoById(qsoOut.ID)
	require.NoError(t, err)
	assert.Equal(t, qsoOut.ID, typQso.ID)
	assert.Equal(t, "M0CMC", typQso.ContactedStation.Call)
	assert.Equal(t, "7Q5MLV", typQso.LoggingStation.StationCallsign)
	assert.Equal(t, "Mzuzu", typQso.LoggingStation.MyCountry)
	assert.Equal(t, "VHQ Hex Beam", typQso.LoggingStation.MyAntenna)

	// Update: change contacted call and persist
	typQso.ContactedStation.Call = "M1NEW"
	require.NoError(t, svc.UpdateQso(typQso))

	// Re-fetch and verify update
	refetched, err := svc.FetchQsoById(qsoOut.ID)
	require.NoError(t, err)
	assert.Equal(t, "M1NEW", refetched.ContactedStation.Call)

	// Delete
	require.NoError(t, svc.DeleteQso(qsoOut.ID))

	// Verify deletion returns an error when fetching
	_, err = svc.FetchQsoById(qsoOut.ID)
	require.Error(t, err)
}

// Test update with invalid ID should return an error
func TestSqliteUpdateInvalidID(t *testing.T) {
	svc, lbID := setupSvcAndDefaultLogbook(t)

	// Prepare and insert a QSO to have a valid row
	qsoIn := types.Qso{
		QsoDetails:       types.QsoDetails{Band: "20m", Freq: "14.320", Mode: "SSB", QsoDate: "20251108", TimeOn: "1140", TimeOff: "1146", RstRcvd: "59", RstSent: "56"},
		ContactedStation: types.ContactedStation{Call: "M0CMC", Country: "England"},
		LoggingStation:   types.LoggingStation{StationCallsign: "7Q5MLV"},
		LogbookID:        lbID,
	}

	qsoOut, err := svc.InsertQso(qsoIn)
	require.NoError(t, err)
	assert.Greater(t, qsoOut.ID, int64(0))

	// Attempt to update a non-existent ID
	bad := qsoOut
	bad.ID = qsoOut.ID + 9999
	bad.ContactedStation.Call = "BADD"
	err = svc.UpdateQso(bad)
	require.Error(t, err)

	// Cleanup
	require.NoError(t, svc.DeleteQso(qsoOut.ID))
}

// Test updating a QSO to reference a non-existent logbook should fail due to FK
func TestSqliteUpdateForeignKeyViolation(t *testing.T) {
	svc, lbID := setupSvcAndDefaultLogbook(t)

	qsoIn := types.Qso{
		QsoDetails:       types.QsoDetails{Band: "20m", Freq: "14.320", Mode: "SSB", QsoDate: "20251108", TimeOn: "1140", TimeOff: "1146", RstRcvd: "59", RstSent: "56"},
		ContactedStation: types.ContactedStation{Call: "M0CMC", Country: "England"},
		LoggingStation:   types.LoggingStation{StationCallsign: "7Q5MLV"},
		LogbookID:        lbID,
	}

	qsoOut, err := svc.InsertQso(qsoIn)
	require.NoError(t, err)
	assert.Greater(t, qsoOut.ID, int64(0))

	// Change logbook_id to a non-existent value and attempt update
	qsoOut.LogbookID = 999999
	err = svc.UpdateQso(qsoOut)
	require.Error(t, err)

	// Cleanup: set back original and delete
	qsoOut.LogbookID = lbID
	require.NoError(t, svc.UpdateQso(qsoOut))
	require.NoError(t, svc.DeleteQso(qsoOut.ID))
}
