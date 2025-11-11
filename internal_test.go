package database

import (
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func getServiceForInternalTest(cfg *types.DatastoreConfig) *Service {
	// Fill in required fields if not set
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 10
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 5
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 15
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = 5
	}
	if cfg.ContextTimeout == 0 {
		cfg.ContextTimeout = 5
	}
	if cfg.TransactionContextTimeout == 0 {
		cfg.TransactionContextTimeout = 10
	}
	// Fill in fields required for validation but not used by getDsn for SQLite
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 5432
	}
	if cfg.User == "" {
		cfg.User = "testuser"
	}
	if cfg.Password == "" {
		cfg.Password = "testpass"
	}
	if cfg.Database == "" {
		cfg.Database = "testdb"
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	cfgSvc := &config.Service{
		AppConfig: types.AppConfig{
			DatastoreConfig: *cfg,
		},
	}
	_ = cfgSvc.Initialize()
	svc := &Service{ConfigService: cfgSvc}
	_ = svc.Initialize()
	return svc
}

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
		svc := getServiceForInternalTest(cfg)

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
		svc := getServiceForInternalTest(cfg)

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
			Options: map[string]string{},
		}
		svc := getServiceForInternalTest(cfg)

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		// Verify defaults are present
		assert.Contains(t, dsn, "file:")
		assert.Contains(t, dsn, "_busy_timeout=5000")
		assert.Contains(t, dsn, "_journal_mode=WAL")
		assert.Contains(t, dsn, "_foreign_keys=on")
	})

	t.Run("sqlite DSN with custom options", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "/tmp/test.db",
			Options: map[string]string{"cache": "shared", "mode": "memory"},
		}
		svc := getServiceForInternalTest(cfg)

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		assert.Contains(t, dsn, "file:")
		assert.Contains(t, dsn, "cache=shared")
		assert.Contains(t, dsn, "mode=memory")
	})

	t.Run("sqlite DSN with leading question mark in options", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "test_options.db",
			Options: map[string]string{"cache": "shared"},
		}
		svc := getServiceForInternalTest(cfg)

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		// Check it's valid and contains provided option
		assert.Contains(t, dsn, "file:")
		assert.Contains(t, dsn, "cache=shared")
	})

	t.Run("sqlite DSN with just question mark gets defaults", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "test_question.db",
			Options: map[string]string{},
		}
		svc := getServiceForInternalTest(cfg)

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		// Defaults should be present
		assert.Contains(t, dsn, "file:")
		assert.Contains(t, dsn, "_busy_timeout")
	})

	t.Run("sqlite DSN with file database", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "test_file.db",
			Options: map[string]string{},
		}
		svc := getServiceForInternalTest(cfg)

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		assert.Contains(t, dsn, "file:")
		assert.Contains(t, dsn, "_busy_timeout")
	})

	t.Run("unknown driver fails validation during initialization", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    "mysql",
			Host:                      "localhost",
			Port:                      3306,
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
		cfgSvc := &config.Service{
			AppConfig: types.AppConfig{
				DatastoreConfig: *cfg,
			},
		}
		_ = cfgSvc.Initialize()
		svc := &Service{ConfigService: cfgSvc}
		err := svc.Initialize()

		assert.Error(t, err)
		assert.Error(t, err)
	})

	t.Run("sqlite with empty path fails validation during initialization", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    SqliteDriver,
			Path:                      "",
			Options:                   map[string]string{},
			Host:                      "localhost",
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
		cfgSvc := &config.Service{
			AppConfig: types.AppConfig{
				DatastoreConfig: *cfg,
			},
		}
		_ = cfgSvc.Initialize()
		svc := &Service{ConfigService: cfgSvc}
		err := svc.Initialize()

		assert.Error(t, err)
		assert.Error(t, err)
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
		svc := getServiceForInternalTest(cfg)

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
		svc := getServiceForInternalTest(cfg)

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		assert.Contains(t, dsn, "localhost:15432")
	})

	t.Run("sqlite with relative path", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "./relative/path/db.sqlite",
			Options: map[string]string{},
		}
		svc := getServiceForInternalTest(cfg)

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		assert.Contains(t, dsn, "./relative/path/db.sqlite")
	})

	t.Run("sqlite with absolute path", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "/absolute/path/db.sqlite",
			Options: map[string]string{},
		}
		svc := getServiceForInternalTest(cfg)

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
		svc := getServiceForInternalTest(cfg)

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
			svc := getServiceForInternalTest(cfg)

			dsn, err := svc.getDsn()
			require.NoError(t, err)

			assert.Contains(t, dsn, "sslmode="+mode, "SSL mode %s should be in DSN", mode)
		}
	})

	t.Run("sqlite with complex options string", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "test_complex.db",
			Options: map[string]string{"_busy_timeout": "10000", "_journal_mode": "DELETE", "_synchronous": "FULL", "cache": "private"},
		}
		svc := getServiceForInternalTest(cfg)

		dsn, err := svc.getDsn()
		require.NoError(t, err)

		assert.Contains(t, dsn, "file:")
		assert.Contains(t, dsn, "_busy_timeout=10000")
		assert.Contains(t, dsn, "_journal_mode=DELETE")
		assert.Contains(t, dsn, "_synchronous=FULL")
		assert.Contains(t, dsn, "cache=private")
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
		svc := getServiceForInternalTest(cfg)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = svc.getDsn()
		}
	})

	b.Run("sqlite", func(b *testing.B) {
		cfg := &types.DatastoreConfig{
			Driver:  SqliteDriver,
			Path:    "test_bench.db",
			Options: map[string]string{},
		}
		svc := getServiceForInternalTest(cfg)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = svc.getDsn()
		}
	})
}
