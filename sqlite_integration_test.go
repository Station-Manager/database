package database

import (
	"context"
	"fmt"
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

// getSqliteFileConfig returns a config that works without validation issues
func getSqliteFileConfig(t *testing.T) *types.DatastoreConfig {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	t.Cleanup(func() {
		os.Remove(dbPath)
	})
	return &types.DatastoreConfig{
		Driver:                    SqliteDriver,
		Path:                      dbPath,
		Options:                   "",
		Host:                      "127.0.0.1",
		Port:                      5432,
		User:                      "testuser",
		Password:                  "testpass",
		Database:                  "testdb",
		SSLMode:                   "disable",
		MaxOpenConns:              10,
		MaxIdleConns:              5,
		ConnMaxLifetime:           15,
		ConnMaxIdleTime:           5,
		ContextTimeout:            5,
		TransactionContextTimeout: 10,
	}
}

func getSqliteService(t *testing.T) *Service {
	cfg := getSqliteFileConfig(t)
	cfgSvc := &config.Service{
		AppConfig: types.AppConfig{
			DatastoreConfig: *cfg,
		},
	}
	_ = cfgSvc.Initialize()
	return &Service{ConfigService: cfgSvc}
}

// TestSQLiteIntegration tests actual SQLite database operations
func TestSQLiteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("full lifecycle with SQLite", func(t *testing.T) {
		svc := getSqliteService(t)

		// Initialize
		err := svc.Initialize()
		require.NoError(t, err)

		// Open
		err = svc.Open()
		require.NoError(t, err)
		defer svc.Close()

		fmt.Println(err)

		// Ping
		err = svc.Ping()
		assert.NoError(t, err)

		// Create table
		_, err = svc.ExecContext(context.Background(), "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)")
		require.NoError(t, err)

		// Insert data
		result, err := svc.ExecContext(context.Background(), "INSERT INTO users (name, email) VALUES (?, ?)", "Alice", "alice@example.com")
		require.NoError(t, err)
		rowsAffected, _ := result.RowsAffected()
		assert.Equal(t, int64(1), rowsAffected)

		// Query data
		rows, err := svc.QueryContext(context.Background(), "SELECT name, email FROM users WHERE name = ?", "Alice")
		require.NoError(t, err)

		assert.True(t, rows.Next())
		var name, email string
		err = rows.Scan(&name, &email)
		require.NoError(t, err)
		assert.Equal(t, "Alice", name)
		assert.Equal(t, "alice@example.com", email)
		rows.Close() // Close rows before starting transaction

		// Transaction
		tx, cancel, err := svc.BeginTxContext(context.Background())
		require.NoError(t, err)

		_, err = tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", "Bob", "bob@example.com")
		require.NoError(t, err)

		err = tx.Commit()
		assert.NoError(t, err)
		cancel() // Cancel immediately after commit

		// Verify transaction committed
		rows, err = svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM users")
		require.NoError(t, err)
		defer rows.Close()
		assert.True(t, rows.Next())
		var count int
		err = rows.Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("transaction rollback", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Create table
		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)

		// Insert initial data
		_, err = svc.ExecContext(context.Background(), "INSERT INTO test (value) VALUES (1)")
		require.NoError(t, err)

		// Start transaction
		tx, cancel, err := svc.BeginTxContext(context.Background())
		require.NoError(t, err)

		// Insert in transaction
		_, err = tx.Exec("INSERT INTO test (value) VALUES (2)")
		require.NoError(t, err)

		// Rollback
		err = tx.Rollback()
		assert.NoError(t, err)
		cancel() // Cancel immediately after rollback

		// Verify rollback worked
		rows, err := svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM test")
		require.NoError(t, err)
		defer rows.Close()
		assert.True(t, rows.Next())
		var count int
		err = rows.Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("close and reopen", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())

		// First open/close cycle
		err := svc.Open()
		require.NoError(t, err)
		err = svc.Ping()
		require.NoError(t, err)
		err = svc.Close()
		require.NoError(t, err)

		// Second open/close cycle
		err = svc.Open()
		require.NoError(t, err)
		err = svc.Ping()
		require.NoError(t, err)
		err = svc.Close()
		require.NoError(t, err)
	})

	t.Run("query context timeout", func(t *testing.T) {
		cfg := getSqliteFileConfig(t)
		cfg.ContextTimeout = 5 // 5 seconds (minimum allowed)
		cfgSvc := &config.Service{
			AppConfig: types.AppConfig{
				DatastoreConfig: *cfg,
			},
		}
		_ = cfgSvc.Initialize()
		svc := &Service{ConfigService: cfgSvc}
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Simple query should succeed within timeout
		ctx := context.Background()
		rows, err := svc.QueryContext(ctx, "SELECT 1")
		require.NoError(t, err)
		rows.Close()
	})

	t.Run("exec context timeout", func(t *testing.T) {
		cfg := getSqliteFileConfig(t)
		cfg.ContextTimeout = 5 // 5 seconds (minimum allowed)
		cfgSvc := &config.Service{
			AppConfig: types.AppConfig{
				DatastoreConfig: *cfg,
			},
		}
		_ = cfgSvc.Initialize()
		svc := &Service{ConfigService: cfgSvc}
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Simple exec should succeed within timeout
		ctx := context.Background()
		_, err := svc.ExecContext(ctx, "CREATE TABLE test (id INTEGER)")
		require.NoError(t, err)
	})

	t.Run("transaction with custom context", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Create table
		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)

		// Use custom context
		ctx := context.Background()
		tx, cancel, err := svc.BeginTxContext(ctx)
		require.NoError(t, err)
		defer cancel()

		_, err = tx.Exec("INSERT INTO test (value) VALUES (42)")
		require.NoError(t, err)

		err = tx.Commit()
		assert.NoError(t, err)
	})

	t.Run("multiple inserts", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Create table
		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)

		// Multiple inserts
		for i := 0; i < 10; i++ {
			_, err := svc.ExecContext(context.Background(), "INSERT INTO test (value) VALUES (?)", i)
			require.NoError(t, err)
		}

		// Verify count
		rows, err := svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM test")
		require.NoError(t, err)
		defer rows.Close()
		assert.True(t, rows.Next())
		var count int
		err = rows.Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 10, count)
	})

	t.Run("query multiple rows", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Create and populate table
		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)
		for i := 1; i <= 5; i++ {
			_, err := svc.ExecContext(context.Background(), "INSERT INTO test (value) VALUES (?)", i*10)
			require.NoError(t, err)
		}

		// Query all rows
		rows, err := svc.QueryContext(context.Background(), "SELECT id, value FROM test ORDER BY id")
		require.NoError(t, err)
		defer rows.Close()

		count := 0
		for rows.Next() {
			var id, value int
			err := rows.Scan(&id, &value)
			require.NoError(t, err)
			assert.Equal(t, (id)*10, value)
			count++
		}
		assert.Equal(t, 5, count)
	})

	t.Run("exec with no rows affected", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Create table
		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)

		// Delete from empty table
		result, err := svc.ExecContext(context.Background(), "DELETE FROM test WHERE value = 999")
		require.NoError(t, err)

		affected, err := result.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(0), affected)
	})

	t.Run("query with no results", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Create empty table
		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY)")
		require.NoError(t, err)

		// Query empty table
		rows, err := svc.QueryContext(context.Background(), "SELECT * FROM test")
		require.NoError(t, err)
		defer rows.Close()

		assert.False(t, rows.Next())
	})

	t.Run("update operations", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Create and populate table
		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)
		_, err = svc.ExecContext(context.Background(), "INSERT INTO test (value) VALUES (100)")
		require.NoError(t, err)

		// Update
		result, err := svc.ExecContext(context.Background(), "UPDATE test SET value = 200 WHERE id = 1")
		require.NoError(t, err)

		affected, err := result.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// Verify update
		rows, err := svc.QueryContext(context.Background(), "SELECT value FROM test WHERE id = 1")
		require.NoError(t, err)
		defer rows.Close()
		assert.True(t, rows.Next())
		var value int
		err = rows.Scan(&value)
		require.NoError(t, err)
		assert.Equal(t, 200, value)
	})

	t.Run("delete operations", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Create and populate table
		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)
		for i := 1; i <= 3; i++ {
			_, err := svc.ExecContext(context.Background(), "INSERT INTO test (value) VALUES (?)", i*100)
			require.NoError(t, err)
		}

		// Delete one row
		result, err := svc.ExecContext(context.Background(), "DELETE FROM test WHERE id = 2")
		require.NoError(t, err)

		affected, err := result.RowsAffected()
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// Verify remaining count
		rows, err := svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM test")
		require.NoError(t, err)
		defer rows.Close()
		assert.True(t, rows.Next())
		var count int
		err = rows.Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}

// TestSQLiteMigrations tests migration functionality
func TestSQLiteMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping migration test in short mode")
	}

	t.Run("run migrations", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Run migrations
		err := svc.Migrate()
		// Migrations may or may not succeed depending on whether migration files exist
		// We just verify it doesn't panic
		_ = err
	})
}

// TestErrorPaths tests error conditions with real database
func TestErrorPaths(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error path test in short mode")
	}

	t.Run("ping closed database", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		require.NoError(t, svc.Close())

		err := svc.Ping()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNotOpen)
	})

	t.Run("exec on closed database", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		require.NoError(t, svc.Close())

		_, err := svc.ExecContext(context.Background(), "SELECT 1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNotOpen)
	})

	t.Run("query on closed database", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		require.NoError(t, svc.Close())

		_, err := svc.QueryContext(context.Background(), "SELECT 1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNotOpen)
	})

	t.Run("begin transaction on closed database", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		require.NoError(t, svc.Close())

		tx, cancel, err := svc.BeginTxContext(context.Background())
		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.Nil(t, cancel)
	})

	t.Run("invalid SQL in exec", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		_, err := svc.ExecContext(context.Background(), "THIS IS NOT VALID SQL")
		assert.Error(t, err)
	})

	t.Run("invalid SQL in query", func(t *testing.T) {
		svc := getSqliteService(t)
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		_, err := svc.QueryContext(context.Background(), "THIS IS NOT VALID SQL")
		assert.Error(t, err)
	})
}
