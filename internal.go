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
		if path == "" {
			return "", errors.New(op).Msg(errMsgEmptyPath)
		}

		opts := s.config.Options

		// Normalize: strip leading '?' if present
		if len(opts) > 0 && opts[0] == '?' {
			opts = opts[1:]
		}

		if opts == "" {
			opts = "_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on&_txlock=immediate"
		}

		u := &url.URL{
			Scheme:   "file",
			Path:     path,
			RawQuery: opts,
		}
		return u.String(), nil
	default:
		return "", errors.New(op).Errorf("Unsupported database driver: %s (expected %q or %q)", s.config.Driver, PostgresDriver, SqliteDriver)
	}
}

func (s *Service) withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, time.Duration(s.config.ContextTimeout)*time.Second)
}

// checkDatabaseFile ensures the database file exists; if not, it creates the necessary directory and file structure.
// Returns an error if any issue occurs during file validation, directory creation, or file creation.
func (s *Service) checkDatabaseFile(dbFilePath string) error {
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

	var f *os.File
	if f, err = os.OpenFile(dbFilePath, os.O_RDWR|os.O_CREATE, 0644); err != nil {
		return errors.New(op).Errorf("os.OpenFile: %w", err)
	}
	if err = f.Close(); err != nil {
		return errors.New(op).Errorf("f.Close: %w", err)
	}

	return nil
}
