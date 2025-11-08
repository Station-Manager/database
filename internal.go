package database

import (
	"context"
	"fmt"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/utils"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// getDsn returns the DSN for the database identified in the config.
func (s *Service) getDsn() (string, error) {
	const op errors.Op = "database.Service.getDsn"
	switch s.DatabaseConfig.Driver {
	case PostgresDriver:
		userInfo := url.UserPassword(s.DatabaseConfig.User, s.DatabaseConfig.Password)
		u := &url.URL{
			Scheme:   "postgres",
			User:     userInfo,
			Host:     fmt.Sprintf("%s:%d", s.DatabaseConfig.Host, s.DatabaseConfig.Port),
			Path:     "/" + s.DatabaseConfig.Database,
			RawQuery: url.Values{"sslmode": {s.DatabaseConfig.SSLMode}}.Encode(),
		}
		return u.String(), nil
	case SqliteDriver:
		path := s.DatabaseConfig.Path
		if path == emptyString {
			return emptyString, errors.New(op).Msg(errMsgEmptyPath)
		}

		opts := s.DatabaseConfig.Options

		// Normalize: strip leading '?' if present
		if len(opts) > 0 && opts[0] == '?' {
			opts = opts[1:]
		}

		dsn := fmt.Sprintf("file:%s?mode=rwc&_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000", path)
		return dsn, nil
	default:
		return emptyString, errors.New(op).Errorf("Unsupported database driver: %s (expected %q or %q)", s.DatabaseConfig.Driver, PostgresDriver, SqliteDriver)
	}
}

// withDefaultTimeout returns a context with a default timeout if none is provided.
// If the context is nil, a new context with a default timeout is returned.
// If the context is already cancelled or has a deadline, it is returned as-is.
func (s *Service) withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), time.Duration(s.DatabaseConfig.ContextTimeout)*time.Second)
	}
	if ctx.Err() != nil {
		cctx, cancel := context.WithCancel(ctx)
		return cctx, cancel
	}
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, time.Duration(s.DatabaseConfig.ContextTimeout)*time.Second)
}

// checkDatabaseDir checks if the database directory exists, and if not, creates it.
func (s *Service) checkDatabaseDir(dbFilePath string) error {
	const op errors.Op = "database.Service.checkDatabaseFile"
	if len(dbFilePath) == 0 {
		return errors.New(op).Msg(errMsgEmptyPath)
	}

	exists, err := utils.PathExists(dbFilePath)
	if err != nil {
		return errors.New(op).Errorf("utils.PathExists: %w", err)
	}
	if exists {
		return nil
	}

	dbDir := filepath.Dir(dbFilePath)
	if err = os.MkdirAll(dbDir, 0755); err != nil {
		return errors.New(op).Errorf("os.MkdirAll: %w", err)
	}

	return nil
}
