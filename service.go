package database

import (
	"context"
	"database/sql"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	_ "github.com/lib/pq"
	"sync"
	"sync/atomic"
	"time"
)

type Service struct {
	config *types.DatastoreConfig
	handle *sql.DB

	mu            sync.RWMutex
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

	s.isInitialized.Store(true)

	return nil
}

// Open opens the database connection.
func (s *Service) Open() error {
	const op errors.Op = "database.Service.Open"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	// Has the service been initialized?
	if !s.isInitialized.Load() {
		return errors.New(op).Msg(errMsgNotInitialized)
	}

	// Quick pre-check to see if the database is already open.
	if s.isOpen.Load() {
		return errors.New(op).Msg(errMsgAlreadyOpen)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-check under lock to avoid TOCTOU
	if s.isOpen.Load() {
		return errors.New(op).Msg(errMsgAlreadyOpen)
	}

	dsn := s.getDsn()
	db, err := sql.Open(s.config.Driver, dsn)
	if err != nil {
		return errors.New(op).Err(err).Msg("Database connection failed.")
	}

	db.SetMaxOpenConns(s.config.MaxOpenConns)
	db.SetMaxIdleConns(s.config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(s.config.ConnMaxLifetime) * time.Minute)
	db.SetConnMaxIdleTime(time.Duration(s.config.ConnMaxIdleTime) * time.Minute)

	ctx, cancel := s.withDefaultTimeout(context.Background())
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return errors.New(op).Err(err).Msg(errMsgPingFailed)
	}

	s.handle = db
	s.isOpen.Store(true)

	return nil
}

func (s *Service) Close() error {
	const op errors.Op = "database.Service.Close"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	// Quick pre-check
	if !s.isOpen.Load() {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-check under lock - TOCTOU
	if !s.isOpen.Load() {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	if err := s.handle.Close(); err != nil {
		return errors.New(op).Err(err).Msg("Failed to close database connection.")
	}

	s.handle = nil
	s.isOpen.Store(false)

	return nil
}

func (s *Service) Ping() error {
	const op errors.Op = "database.Service.Ping"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.handle == nil {
		return errors.New(op).Msg(errMsgNilHandle)
	}

	if !s.isOpen.Load() {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	ctx, cancel := s.withDefaultTimeout(context.Background())
	defer cancel()
	if err := s.handle.PingContext(ctx); err != nil {
		return errors.New(op).Err(err).Msg(errMsgPingFailed)
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
		return nil, errors.New(op).Msg(errMsgNilService)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.handle == nil || !s.isOpen.Load() {
		return nil, errors.New(op).Msg(errMsgNotOpen)
	}

	return s.handle.ExecContext(ctx, query, args...)
}

func (s *Service) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	const op errors.Op = "database.Service.QueryContext"
	if s == nil {
		return nil, errors.New(op).Msg(errMsgNilService)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.handle == nil || !s.isOpen.Load() {
		return nil, errors.New(op).Msg(errMsgNotOpen)
	}

	return s.handle.QueryContext(ctx, query, args...)
}

func (s *Service) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if s == nil {
		panic(errMsgNilService)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.handle == nil || !s.isOpen.Load() {
		panic(errMsgNotOpen)
	}

	return s.handle.QueryRowContext(ctx, query, args...)
}
