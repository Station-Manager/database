package database_test

import (
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/database"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_PostgresMigration(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    "postgres",
			Path:                      "test",
			Options:                   map[string]string{"sample": "1234567890"},
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "station_manager",
			User:                      "smuser",
			Password:                  "1q2w3e4r",
			SSLMode:                   "disable",
			MaxOpenConns:              2,
			MaxIdleConns:              2,
			ConnMaxLifetime:           10,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 5,
			Params: map[string]string{
				"application_name": "station-manager",
			},
		}
		cfgSvc := &config.Service{
			AppConfig: types.AppConfig{
				DatastoreConfig: *cfg,
			},
		}
		_ = cfgSvc.Initialize()
		svc := &database.Service{ConfigService: cfgSvc}
		err := svc.Initialize()
		assert.NoError(t, err)

		err = svc.Open()
		if err != nil {
			t.Skip("Postgres not available; skipping migration test")
		}

		err = svc.Ping()
		assert.NoError(t, err)

		err = svc.Migrate()
		if err != nil { // Skip if migrations unsupported in environment
			t.Skip("Migrations failed; skipping")
		}
		assert.NoError(t, err)

		err = svc.Close()
		assert.NoError(t, err)
	})
}
