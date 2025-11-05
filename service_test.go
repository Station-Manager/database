package database

import (
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	t.Run("nil config service", func(t *testing.T) {
		svc := &Service{}
		err := svc.Initialize()
		assert.Error(t, err)
		assert.Equal(t, errMsgAppConfigNil, err.Error())
	})
	t.Run("invalid config", func(t *testing.T) {
		cfg := &types.DatastoreConfig{}
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
	t.Run("valid config", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    "postgres",
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "station_manager",
			User:                      "smuser",
			Password:                  "password",
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
		svc := &Service{ConfigService: cfgSvc}
		err := svc.Initialize()
		assert.NoError(t, err)
	})
	t.Run("invalid driver", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:                    "invalid",
			Host:                      "localhost",
			Port:                      5432,
			Database:                  "station_manager",
			User:                      "smuser",
			Password:                  "password",
			SSLMode:                   "disable",
			MaxOpenConns:              2,
			MaxIdleConns:              2,
			ConnMaxLifetime:           10,
			ConnMaxIdleTime:           5,
			ContextTimeout:            5,
			TransactionContextTimeout: 5,
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
