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
		q := url.Values{}
		if s.DatabaseConfig.SSLMode != "" {
			q.Set("sslmode", s.DatabaseConfig.SSLMode)
		}
		u := &url.URL{
			Scheme:   "postgres",
			User:     userInfo,
			Host:     fmt.Sprintf("%s:%d", s.DatabaseConfig.Host, s.DatabaseConfig.Port),
			Path:     "/" + s.DatabaseConfig.Database,
			RawQuery: q.Encode(),
		}
		return u.String(), nil

	case SqliteDriver:
		path := s.DatabaseConfig.Path
		if path == emptyString {
			return emptyString, errors.New(op).Msg(errMsgEmptyPath)
		}

		q := url.Values{}
		for k, v := range s.DatabaseConfig.Options {
			q.Set(k, v)
		}

		if len(q) == 0 {
			return fmt.Sprintf("file:%s", path), nil
		}
		return fmt.Sprintf("file:%s?%s", path, q.Encode()), nil

	default:
		return emptyString, errors.New(op).Errorf("Unsupported database driver: %s (expected %q or %q)", s.DatabaseConfig.Driver, PostgresDriver, SqliteDriver)
	}
}

// withDefaultTimeout returns a context with a default timeout if none is provided.
// If the context is nil, a new context with a default timeout is returned.
// If the context is already cancelled or has a deadline, it is returned as-is.
func (s *Service) withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := time.Duration(s.DatabaseConfig.ContextTimeout) * time.Second

	if ctx == nil {
		if timeout > 0 {
			return context.WithTimeout(context.Background(), timeout)
		}
		return context.WithCancel(context.Background())
	}
	if ctx.Err() != nil {
		return ctx, func() {}
	}
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {}
	}
	if timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

// checkDatabaseDir checks if the database directory exists, and if not, creates it.
func (s *Service) checkDatabaseDir(dbFilePath string) error {
	const op errors.Op = "database.Service.checkDatabaseDir"

	if len(dbFilePath) == 0 {
		return errors.New(op).Msg(errMsgEmptyPath)
	}

	dbDir := filepath.Dir(dbFilePath)

	exists, err := utils.PathExists(dbDir)
	if err != nil {
		return errors.New(op).Errorf("utils.PathExists: %w", err)
	}
	if exists {
		return nil
	}

	if err = os.MkdirAll(dbDir, 0o755); err != nil {
		return errors.New(op).Errorf("os.MkdirAll: %w", err)
	}

	return nil
}
