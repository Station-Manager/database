package database

import (
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		svc := &Service{}
		err := svc.Initialize()
		assert.Error(t, err)
		assert.Equal(t, errMsgNilConfig, err.Error())
	})
	t.Run("invalid config", func(t *testing.T) {
		cfg := &types.DatastoreConfig{}
		svc := &Service{config: cfg}
		err := svc.Initialize()
		assert.Error(t, err)
		assert.Equal(t, errMsgConfigInvalid, err.Error())
	})
	t.Run("valid config", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:          "postgres",
			Host:            "localhost",
			Port:            5432,
			Database:        "station_manager",
			User:            "smuser",
			Password:        "password",
			SSLMode:         "disable",
			MaxOpenConns:    2,
			MaxIdleConns:    2,
			ConnMaxLifetime: 10,
			ConnMaxIdleTime: 5,
			ContextTimeout:  5,
		}
		svc := &Service{config: cfg}
		err := svc.Initialize()
		assert.NoError(t, err)
	})
	t.Run("invalid driver", func(t *testing.T) {
		cfg := &types.DatastoreConfig{
			Driver:          "invalid",
			Host:            "localhost",
			Port:            5432,
			Database:        "station_manager",
			User:            "smuser",
			Password:        "password",
			SSLMode:         "disable",
			MaxOpenConns:    2,
			MaxIdleConns:    2,
			ConnMaxLifetime: 10,
			ConnMaxIdleTime: 5,
			ContextTimeout:  5,
		}
		svc := &Service{config: cfg}
		err := svc.Initialize()
		assert.Error(t, err)
		assert.Equal(t, errMsgConfigInvalid, err.Error())
		dErr, ok := errors.AsDetailedError(err)
		assert.True(t, ok)
		assert.NotNil(t, dErr)
		assert.Contains(t, dErr.Cause().Error(), "Config.Driver")
	})
}
