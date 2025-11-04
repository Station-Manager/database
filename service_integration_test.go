package database

import (
	"fmt"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_PostgresMigration(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    "postgres",
			Path:                      "test",
			Options:                   "1234567890",
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
		fmt.Println(err)
		assert.NoError(t, err)

		err = svc.Ping()
		assert.NoError(t, err)

		err = svc.Migrate()
		assert.NoError(t, err)

		err = svc.Close()
		assert.NoError(t, err)
	})
}
