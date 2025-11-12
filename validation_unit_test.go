package database

import (
	"github.com/Station-Manager/types"
	"testing"
)

func TestValidateConfig_SqliteDefaults(t *testing.T) {
	cfg := &types.DatastoreConfig{
		Driver:                    SqliteDriver,
		Path:                      "/tmp/test.db",
		Options:                   map[string]string{"_journal_mode": "WAL", "_busy_timeout": "5000", "_foreign_keys": "on"},
		MaxOpenConns:              4,
		MaxIdleConns:              4,
		ConnMaxLifetime:           0,
		ConnMaxIdleTime:           5,
		ContextTimeout:            5,
		TransactionContextTimeout: 10,
	}
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig(sqlite) returned error: %v", err)
	}
}

func TestValidateConfig_PostgresSmallPools(t *testing.T) {
	cfg := &types.DatastoreConfig{
		Driver:                    PostgresDriver,
		Host:                      "localhost",
		Port:                      5432,
		Database:                  "testdb",
		User:                      "testuser",
		Password:                  "testpass",
		SSLMode:                   "disable",
		MaxOpenConns:              2, // allowed by tags (min=1)
		MaxIdleConns:              2, // allowed by tags (min=1)
		ConnMaxLifetime:           30,
		ConnMaxIdleTime:           5,
		ContextTimeout:            5,
		TransactionContextTimeout: 15,
	}
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig(postgres) returned error: %v", err)
	}
}
