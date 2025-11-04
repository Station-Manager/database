package database

import (
	"github.com/Station-Manager/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetDsn(t *testing.T) {
	t.Run("SQLite empty options gets defaults", func(t *testing.T) {
		svc := &Service{
			config: &types.DatastoreConfig{
				Driver:  SqliteDriver,
				Path:    "/tmp/test.db",
				Options: "",
			},
		}
		dsn, err := svc.getDsn()
		assert.NoError(t, err)
		assert.Contains(t, dsn, "_busy_timeout=5000")
		assert.Contains(t, dsn, "_journal_mode=WAL")
	})

	t.Run("SQLite just question mark gets defaults", func(t *testing.T) {
		svc := &Service{
			config: &types.DatastoreConfig{
				Driver:  SqliteDriver,
				Path:    "/tmp/test.db",
				Options: "?",
			},
		}
		dsn, err := svc.getDsn()
		assert.NoError(t, err)
		assert.Contains(t, dsn, "_busy_timeout=5000")
	})

	t.Run("SQLite custom options preserved", func(t *testing.T) {
		svc := &Service{
			config: &types.DatastoreConfig{
				Driver:  SqliteDriver,
				Path:    "/tmp/test.db",
				Options: "?cache=shared&mode=memory",
			},
		}
		dsn, err := svc.getDsn()
		assert.NoError(t, err)
		assert.Contains(t, dsn, "cache=shared")
		assert.NotContains(t, dsn, "_busy_timeout")
	})

	t.Run("unknown driver returns error", func(t *testing.T) {
		svc := &Service{
			config: &types.DatastoreConfig{
				Driver: "mysql",
			},
		}
		_, err := svc.getDsn()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Unsupported database driver:")
	})

	t.Run("SQLite empty path returns error", func(t *testing.T) {
		svc := &Service{
			config: &types.DatastoreConfig{
				Driver: SqliteDriver,
				Path:   "",
			},
		}
		_, err := svc.getDsn()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path cannot be empty")
	})
}
