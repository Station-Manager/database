package database

import (
	"context"
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// Test helper functions

func getValidPostgresConfig() *types.DatastoreConfig {
	return &types.DatastoreConfig{
		Driver:                    PostgresDriver,
		Host:                      "localhost",
		Port:                      5432,
		Database:                  "testdb",
		User:                      "testuser",
		Password:                  "testpass",
		SSLMode:                   "disable",
		MaxOpenConns:              10,
		MaxIdleConns:              5,
		ConnMaxLifetime:           15,
		ConnMaxIdleTime:           5,
		ContextTimeout:            5,
		TransactionContextTimeout: 10,
	}
}

func getValidSqliteConfig(t *testing.T) *types.DatastoreConfig {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	t.Cleanup(func() {
		os.Remove(dbPath)
	})
	return &types.DatastoreConfig{
		Driver:                    SqliteDriver,
		Path:                      dbPath,
		Options:                   "",
		Host:                      "localhost",
		Port:                      1,
		User:                      "test",
		Password:                  "test",
		Database:                  "test",
		SSLMode:                   "disable",
		MaxOpenConns:              10,
		MaxIdleConns:              5,
		ConnMaxLifetime:           15,
		ConnMaxIdleTime:           5,
		ContextTimeout:            5,
		TransactionContextTimeout: 10,
	}
}

func getServiceWithPostgresConfig() *Service {
	cfg := getValidPostgresConfig()
	cfgSvc := &config.Service{
		AppConfig: types.AppConfig{
			DatastoreConfig: *cfg,
		},
	}
	_ = cfgSvc.Initialize()
	return &Service{ConfigService: cfgSvc}
}

func getServiceWithSqliteConfig(t *testing.T) *Service {
	cfg := getValidSqliteConfig(t)
	cfgSvc := &config.Service{
		AppConfig: types.AppConfig{
			DatastoreConfig: *cfg,
		},
	}
	_ = cfgSvc.Initialize()
	return &Service{ConfigService: cfgSvc}
}

func getServiceWithConfig(cfg *types.DatastoreConfig) *Service {
	cfgSvc := &config.Service{
		AppConfig: types.AppConfig{
			DatastoreConfig: *cfg,
		},
	}
	_ = cfgSvc.Initialize()
	return &Service{ConfigService: cfgSvc}
}

// TestService_Initialize tests the Initialize method
func TestService_Initialize(t *testing.T) {
	t.Run("nil service returns error", func(t *testing.T) {
		var svc *Service
		err := svc.Initialize()
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNilService)
	})

	t.Run("nil config service returns error", func(t *testing.T) {
		svc := &Service{}
		err := svc.Initialize()
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgAppConfigNil)
	})

	t.Run("empty config fails validation", func(t *testing.T) {
		svc := getServiceWithConfig(&types.DatastoreConfig{})
		err := svc.Initialize()
		require.Error(t, err)
		assert.Error(t, err)
	})

	t.Run("valid postgres config succeeds", func(t *testing.T) {
		svc := getServiceWithPostgresConfig()
		err := svc.Initialize()
		assert.NoError(t, err)
		assert.True(t, svc.isInitialized.Load())
	})

	t.Run("valid sqlite config succeeds", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		err := svc.Initialize()
		assert.NoError(t, err)
		assert.True(t, svc.isInitialized.Load())
	})

	t.Run("invalid driver fails validation", func(t *testing.T) {
		cfg := getValidPostgresConfig()
		cfg.Driver = "mysql"
		svc := getServiceWithConfig(cfg)
		err := svc.Initialize()
		require.Error(t, err)
		assert.Error(t, err)
	})

	t.Run("double initialization succeeds", func(t *testing.T) {
		svc := getServiceWithPostgresConfig()
		err := svc.Initialize()
		require.NoError(t, err)
		err = svc.Initialize()
		assert.NoError(t, err)
	})
}

// TestService_Open tests the Open method
func TestService_Open(t *testing.T) {
	t.Run("nil service returns error", func(t *testing.T) {
		var svc *Service
		err := svc.Open()
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNilService)
	})

	t.Run("uninitialized service returns error", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		err := svc.Open()
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNotInitialized)
	})

	t.Run("valid sqlite config opens successfully", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())

		err := svc.Open()
		assert.NoError(t, err)
		assert.True(t, svc.isOpen.Load())
		assert.NotNil(t, svc.handle)

		// Cleanup
		_ = svc.Close()
	})

	t.Run("double open returns error", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())

		err := svc.Open()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errMsgAlreadyOpen)

		// Cleanup
		_ = svc.Close()
	})

	t.Run("concurrent open attempts are safe", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())

		var wg sync.WaitGroup
		successCount := 0
		errorCount := 0
		var mu sync.Mutex

		// Attempt to open 10 times concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := svc.Open()
				mu.Lock()
				if err != nil {
					errorCount++
				} else {
					successCount++
				}
				mu.Unlock()
			}()
		}

		wg.Wait()

		// Exactly one should succeed
		assert.Equal(t, 1, successCount, "exactly one Open() should succeed")
		assert.Equal(t, 9, errorCount, "nine Open() calls should fail")
		assert.True(t, svc.isOpen.Load())

		// Cleanup
		_ = svc.Close()
	})
}

// TestService_Close tests the Close method
func TestService_Close(t *testing.T) {
	t.Run("nil service returns error", func(t *testing.T) {
		var svc *Service
		err := svc.Close()
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNilService)
	})

	t.Run("closing unopened service is idempotent", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		err := svc.Close()
		// Close is idempotent, should not error
		assert.NoError(t, err)
	})

	t.Run("close after open succeeds", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())

		err := svc.Close()
		assert.NoError(t, err)
		assert.False(t, svc.isOpen.Load())
		assert.Nil(t, svc.handle)
	})

	t.Run("double close is idempotent", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		require.NoError(t, svc.Close())

		// Close is idempotent, should not error on second close
		err := svc.Close()
		assert.NoError(t, err)
	})

	t.Run("concurrent close attempts are safe", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())

		var wg sync.WaitGroup

		// Attempt to close 10 times concurrently
		// Since Close is idempotent, all should succeed without error
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := svc.Close()
				// Close is idempotent, should never error
				assert.NoError(t, err)
			}()
		}

		wg.Wait()

		// Database should be closed
		assert.False(t, svc.isOpen.Load())
	})
}

// TestService_Ping tests the Ping method
func TestService_Ping(t *testing.T) {
	t.Run("nil service returns error", func(t *testing.T) {
		var svc *Service
		err := svc.Ping()
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNilService)
	})

	t.Run("ping unopened service returns error", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		err := svc.Ping()
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNotOpen)
	})

	t.Run("ping open service succeeds", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		err := svc.Ping()
		assert.NoError(t, err)
	})

	t.Run("concurrent pings are safe", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := svc.Ping()
				assert.NoError(t, err)
			}()
		}

		wg.Wait()
	})
}

// TestService_BeginTxContext tests the BeginTxContext method
func TestService_BeginTxContext(t *testing.T) {
	t.Run("nil service returns error", func(t *testing.T) {
		var svc *Service
		tx, cancel, err := svc.BeginTxContext(context.Background())
		assert.Nil(t, tx)
		assert.Nil(t, cancel)
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNilService)
	})

	t.Run("unopened service returns error", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		tx, cancel, err := svc.BeginTxContext(context.Background())
		assert.Nil(t, tx)
		assert.Nil(t, cancel)
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNotOpen)
	})

	t.Run("begin transaction succeeds", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		tx, cancel, err := svc.BeginTxContext(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, tx)
		assert.NotNil(t, cancel)

		// Cleanup
		_ = tx.Rollback()
		cancel()
	})

	t.Run("context without deadline gets timeout", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		ctx := context.Background()
		tx, cancel, err := svc.BeginTxContext(ctx)
		require.NoError(t, err)
		assert.NotNil(t, cancel)

		// Cleanup
		_ = tx.Rollback()
		cancel()
	})

	t.Run("context with deadline is preserved", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer ctxCancel()

		tx, cancel, err := svc.BeginTxContext(ctx)
		require.NoError(t, err)
		assert.NotNil(t, cancel)

		// Cleanup
		_ = tx.Rollback()
		cancel()
	})

	t.Run("multiple concurrent transactions", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				tx, cancel, err := svc.BeginTxContext(context.Background())
				assert.NoError(t, err)
				assert.NotNil(t, tx)
				_ = tx.Rollback()
				cancel()
			}()
		}

		wg.Wait()
	})
}

// TestService_ExecContext tests the ExecContext method
func TestService_ExecContext(t *testing.T) {
	t.Run("nil service returns error", func(t *testing.T) {
		var svc *Service
		res, err := svc.ExecContext(context.Background(), "SELECT 1")
		assert.Nil(t, res)
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNilService)
	})

	t.Run("unopened service returns error", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		res, err := svc.ExecContext(context.Background(), "SELECT 1")
		assert.Nil(t, res)
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNotOpen)
	})

	t.Run("create table succeeds", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		res, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("insert and get affected rows", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
		require.NoError(t, err)

		res, err := svc.ExecContext(context.Background(), "INSERT INTO test (name) VALUES (?)", "Alice")
		require.NoError(t, err)

		affected, err := res.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)
	})

	t.Run("context with deadline is respected", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		res, err := svc.ExecContext(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY)")
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("invalid SQL returns error", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		res, err := svc.ExecContext(context.Background(), "INVALID SQL")
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

// TestService_QueryContext tests the QueryContext method
func TestService_QueryContext(t *testing.T) {
	t.Run("nil service returns error", func(t *testing.T) {
		var svc *Service
		rows, err := svc.QueryContext(context.Background(), "SELECT 1")
		assert.Nil(t, rows)
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNilService)
	})

	t.Run("unopened service returns error", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		rows, err := svc.QueryContext(context.Background(), "SELECT 1")
		assert.Nil(t, rows)
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNotOpen)
	})

	t.Run("simple query succeeds", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		rows, err := svc.QueryContext(context.Background(), "SELECT 1")
		require.NoError(t, err)
		assert.NotNil(t, rows)
		defer rows.Close()

		assert.True(t, rows.Next())
		var result int
		err = rows.Scan(&result)
		assert.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("query with parameters", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
		require.NoError(t, err)
		_, err = svc.ExecContext(context.Background(), "INSERT INTO test (name) VALUES (?), (?)", "Alice", "Bob")
		require.NoError(t, err)

		rows, err := svc.QueryContext(context.Background(), "SELECT name FROM test WHERE name = ?", "Alice")
		require.NoError(t, err)
		defer rows.Close()

		assert.True(t, rows.Next())
		var name string
		err = rows.Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, "Alice", name)
	})

	t.Run("context with deadline is respected", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		rows, err := svc.QueryContext(ctx, "SELECT 1")
		assert.NoError(t, err)
		assert.NotNil(t, rows)
		rows.Close()
	})

	t.Run("invalid SQL returns error", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		rows, err := svc.QueryContext(context.Background(), "INVALID SQL")
		assert.Error(t, err)
		assert.Nil(t, rows)
	})
}

// TestService_Lifecycle tests the full lifecycle
func TestService_Lifecycle(t *testing.T) {
	t.Run("full lifecycle with operations", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)

		// Initialize
		err := svc.Initialize()
		require.NoError(t, err)
		assert.True(t, svc.isInitialized.Load())

		// Open
		err = svc.Open()
		require.NoError(t, err)
		assert.True(t, svc.isOpen.Load())

		// Ping
		err = svc.Ping()
		assert.NoError(t, err)

		// Create table
		_, err = svc.ExecContext(context.Background(), "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
		assert.NoError(t, err)

		// Insert data
		_, err = svc.ExecContext(context.Background(), "INSERT INTO users (name) VALUES (?)", "John")
		assert.NoError(t, err)

		// Query data
		rows, err := svc.QueryContext(context.Background(), "SELECT name FROM users")
		require.NoError(t, err)
		assert.True(t, rows.Next())
		var name string
		err = rows.Scan(&name)
		assert.NoError(t, err)
		assert.Equal(t, "John", name)
		rows.Close()

		// Transaction
		tx, cancel, err := svc.BeginTxContext(context.Background())
		require.NoError(t, err)
		_, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "Jane")
		assert.NoError(t, err)
		err = tx.Commit()
		assert.NoError(t, err)
		cancel()

		// Close
		err = svc.Close()
		assert.NoError(t, err)
		assert.False(t, svc.isOpen.Load())
	})

	t.Run("reopen after close", func(t *testing.T) {
		svc := getServiceWithSqliteConfig(t)
		require.NoError(t, svc.Initialize())

		// Open, close, reopen
		require.NoError(t, svc.Open())
		require.NoError(t, svc.Close())
		err := svc.Open()
		assert.NoError(t, err)
		assert.True(t, svc.isOpen.Load())

		_ = svc.Close()
	})
}
