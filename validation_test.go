package database

import (
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	t.Run("nil config returns error", func(t *testing.T) {
		err := validateConfig(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgNilConfig)
	})

	t.Run("empty config fails validation", func(t *testing.T) {
		cfg := &types.DatastoreConfig{}
		err := validateConfig(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgConfigInvalid)
	})

	t.Run("valid postgres config passes", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("valid sqlite config passes", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    SqliteDriver,
			Path:                      "/tmp/test.db",
			Options:                   "_busy_timeout=5000",
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
		err := validateConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("invalid driver fails", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    "mysql",
			Host:                      "localhost",
			Port:                      3306,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), errMsgConfigInvalid)
	})

	t.Run("missing required field Driver", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("invalid port zero", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      0,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("invalid port too high", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      70000,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("invalid SSLMode", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "invalid",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("invalid MaxOpenConns zero", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              0,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("invalid MaxIdleConns zero", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              0,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("ContextTimeout too low", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            2,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("TransactionContextTimeout too low", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 2,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("empty host fails", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("empty user fails", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("empty password fails", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("empty database fails", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("all SSL modes are valid", func(t *testing.T) {
		sslModes := []string{"disable", "require", "verify-ca", "verify-full"}

		for _, mode := range sslModes {
			cfg := &types.DatastoreConfig{
				Driver:                    PostgresDriver,
				Host:                      "localhost",
				Port:                      5432,
				Database:                  "testdb",
				User:                      "testuser",
				Password:                  "password",
				SSLMode:                   mode,
				MaxOpenConns:              10,
				MaxIdleConns:              5,
				ConnMaxLifetime:           15,
				ConnMaxIdleTime:           5,
				ContextTimeout:            5,
				TransactionContextTimeout: 10,
			}
			err := validateConfig(cfg)
			assert.NoError(t, err, "SSL mode %s should be valid", mode)
		}
	})

	t.Run("zero ConnMaxLifetime is valid", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           0,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("zero ConnMaxIdleTime is valid", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    PostgresDriver,
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "testdb",
			User:                      "testuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              10,
			MaxIdleConns:              5,
			ConnMaxLifetime:           15,
			ConnMaxIdleTime:           0,
			ContextTimeout:            5,
			TransactionContextTimeout: 10,
		}
		err := validateConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("sqlite without path fails", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    SqliteDriver,
			Path:                      "",
			Options:                   "_busy_timeout=5000",
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
		err := validateConfig(cfg)
		require.Error(t, err)
	})

	t.Run("sqlite without options fails", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    SqliteDriver,
			Path:                      "/tmp/test.db",
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
		err := validateConfig(cfg)
		// Based on the validation tags, Options is required for sqlite3
		// However, the getDsn() function provides defaults for empty options
		// This might be a validation/implementation mismatch
		require.Error(t, err)
	})

	t.Run("validator is only initialized once", func(t *testing.T) {
		// This test ensures the sync.Once works correctly
		cfg1 := getValidPostgresConfig()
		cfg2 := getValidPostgresConfig()

		err1 := validateConfig(cfg1)
		err2 := validateConfig(cfg2)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
	})

	t.Run("boundary values for ports", func(t *testing.T) {
		testCases := []struct {
			port  int
			valid bool
			name  string
		}{
			{1, true, "minimum valid port"},
			{65535, true, "maximum valid port"},
			{0, false, "zero port"},
			{65536, false, "port too high"},
			{-1, false, "negative port"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := &types.DatastoreConfig{
					Driver:                    PostgresDriver,
					Host:                      "localhost",
					Port:                      tc.port,
					Database:                  "testdb",
					User:                      "testuser",
					Password:                  "password",
					SSLMode:                   "disable",
					MaxOpenConns:              10,
					MaxIdleConns:              5,
					ConnMaxLifetime:           15,
					ConnMaxIdleTime:           5,
					ContextTimeout:            5,
					TransactionContextTimeout: 10,
				}
				err := validateConfig(cfg)
				if tc.valid {
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}
			})
		}
	})
}
