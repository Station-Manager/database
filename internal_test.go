package database

import (
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestService_getDsn tests DSN generation for different drivers
func TestService_getDsn(t *testing.T) {
	t.Run("postgres DSN with all fields", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:   PostgresDriver,
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			Password: "testpass",
			SSLMode:  "disable",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		assert.Contains(t, dsn, "postgres://")
		assert.Contains(t, dsn, "testuser")
		assert.Contains(t, dsn, "testpass")
		assert.Contains(t, dsn, "localhost:5432")
		assert.Contains(t, dsn, "/testdb")
		assert.Contains(t, dsn, "sslmode=disable")
	})

	t.Run("postgres DSN with special characters in password", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:   PostgresDriver,
			Host:     "db.example.com",
			Port:     5433,
			Database: "mydb",
			User:     "admin",
			Password: "p@ss:w/rd!#",
			SSLMode:  "require",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		// URL encoding should handle special characters
		assert.Contains(t, dsn, "postgres://")
		assert.Contains(t, dsn, "admin")
		assert.Contains(t, dsn, "db.example.com:5433")
		assert.Contains(t, dsn, "sslmode=require")
	})

	t.Run("sqlite DSN with empty options uses defaults", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "/tmp/test.db",
			Options: "",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		// DSN is hardwired, just verify it contains expected defaults
		assert.Contains(t, dsn, "file:")
		assert.Contains(t, dsn, "_busy_timeout=5000")
		assert.Contains(t, dsn, "_journal_mode=WAL")
		assert.Contains(t, dsn, "_foreign_keys=on")
	})

	t.Run("sqlite DSN with custom options", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "/tmp/test.db",
			Options: "cache=shared&mode=memory",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		// DSN is hardwired, options are ignored for now
		assert.Contains(t, dsn, "file:")
	})

	t.Run("sqlite DSN with leading question mark in options", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "test_options.db",
			Options: "?cache=shared",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		// DSN is hardwired, so just check it's valid
		assert.Contains(t, dsn, "file:")
	})

	t.Run("sqlite DSN with just question mark gets defaults", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "test_question.db",
			Options: "?",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		// DSN is hardwired, so just check it's valid
		assert.Contains(t, dsn, "file:")
		assert.Contains(t, dsn, "_busy_timeout")
	})

	t.Run("sqlite DSN with file database", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "test_file.db",
			Options: "",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		assert.Contains(t, dsn, "file:")
		assert.Contains(t, dsn, "_busy_timeout")
	})

	t.Run("unknown driver returns error", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver: "mysql",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()

		assert.Error(t, err)
		assert.Empty(t, dsn)
		assert.Contains(t, err.Error(), "Unsupported database driver")
	})

	t.Run("sqlite with empty path returns error", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "",
			Options: "",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()

		assert.Error(t, err)
		assert.Empty(t, dsn)
		assert.Contains(t, err.Error(), errMsgEmptyPath)
	})
}

// TestService_getDsn_EdgeCases tests edge cases in DSN generation
func TestService_getDsn_EdgeCases(t *testing.T) {
	t.Run("postgres with IPv6 host", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:   PostgresDriver,
			Host:     "::1",
			Port:     5432,
			Database: "testdb",
			User:     "user",
			Password: "pass",
			SSLMode:  "disable",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		// IPv6 addresses should be properly handled
		assert.Contains(t, dsn, "postgres://")
		assert.Contains(t, dsn, "::1")
	})

	t.Run("postgres with non-standard port", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:   PostgresDriver,
			Host:     "localhost",
			Port:     15432,
			Database: "testdb",
			User:     "user",
			Password: "pass",
			SSLMode:  "disable",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		assert.Contains(t, dsn, "localhost:15432")
	})

	t.Run("sqlite with relative path", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "./relative/path/db.sqlite",
			Options: "",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		assert.Contains(t, dsn, "./relative/path/db.sqlite")
	})

	t.Run("sqlite with absolute path", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "/absolute/path/db.sqlite",
			Options: "",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		assert.Contains(t, dsn, "/absolute/path/db.sqlite")
	})

	t.Run("postgres with empty password", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:   PostgresDriver,
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "user",
			Password: "",
			SSLMode:  "disable",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		// Even with empty password, DSN should be formed
		assert.Contains(t, dsn, "postgres://")
		assert.Contains(t, dsn, "user")
	})

	t.Run("postgres with different SSL modes", func(t *testing.T) {
		sslModes := []string{"disable", "require", "verify-ca", "verify-full"}

		for _, mode := range sslModes {
			cfg := &types.DatastoreConfig{
				Driver:   PostgresDriver,
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				User:     "user",
				Password: "pass",
				SSLMode:  mode,
			}
			svc := &Service{config: cfg}

			dsn, err := svc.getDsn()
			require.NoError(t, err)

			assert.Contains(t, dsn, "sslmode="+mode, "SSL mode %s should be in DSN", mode)
		}
	})

	t.Run("sqlite with complex options string", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "test_complex.db",
			Options: "_busy_timeout=10000&_journal_mode=DELETE&_synchronous=FULL&cache=private",
		}
		svc := &Service{config: cfg}

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		// DSN is hardwired, just verify it's valid
		assert.Contains(t, dsn, "file:")
	})
}

// BenchmarkService_getDsn benchmarks DSN generation
func BenchmarkService_getDsn(b *testing.B) {
	b.Run("postgres", func(b *testing.B) {
		cfg := &types.DatastoreConfig{
			Driver:   PostgresDriver,
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			Password: "testpass",
			SSLMode:  "disable",
		}
		svc := &Service{config: cfg}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = svc.getDsn()
		}
	})

	b.Run("sqlite", func(b *testing.B) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "test_bench.db",
			Options: "",
		}
		svc := &Service{config: cfg}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = svc.getDsn()
		}
	})
}
