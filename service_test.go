package database

import (
	"fmt"
	"github.com/Station-Manager/errors"
	types "github.com/Station-Manager/types/database"
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
		cfg := &types.Config{}
		svc := &Service{config: cfg}
		err := svc.Initialize()
		assert.Error(t, err)
		assert.Equal(t, errMsgConfigInvalid, err.Error())
		dErr, ok := errors.AsDetailedError(err)
		assert.True(t, ok)
		assert.NotNil(t, dErr)
		fmt.Println(dErr.Cause())
	})
	t.Run("valid config", func(t *testing.T) {
		cfg := &types.Config{
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
		}
		svc := &Service{config: cfg}
		err := svc.Initialize()
		assert.NoError(t, err)
	})
}
