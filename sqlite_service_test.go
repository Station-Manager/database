package database

import (
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSqliteService(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    "sqlite",
			Path:                      "data/test.db",
			Options:                   "",
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
		svc := &Service{config: cfg}
		err := svc.Initialize()
		assert.NoError(t, err)

		err = svc.Open()
		assert.NoError(t, err)
		assert.True(t, svc.isOpen.Load())
		_ = svc.Close()
	})
}
