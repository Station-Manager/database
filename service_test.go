package database

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		svc, err := New(nil)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})
}
