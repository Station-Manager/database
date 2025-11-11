package database

import (
	"context"
	"testing"
	"time"

	"github.com/Station-Manager/config"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to build a sqlite service quickly
func newSqliteService(t *testing.T) *Service {
	t.Helper()
	cfg := types.DatastoreConfig{
		Driver:                    SqliteDriver,
		Path:                      t.TempDir() + "/ctx_crud.db",
		Options:                   map[string]string{},
		MaxOpenConns:              2,
		MaxIdleConns:              2,
		ConnMaxLifetime:           5,
		ConnMaxIdleTime:           5,
		ContextTimeout:            5,
		TransactionContextTimeout: 5,
		Host:                      "localhost",
		Port:                      1,
		User:                      "u",
		Password:                  "p",
		Database:                  "d",
		SSLMode:                   "disable",
	}
	appCfg := types.AppConfig{DatastoreConfig: cfg}
	cfgSvc := &config.Service{AppConfig: appCfg}
	require.NoError(t, cfgSvc.Initialize())
	svc := &Service{ConfigService: cfgSvc}
	require.NoError(t, svc.Initialize())
	require.NoError(t, svc.Open())
	t.Cleanup(func() { _ = svc.Close() })
	require.NoError(t, svc.Migrate())
	return svc
}

func TestContextCRUD_InsertFetchUpdateDelete(t *testing.T) {
	svc := newSqliteService(t)
	// Create logbook to satisfy FK (manually include api_key)
	ctx := context.Background()
	_, err := svc.ExecContext(ctx, "INSERT INTO logbook (name, callsign, api_key, description) VALUES (?,?,?,?)", "ctxlb", "CALL", "CTXKEY1", "descr")
	require.NoError(t, err)
	rows, err := svc.QueryContext(ctx, "SELECT id FROM logbook WHERE name = ?", "ctxlb")
	require.NoError(t, err)
	defer rows.Close()
	var lbID int64
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&lbID))

	base := types.Qso{
		QsoDetails:       types.QsoDetails{Band: "20m", Freq: "14.320", Mode: "SSB", QsoDate: "20251110", TimeOn: "1010", TimeOff: "1015", RstRcvd: "59", RstSent: "59"},
		ContactedStation: types.ContactedStation{Call: "M0CTX"},
		LoggingStation:   types.LoggingStation{StationCallsign: "STN", MyCountry: "Wonderland"},
		LogbookID:        lbID,
	}

	ins, err := svc.InsertQsoContext(ctx, base)
	if err != nil {
		if de, ok := errors.AsDetailedError(err); ok && de.Cause() != nil {
			t.Fatalf("insert failed: %s | cause: %v", de.Error(), de.Cause())
		}
		require.NoError(t, err)
	}
	assert.True(t, ins.ID > 0)

	fetched, err := svc.FetchQsoByIdContext(ctx, ins.ID)
	require.NoError(t, err)
	assert.Equal(t, "M0CTX", fetched.ContactedStation.Call)

	fetched.ContactedStation.Call = "M1CTX"
	require.NoError(t, svc.UpdateQsoContext(ctx, fetched))

	refetched, err := svc.FetchQsoByIdContext(ctx, fetched.ID)
	require.NoError(t, err)
	assert.Equal(t, "M1CTX", refetched.ContactedStation.Call)

	require.NoError(t, svc.DeleteQsoContext(ctx, fetched.ID))
	_, err = svc.FetchQsoByIdContext(ctx, fetched.ID)
	require.Error(t, err)
}

func TestContextCRUD_CancelledContext(t *testing.T) {
	svc := newSqliteService(t)
	ctx := context.Background()
	_, err := svc.ExecContext(ctx, "INSERT INTO logbook (name, callsign, api_key, description) VALUES (?,?,?,?)", "ctxlb2", "CALL2", "CTXKEY2", "descr")
	require.NoError(t, err)
	rows, err := svc.QueryContext(ctx, "SELECT id FROM logbook WHERE name = ?", "ctxlb2")
	require.NoError(t, err)
	defer rows.Close()
	var lbID int64
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&lbID))

	base := types.Qso{
		QsoDetails:       types.QsoDetails{Band: "40m", Freq: "7.050", Mode: "CW", QsoDate: "20251110", TimeOn: "1110", TimeOff: "1112", RstRcvd: "579", RstSent: "579"},
		ContactedStation: types.ContactedStation{Call: "M0CXL"},
		LoggingStation:   types.LoggingStation{StationCallsign: "STN", MyCountry: "Wonderland"},
		LogbookID:        lbID,
	}

	cctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	_, err = svc.InsertQsoContext(cctx, base)
	require.Error(t, err, "expected error due to cancelled context")
}

func TestContextCRUD_TimeoutApplied(t *testing.T) {
	svc := newSqliteService(t)
	ctx := context.Background()
	_, err := svc.ExecContext(ctx, "INSERT INTO logbook (name, callsign, api_key, description) VALUES (?,?,?,?)", "ctxlb3", "CALL3", "CTXKEY3", "descr")
	require.NoError(t, err)
	rows, err := svc.QueryContext(ctx, "SELECT id FROM logbook WHERE name = ?", "ctxlb3")
	require.NoError(t, err)
	defer rows.Close()
	var lbID int64
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&lbID))

	base := types.Qso{
		QsoDetails:       types.QsoDetails{Band: "10m", Freq: "28.400", Mode: "SSB", QsoDate: "20251110", TimeOn: "1210", TimeOff: "1212", RstRcvd: "59", RstSent: "59"},
		ContactedStation: types.ContactedStation{Call: "M0TM"},
		LoggingStation:   types.LoggingStation{StationCallsign: "STN", MyCountry: "Wonderland"},
		LogbookID:        lbID,
	}

	// Provide a context without deadline; service should apply default timeout and still succeed
	q, err := svc.InsertQsoContext(context.Background(), base)
	require.NoError(t, err)
	assert.True(t, q.ID > 0)

	// Use explicit short timeout for fetch (should be fine)
	shortCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = svc.FetchQsoByIdContext(shortCtx, q.ID)
	require.NoError(t, err)
}
