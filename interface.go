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
	BeginTxContext(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, context.CancelFunc, error)
}
