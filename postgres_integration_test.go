package database

//
//import (
//	"context"
//	"github.com/Station-Manager/types"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//	"testing"
//	"time"
//)
//
//// getPostgresTestConfig returns a config for the test PostgreSQL database
//// Using the same config format as service_integration_test.go which is known to work
//func getPostgresTestConfig() *types.DatastoreConfig {
//	return &types.DatastoreConfig{
//		Driver:                    PostgresDriver,
//		Path:                      "test",
//		Options:                   "1234567890",
//		Host:                      "localhost",
//		Port:                      5432,
//		Database:                  "station_manager",
//		User:                      "smuser",
//		Password:                  "1q2w3e4r",
//		SSLMode:                   "disable",
//		MaxOpenConns:              2,
//		MaxIdleConns:              2,
//		ConnMaxLifetime:           10,
//		ConnMaxIdleTime:           5,
//		ContextTimeout:            5,
//		TransactionContextTimeout: 5,
//		Params: map[string]string{
//			"application_name": "station-manager",
//		},
//	}
//}
//
//// TestPostgresIntegration tests actual PostgreSQL database operations
//func TestPostgresIntegration(t *testing.T) {
//	if testing.Short() {
//		t.Skip("Skipping PostgreSQL integration test in short mode")
//	}
//
//	t.Run("full lifecycle with PostgreSQL", func(t *testing.T) {
//		cfg := getPostgresTestConfig()
//		svc := &Service{config: cfg}
//
//		// Initialize
//		err := svc.Initialize()
//		require.NoError(t, err)
//
//		// Open
//		err = svc.Open()
//		require.NoError(t, err)
//		defer svc.Close()
//
//		// Ping
//		err = svc.Ping()
//		assert.NoError(t, err)
//
//		// Create table
//		_, err = svc.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS test_users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")
//		require.NoError(t, err)
//
//		// Clean up from previous runs
//		_, _ = svc.ExecContext(context.Background(), "DELETE FROM test_users")
//
//		// Insert data
//		result, err := svc.ExecContext(context.Background(), "INSERT INTO test_users (name, email) VALUES ($1, $2)", "Alice", "alice@example.com")
//		require.NoError(t, err)
//		rowsAffected, _ := result.RowsAffected()
//		assert.Equal(t, int64(1), rowsAffected)
//
//		// Query data
//		rows, err := svc.QueryContext(context.Background(), "SELECT name, email FROM test_users WHERE name = $1", "Alice")
//		require.NoError(t, err)
//		defer rows.Close()
//
//		assert.True(t, rows.Next())
//		var name, email string
//		err = rows.Scan(&name, &email)
//		require.NoError(t, err)
//		assert.Equal(t, "Alice", name)
//		assert.Equal(t, "alice@example.com", email)
//
//		// Transaction
//		tx, cancel, err := svc.BeginTxContext(context.Background())
//		require.NoError(t, err)
//		defer cancel()
//
//		_, err = tx.Exec("INSERT INTO test_users (name, email) VALUES ($1, $2)", "Bob", "bob@example.com")
//		require.NoError(t, err)
//
//		err = tx.Commit()
//		assert.NoError(t, err)
//
//		// Verify transaction committed
//		rows, err = svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM test_users")
//		require.NoError(t, err)
//		defer rows.Close()
//		assert.True(t, rows.Next())
//		var count int
//		err = rows.Scan(&count)
//		require.NoError(t, err)
//		assert.Equal(t, 2, count)
//
//		// Clean up
//		_, _ = svc.ExecContext(context.Background(), "DROP TABLE IF EXISTS test_users")
//	})
//
//	t.Run("transaction rollback", func(t *testing.T) {
//		cfg := getPostgresTestConfig()
//		svc := &Service{config: cfg}
//		require.NoError(t, svc.Initialize())
//		require.NoError(t, svc.Open())
//		defer svc.Close()
//
//		// Create table
//		_, err := svc.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS test_rollback (id SERIAL PRIMARY KEY, value INTEGER)")
//		require.NoError(t, err)
//		defer func() {
//			_, _ = svc.ExecContext(context.Background(), "DROP TABLE IF EXISTS test_rollback")
//		}()
//
//		// Clean and insert initial data
//		_, _ = svc.ExecContext(context.Background(), "DELETE FROM test_rollback")
//		_, err = svc.ExecContext(context.Background(), "INSERT INTO test_rollback (value) VALUES (1)")
//		require.NoError(t, err)
//
//		// Start transaction
//		tx, cancel, err := svc.BeginTxContext(context.Background())
//		require.NoError(t, err)
//		defer cancel()
//
//		// Insert in transaction
//		_, err = tx.Exec("INSERT INTO test_rollback (value) VALUES (2)")
//		require.NoError(t, err)
//
//		// Rollback
//		err = tx.Rollback()
//		assert.NoError(t, err)
//
//		// Verify rollback worked
//		rows, err := svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM test_rollback")
//		require.NoError(t, err)
//		defer rows.Close()
//		assert.True(t, rows.Next())
//		var count int
//		err = rows.Scan(&count)
//		require.NoError(t, err)
//		assert.Equal(t, 1, count)
//	})
//
//	t.Run("close and reopen", func(t *testing.T) {
//		cfg := getPostgresTestConfig()
//		svc := &Service{config: cfg}
//		require.NoError(t, svc.Initialize())
//
//		// First open/close cycle
//		err := svc.Open()
//		require.NoError(t, err)
//		err = svc.Ping()
//		require.NoError(t, err)
//		err = svc.Close()
//		require.NoError(t, err)
//
//		// Second open/close cycle
//		err = svc.Open()
//		require.NoError(t, err)
//		err = svc.Ping()
//		require.NoError(t, err)
//		err = svc.Close()
//		require.NoError(t, err)
//	})
//
//	t.Run("multiple concurrent operations", func(t *testing.T) {
//		cfg := getPostgresTestConfig()
//		svc := &Service{config: cfg}
//		require.NoError(t, svc.Initialize())
//		require.NoError(t, svc.Open())
//		defer svc.Close()
//
//		// Create table
//		_, err := svc.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS test_concurrent (id SERIAL PRIMARY KEY, value INTEGER)")
//		require.NoError(t, err)
//		defer func() {
//			_, _ = svc.ExecContext(context.Background(), "DROP TABLE IF EXISTS test_concurrent")
//		}()
//
//		// Clean
//		_, _ = svc.ExecContext(context.Background(), "DELETE FROM test_concurrent")
//
//		// Multiple inserts
//		for i := 0; i < 10; i++ {
//			_, err := svc.ExecContext(context.Background(), "INSERT INTO test_concurrent (value) VALUES ($1)", i)
//			require.NoError(t, err)
//		}
//
//		// Verify count
//		rows, err := svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM test_concurrent")
//		require.NoError(t, err)
//		defer rows.Close()
//		assert.True(t, rows.Next())
//		var count int
//		err = rows.Scan(&count)
//		require.NoError(t, err)
//		assert.Equal(t, 10, count)
//	})
//
//	t.Run("context with custom deadline", func(t *testing.T) {
//		cfg := getPostgresTestConfig()
//		svc := &Service{config: cfg}
//		require.NoError(t, svc.Initialize())
//		require.NoError(t, svc.Open())
//		defer svc.Close()
//
//		// Create context with deadline
//		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//		defer cancel()
//
//		// Query should respect context
//		rows, err := svc.QueryContext(ctx, "SELECT 1")
//		require.NoError(t, err)
//		rows.Close()
//	})
//
//	t.Run("exec and query operations", func(t *testing.T) {
//		cfg := getPostgresTestConfig()
//		svc := &Service{config: cfg}
//		require.NoError(t, svc.Initialize())
//		require.NoError(t, svc.Open())
//		defer svc.Close()
//
//		// Create table
//		_, err := svc.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS test_ops (id SERIAL PRIMARY KEY, name TEXT, value INTEGER)")
//		require.NoError(t, err)
//		defer func() {
//			_, _ = svc.ExecContext(context.Background(), "DROP TABLE IF EXISTS test_ops")
//		}()
//
//		// Clean
//		_, _ = svc.ExecContext(context.Background(), "DELETE FROM test_ops")
//
//		// Insert
//		result, err := svc.ExecContext(context.Background(), "INSERT INTO test_ops (name, value) VALUES ($1, $2)", "test", 100)
//		require.NoError(t, err)
//		affected, _ := result.RowsAffected()
//		assert.Equal(t, int64(1), affected)
//
//		// Update
//		result, err = svc.ExecContext(context.Background(), "UPDATE test_ops SET value = $1 WHERE name = $2", 200, "test")
//		require.NoError(t, err)
//		affected, _ = result.RowsAffected()
//		assert.Equal(t, int64(1), affected)
//
//		// Query to verify
//		rows, err := svc.QueryContext(context.Background(), "SELECT name, value FROM test_ops WHERE name = $1", "test")
//		require.NoError(t, err)
//		defer rows.Close()
//		assert.True(t, rows.Next())
//		var name string
//		var value int
//		err = rows.Scan(&name, &value)
//		require.NoError(t, err)
//		assert.Equal(t, "test", name)
//		assert.Equal(t, 200, value)
//
//		// Delete
//		result, err = svc.ExecContext(context.Background(), "DELETE FROM test_ops WHERE name = $1", "test")
//		require.NoError(t, err)
//		affected, _ = result.RowsAffected()
//		assert.Equal(t, int64(1), affected)
//	})
//
//	t.Run("migrations", func(t *testing.T) {
//		cfg := getPostgresTestConfig()
//		svc := &Service{config: cfg}
//		require.NoError(t, svc.Initialize())
//		require.NoError(t, svc.Open())
//		defer svc.Close()
//
//		// Run migrations
//		err := svc.Migrate()
//		// Migrations should succeed or return no change
//		assert.NoError(t, err)
//	})
//}
//
//// TestPostgresErrorPaths tests error conditions with real database
//func TestPostgresErrorPaths(t *testing.T) {
//	if testing.Short() {
//		t.Skip("Skipping error path test in short mode")
//	}
//
//	t.Run("invalid SQL in exec", func(t *testing.T) {
//		cfg := getPostgresTestConfig()
//		svc := &Service{config: cfg}
//		require.NoError(t, svc.Initialize())
//		require.NoError(t, svc.Open())
//		defer svc.Close()
//
//		_, err := svc.ExecContext(context.Background(), "THIS IS NOT VALID SQL")
//		assert.Error(t, err)
//	})
//
//	t.Run("invalid SQL in query", func(t *testing.T) {
//		cfg := getPostgresTestConfig()
//		svc := &Service{config: cfg}
//		require.NoError(t, svc.Initialize())
//		require.NoError(t, svc.Open())
//		defer svc.Close()
//
//		_, err := svc.QueryContext(context.Background(), "THIS IS NOT VALID SQL")
//		assert.Error(t, err)
//	})
//
//	t.Run("operations on closed database", func(t *testing.T) {
//		cfg := getPostgresTestConfig()
//		svc := &Service{config: cfg}
//		require.NoError(t, svc.Initialize())
//		require.NoError(t, svc.Open())
//		require.NoError(t, svc.Close())
//
//		// All operations should fail
//		err := svc.Ping()
//		assert.Error(t, err)
//
//		_, err = svc.ExecContext(context.Background(), "SELECT 1")
//		assert.Error(t, err)
//
//		_, err = svc.QueryContext(context.Background(), "SELECT 1")
//		assert.Error(t, err)
//
//		tx, cancel, err := svc.BeginTxContext(context.Background())
//		assert.Error(t, err)
//		assert.Nil(t, tx)
//		assert.Nil(t, cancel)
//	})
//}
