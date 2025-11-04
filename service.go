package database

import (
	"context"
	"database/sql"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	_ "github.com/jackc/pgx/v5/stdlib"
	"sync"
	"sync/atomic"
)

type Service struct {
	config        *types.DatastoreConfig
	handle        *sql.DB
	mu            sync.Mutex
	isCfgValid    atomic.Bool
	isInitialized atomic.Bool
	isOpen        atomic.Bool
}

// Initialize initializes the database service.
func (s *Service) Initialize() error {
	const op errors.Op = "database.Service.Initialize"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	if s.config == nil {
		return errors.New(op).Msg(errMsgNilConfig)
	}

	if err := validateConfig(s.config); err != nil {
		return err
	}

	s.isCfgValid.Store(true)

	s.isInitialized.Store(true)

	return nil
}

// Open opens the database connection.
func (s *Service) Open() error {
	const op errors.Op = "database.Service.Open"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isInitialized.Load() {
		return errors.New(op).Msg(errMsgNotInitialized)
	}

	if !s.isOpen.Load() {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	dsn := s.getDsn()
	var err error
	if s.handle, err = sql.Open(s.config.Driver, dsn); err != nil {
		return errors.New(op).Err(err).Msg("Database connection failed.")
	}

	s.isOpen.Store(true)

	return nil
}

func (s *Service) Close() error {
	const op errors.Op = "database.Service.Close"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}
	return nil
}

func (s *Service) Ping() error {
	const op errors.Op = "database.Service.Ping"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}
	return nil
}

func (s *Service) Migrate() error {
	const op errors.Op = "database.Service.Migrate"
	if s == nil {
		return errors.New(op).Msg("Service is nil.")
	}
	return nil
}

func (s *Service) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	const op errors.Op = "database.Service.ExecContext"
	if s == nil {
		return nil, errors.New(op).Msg("Service is nil.")
	}
	return s.handle.ExecContext(ctx, query, args...)
}

func (s *Service) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	const op errors.Op = "database.Service.QueryContext"
	if s == nil {
		return nil, errors.New(op).Msg("Service is nil.")
	}
	return s.handle.QueryContext(ctx, query, args...)
}

func (s *Service) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if s == nil {
		panic("Service is nil.")
	}
	return s.handle.QueryRowContext(ctx, query, args...)
}
