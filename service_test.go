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
		svc, err := New(nil)
		assert.Error(t, err)
		assert.Nil(t, svc)
		assert.Equal(t, "Config parameter is nil.", err.Error())
	})
	t.Run("invalid config", func(t *testing.T) {
		cfg := &types.Config{}
		svc, err := New(cfg)
		assert.Error(t, err)
		assert.Nil(t, svc)
		assert.Equal(t, "Database configuration is invalid.", err.Error())
		dErr, ok := errors.AsDetailedError(err)
		assert.True(t, ok)
		assert.NotNil(t, dErr)
		fmt.Println(dErr.Cause())
	})
	t.Run("valid config", func(t *testing.T) {
		cfg := &types.Config{
			Driver: "postgres",
			DSN:    "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		}
		svc, err := New(cfg)
		assert.Error(t, err)
		assert.Nil(t, svc)
		assert.Equal(t, "Database configuration is invalid.", err.Error())
		dErr, ok := errors.AsDetailedError(err)
		assert.True(t, ok)
		assert.NotNil(t, dErr)
		fmt.Println(dErr.Cause())
	})
}
