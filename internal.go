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
	switch s.config.Driver {
	case PostgresDriver:
		userInfo := url.UserPassword(s.config.User, s.config.Password)
		u := &url.URL{
			Scheme:   "postgres",
			User:     userInfo,
			Host:     fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
			Path:     "/" + s.config.Database,
			RawQuery: url.Values{"sslmode": {s.config.SSLMode}}.Encode(),
		}
		return u.String(), nil
	case SqliteDriver:
		path := s.config.Path
		if path == emptyString {
			return emptyString, errors.New(op).Msg(errMsgEmptyPath)
		}

		opts := s.config.Options

		// Normalize: strip leading '?' if present
		if len(opts) > 0 && opts[0] == '?' {
			opts = opts[1:]
		}

		dsn := fmt.Sprintf("file:%s?mode=rwc&_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000", path)
		return dsn, nil
	default:
		return emptyString, errors.New(op).Errorf("Unsupported database driver: %s (expected %q or %q)", s.config.Driver, PostgresDriver, SqliteDriver)
	}
}

// withDefaultTimeout returns a context with a default timeout if none is provided.
func (s *Service) withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, time.Duration(s.config.ContextTimeout)*time.Second)
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
