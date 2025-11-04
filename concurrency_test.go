package database

import (
	"context"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func getValidSqliteConfigForConcurrency(t *testing.T) *types.DatastoreConfig {
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
		MaxOpenConns:              1,
		MaxIdleConns:              1,
		ConnMaxLifetime:           15,
		ConnMaxIdleTime:           5,
		ContextTimeout:            5,
		TransactionContextTimeout: 10,
	}
}

// TestConcurrency_OpenClose tests concurrent open and close operations
func TestConcurrency_OpenClose(t *testing.T) {
	t.Run("concurrent open and close are safe", func(t *testing.T) {
		svc := &Service{config: getValidSqliteConfigForConcurrency(t)}
		require.NoError(t, svc.Initialize())

		var wg sync.WaitGroup
		iterations := 100

		// Repeatedly open and close
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = svc.Open()
				time.Sleep(1 * time.Millisecond)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = svc.Close()
				time.Sleep(1 * time.Millisecond)
			}
		}()

		wg.Wait()

		// Final state should be consistent
		assert.Equal(t, svc.isOpen.Load(), svc.handle != nil)
	})

	t.Run("concurrent pings during open/close", func(t *testing.T) {
		svc := &Service{config: getValidSqliteConfigForConcurrency(t)}
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())

		var wg sync.WaitGroup
		done := make(chan struct{})

		// Continuously ping
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_ = svc.Ping()
					time.Sleep(1 * time.Millisecond)
				}
			}
		}()

		// Close and reopen
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			_ = svc.Close()
			time.Sleep(10 * time.Millisecond)
			_ = svc.Open()
		}()

		time.Sleep(50 * time.Millisecond)
		close(done)
		wg.Wait()
	})
}

// TestConcurrency_QueryExec tests concurrent query and exec operations
func TestConcurrency_QueryExec(t *testing.T) {
	t.Run("concurrent queries are safe", func(t *testing.T) {
		svc := &Service{config: getValidSqliteConfigForConcurrency(t)}
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		// Create test table
		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)
		_, err = svc.ExecContext(context.Background(), "INSERT INTO test (value) VALUES (1), (2), (3)")
		require.NoError(t, err)

		var wg sync.WaitGroup
		numGoroutines := 10
		queriesPerGoroutine := 50

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < queriesPerGoroutine; j++ {
					rows, err := svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM test")
					if err == nil {
						rows.Close()
					}
				}
			}()
		}

		wg.Wait()
	})

	t.Run("concurrent execs are safe", func(t *testing.T) {
		svc := &Service{config: getValidSqliteConfigForConcurrency(t)}
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)

		var wg sync.WaitGroup
		numGoroutines := 10
		insertsPerGoroutine := 10
		var successCount atomic.Int64

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < insertsPerGoroutine; j++ {
					_, err := svc.ExecContext(context.Background(), "INSERT INTO test (value) VALUES (?)", j)
					if err == nil {
						successCount.Add(1)
					}
				}
			}()
		}

		wg.Wait()

		// Verify all inserts succeeded
		rows, err := svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM test")
		require.NoError(t, err)
		defer rows.Close()

		var count int
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&count))
		assert.Equal(t, int(successCount.Load()), count)
	})

	t.Run("concurrent mixed query and exec", func(t *testing.T) {
		svc := &Service{config: getValidSqliteConfigForConcurrency(t)}
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)

		var wg sync.WaitGroup
		numGoroutines := 20
		operationsPerGoroutine := 25

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					if j%2 == 0 {
						// Query
						rows, err := svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM test")
						if err == nil {
							rows.Close()
						}
					} else {
						// Insert
						_, _ = svc.ExecContext(context.Background(), "INSERT INTO test (value) VALUES (?)", id*1000+j)
					}
				}
			}(i)
		}

		wg.Wait()
	})
}

// TestConcurrency_Transactions tests concurrent transaction operations
func TestConcurrency_Transactions(t *testing.T) {
	t.Run("multiple concurrent transactions", func(t *testing.T) {
		svc := &Service{config: getValidSqliteConfigForConcurrency(t)}
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)

		var wg sync.WaitGroup
		numTransactions := 10
		var successCount atomic.Int64

		for i := 0; i < numTransactions; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				tx, cancel, err := svc.BeginTxContext(context.Background())
				if err != nil {
					return
				}
				defer cancel()

				_, err = tx.Exec("INSERT INTO test (value) VALUES (?)", id)
				if err != nil {
					_ = tx.Rollback()
					return
				}

				err = tx.Commit()
				if err == nil {
					successCount.Add(1)
				}
			}(i)
		}

		wg.Wait()

		// All transactions should succeed
		assert.Equal(t, int64(numTransactions), successCount.Load())

		// Verify row count
		rows, err := svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM test")
		require.NoError(t, err)
		defer rows.Close()

		var count int
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&count))
		assert.Equal(t, numTransactions, count)
	})

	t.Run("transaction rollback and commit mix", func(t *testing.T) {
		svc := &Service{config: getValidSqliteConfigForConcurrency(t)}
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		_, err := svc.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, value INTEGER)")
		require.NoError(t, err)

		var wg sync.WaitGroup
		numTransactions := 20
		var commitCount atomic.Int64

		for i := 0; i < numTransactions; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				tx, cancel, err := svc.BeginTxContext(context.Background())
				if err != nil {
					return
				}
				defer cancel()

				_, err = tx.Exec("INSERT INTO test (value) VALUES (?)", id)
				if err != nil {
					_ = tx.Rollback()
					return
				}

				// Commit even transactions, rollback odd ones
				if id%2 == 0 {
					err = tx.Commit()
					if err == nil {
						commitCount.Add(1)
					}
				} else {
					_ = tx.Rollback()
				}
			}(i)
		}

		wg.Wait()

		// Verify only committed transactions persisted
		rows, err := svc.QueryContext(context.Background(), "SELECT COUNT(*) FROM test")
		require.NoError(t, err)
		defer rows.Close()

		var count int
		require.True(t, rows.Next())
		require.NoError(t, rows.Scan(&count))
		assert.Equal(t, int(commitCount.Load()), count)
	})
}

// TestConcurrency_OperationsDuringClose tests operations during close
func TestConcurrency_OperationsDuringClose(t *testing.T) {
	t.Run("queries during close handle gracefully", func(t *testing.T) {
		svc := &Service{config: getValidSqliteConfigForConcurrency(t)}
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())

		var wg sync.WaitGroup
		done := make(chan struct{})

		// Continuous queries
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					default:
						_, _ = svc.QueryContext(context.Background(), "SELECT 1")
						time.Sleep(1 * time.Millisecond)
					}
				}
			}()
		}

		// Close after a delay
		time.Sleep(50 * time.Millisecond)
		err := svc.Close()
		assert.NoError(t, err)

		close(done)
		wg.Wait()
	})

	t.Run("transactions during close handle gracefully", func(t *testing.T) {
		svc := &Service{config: getValidSqliteConfigForConcurrency(t)}
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())

		var wg sync.WaitGroup
		done := make(chan struct{})

		// Continuous transaction attempts
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					default:
						tx, cancel, err := svc.BeginTxContext(context.Background())
						if err == nil {
							_ = tx.Rollback()
							cancel()
						}
						time.Sleep(5 * time.Millisecond)
					}
				}
			}()
		}

		// Close after a delay
		time.Sleep(50 * time.Millisecond)
		err := svc.Close()
		assert.NoError(t, err)

		close(done)
		wg.Wait()
	})
}

// TestConcurrency_StateConsistency tests state consistency
func TestConcurrency_StateConsistency(t *testing.T) {
	t.Run("isOpen and handle are always consistent", func(t *testing.T) {
		svc := &Service{config: getValidSqliteConfigForConcurrency(t)}
		require.NoError(t, svc.Initialize())

		var wg sync.WaitGroup
		done := make(chan struct{})
		var inconsistencies atomic.Int64

		// Goroutine that checks consistency
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					default:
						svc.mu.RLock()
						isOpen := svc.isOpen.Load()
						handle := svc.handle
						svc.mu.RUnlock()

						// If open, handle should not be nil
						// If not open, handle should be nil
						if isOpen && handle == nil {
							inconsistencies.Add(1)
						}
						if !isOpen && handle != nil {
							inconsistencies.Add(1)
						}
					}
				}
			}()
		}

		// Repeatedly open and close
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				_ = svc.Open()
				time.Sleep(2 * time.Millisecond)
				_ = svc.Close()
				time.Sleep(2 * time.Millisecond)
			}
		}()

		time.Sleep(500 * time.Millisecond)
		close(done)
		wg.Wait()

		assert.Equal(t, int64(0), inconsistencies.Load(), "state should always be consistent")
	})
}

// TestConcurrency_PingStorm tests many concurrent pings
func TestConcurrency_PingStorm(t *testing.T) {
	t.Run("handle many concurrent pings", func(t *testing.T) {
		svc := &Service{config: getValidSqliteConfigForConcurrency(t)}
		require.NoError(t, svc.Initialize())
		require.NoError(t, svc.Open())
		defer svc.Close()

		var wg sync.WaitGroup
		numGoroutines := 100
		pingsPerGoroutine := 10
		var successCount atomic.Int64

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < pingsPerGoroutine; j++ {
					err := svc.Ping()
					if err == nil {
						successCount.Add(1)
					}
				}
			}()
		}

		wg.Wait()

		assert.Equal(t, int64(numGoroutines*pingsPerGoroutine), successCount.Load())
	})
}

// BenchmarkConcurrency benchmarks concurrent operations
func BenchmarkConcurrency_Query(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")
	cfg := &types.DatastoreConfig{
		Driver:                    SqliteDriver,
		Path:                      dbPath,
		MaxOpenConns:              10,
		MaxIdleConns:              5,
		ConnMaxLifetime:           15,
		ConnMaxIdleTime:           5,
		ContextTimeout:            5,
		TransactionContextTimeout: 10,
	}
	svc := &Service{config: cfg}
	_ = svc.Initialize()
	_ = svc.Open()
	defer svc.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rows, err := svc.QueryContext(context.Background(), "SELECT 1")
			if err == nil {
				rows.Close()
			}
		}
	})
}

func BenchmarkConcurrency_Exec(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")
	cfg := &types.DatastoreConfig{
		Driver:                    SqliteDriver,
		Path:                      dbPath,
		MaxOpenConns:              10,
		MaxIdleConns:              5,
		ConnMaxLifetime:           15,
		ConnMaxIdleTime:           5,
		ContextTimeout:            5,
		TransactionContextTimeout: 10,
	}
	svc := &Service{config: cfg}
	_ = svc.Initialize()
	_ = svc.Open()
	defer svc.Close()

	_, _ = svc.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS bench (id INTEGER PRIMARY KEY, value INTEGER)")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = svc.ExecContext(context.Background(), "INSERT INTO bench (value) VALUES (1)")
		}
	})
}

func BenchmarkConcurrency_Transaction(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "bench.db")
	cfg := &types.DatastoreConfig{
		Driver:                    SqliteDriver,
		Path:                      dbPath,
		MaxOpenConns:              10,
		MaxIdleConns:              5,
		ConnMaxLifetime:           15,
		ConnMaxIdleTime:           5,
		ContextTimeout:            5,
		TransactionContextTimeout: 10,
	}
	svc := &Service{config: cfg}
	_ = svc.Initialize()
	_ = svc.Open()
	defer svc.Close()

	_, _ = svc.ExecContext(context.Background(), "CREATE TABLE IF NOT EXISTS bench (id INTEGER PRIMARY KEY, value INTEGER)")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tx, cancel, err := svc.BeginTxContext(context.Background())
			if err == nil {
				_, _ = tx.Exec("INSERT INTO bench (value) VALUES (1)")
				_ = tx.Commit()
				cancel()
			}
		}
	})
}
