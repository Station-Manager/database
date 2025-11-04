package database

import (
	"context"
	"database/sql"
)

type Database interface {
	Open() error
	Close() error
	Ping() error
	Migrate() error
	BeginTxContext(context.Context) (*sql.Tx, context.CancelFunc, error)
}
