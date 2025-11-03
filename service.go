package database

import (
	"context"
	"database/sql"
	types "github.com/Station-Manager/types/database"
)

type Service struct {
	db     *sql.DB
	driver string
}

func New(cfg *types.Config) (*Service, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Service) Initialize() error {
	return nil
}

func (s *Service) Open() error {
	return nil
}

func (s *Service) Close() error {
	return nil
}

func (s *Service) Ping() error {
	return nil
}

func (s *Service) Migrate() error {
	return nil
}

func (s *Service) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

func (s *Service) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

func (s *Service) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}
