package database

import (
	"context"
	"database/sql"
	stderr "errors"
	"github.com/Station-Manager/config"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
	"sync"
	"sync/atomic"
	"time"
)

type Service struct {
	ConfigService *config.Service `inject:"configservice"`
	config        *types.DatastoreConfig
	handle        *sql.DB

	mu            sync.RWMutex
	isInitialized atomic.Bool
	isOpen        atomic.Bool
}

// Initialize initializes the database service. No constructor is provided as this service is to be
// initialized within an IOC/DI container.
func (s *Service) Initialize() error {
	const op errors.Op = "database.Service.Initialize"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	if s.isInitialized.Load() {
		return nil // Exit gracefully
	}

	if s.ConfigService == nil {
		return errors.New(op).Msg(errMsgAppConfigNil)
	}

	dbCfg, err := s.ConfigService.DatastoreConfig()
	if err != nil {
		return errors.New(op).Err(err)
	}

	if err = validateConfig(&dbCfg); err != nil {
		return errors.New(op).Err(err)
	}
	s.config = &dbCfg

	if s.config.Driver == SqliteDriver {
		// Ensure the database directory exists
		if err = s.checkDatabaseDir(s.config.Path); err != nil {
			return errors.New(op).Err(err)
		}
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

	// Outside the mutex as its config is read-only
	dsn, err := s.getDsn()
	if err != nil {
		return errors.New(op).Err(err).Msg(errMsgDsnBuildError)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-check under lock to avoid TOCTOU
	if s.isOpen.Load() {
		return errors.New(op).Msg(errMsgAlreadyOpen)
	}

	db, err := sql.Open(s.config.Driver, dsn)
	if err != nil {
		return errors.New(op).Err(err).Msg(errMsgConnFailed)
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

// Close closes the database connection.
func (s *Service) Close() error {
	const op errors.Op = "database.Service.Close"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	// Quick pre-check
	if !s.isOpen.Load() {
		return nil // Idempotent
		// return errors.New(op).Msg(errMsgNotOpen)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-check under lock - TOCTOU
	if !s.isOpen.Load() {
		return nil // Idempotent
		// return errors.New(op).Msg(errMsgNotOpen)
	}

	if err := s.handle.Close(); err != nil {
		return errors.New(op).Err(err).Msg(errMsgFailedClose)
	}

	s.handle = nil
	s.isOpen.Store(false)

	return nil
}

// Ping pings the database connection.
func (s *Service) Ping() error {
	const op errors.Op = "database.Service.Ping"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.handle == nil || !s.isOpen.Load() {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	ctx, cancel := s.withDefaultTimeout(context.Background())
	defer cancel()
	if err := s.handle.PingContext(ctx); err != nil {
		return errors.New(op).Err(err).Msg(errMsgPingFailed)
	}

	return nil
}

// Migrate runs the database migrations.
func (s *Service) Migrate() error {
	const op errors.Op = "database.Service.Migrate"
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.handle == nil || !s.isOpen.Load() {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	err := s.doMigrations()
	if err != nil {
		return errors.New(op).Err(err).Msg(errMsgMigrateFailed)
	}

	return nil
}

// BeginTxContext starts a new transaction.
func (s *Service) BeginTxContext(ctx context.Context) (*sql.Tx, context.CancelFunc, error) {
	const op errors.Op = "database.Service.BeginTxContext"
	if s == nil {
		return nil, nil, errors.New(op).Msg(errMsgNilService)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.handle == nil || !s.isOpen.Load() {
		return nil, nil, errors.New(op).Msg(errMsgNotOpen)
	}

	_, hasDeadline := ctx.Deadline()
	var txCtx context.Context
	var cancel context.CancelFunc

	if !hasDeadline {
		txCtx, cancel = context.WithTimeout(ctx, time.Duration(s.config.TransactionContextTimeout)*time.Second)
	} else {
		txCtx = ctx
		cancel = func() {} // No-op cancel
	}

	tx, err := s.handle.BeginTx(txCtx, nil)
	if err != nil {
		cancel()
		if stderr.Is(err, context.DeadlineExceeded) {
			return nil, nil, errors.New(op).Err(err).Msg("Transaction context timed out.")
		}
		return nil, nil, errors.New(op).Errorf("creating new transaction: %w", err)
	}

	return tx, cancel, nil
}

func (s *Service) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	const op errors.Op = "database.Service.ExecContext"
	if s == nil {
		return nil, errors.New(op).Msg(errMsgNilService)
	}

	// Holding s.mu.RLock() while performing s.handle.ExecContext(...) means the read lock is held for the duration
	// of the exec. This can block Close()/Migrate(), which need the write lock. So, we copy the *sql.DB handle under
	// the lock, release the lock, then call ExecContext so long-running ops don’t hold the lock.
	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if h == nil || !isOpen {
		return nil, errors.New(op).Msg(errMsgNotOpen)
	}

	var cancel context.CancelFunc
	_, ok := ctx.Deadline()
	if !ok {
		ctx, cancel = s.withDefaultTimeout(ctx)
	} else {
		cancel = func() {} // No-op
	}
	defer cancel()

	res, err := h.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(op).Errorf("s.handle.ExecContext: %w", err)
	}

	return res, nil
}

func (s *Service) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	const op errors.Op = "database.Service.QueryContext"
	if s == nil {
		return nil, errors.New(op).Msg(errMsgNilService)
	}

	// Holding s.mu.RLock() while performing s.handle.ExecContext(...) means the read lock is held for the duration
	// of the exec. This can block Close()/Migrate(), which need the write lock. So, we copy the *sql.DB handle under
	// the lock, release the lock, then call ExecContext so long-running ops don’t hold the lock.
	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()

	if h == nil || !isOpen {
		return nil, errors.New(op).Msg(errMsgNotOpen)
	}

	// Note: We do NOT defer cancel() here because the returned *sql.Rows needs
	// the context to remain valid while the caller iterates over rows.
	// The caller must ensure rows.Close() is called to release resources.
	// The timeout context (if created) will be automatically cleaned up after the timeout expires.
	_, ok := ctx.Deadline()
	if !ok {
		ctx, _ = s.withDefaultTimeout(ctx)
	}

	res, err := h.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(op).Errorf("s.handle.QueryContext: %w", err)
	}

	return res, nil
}
