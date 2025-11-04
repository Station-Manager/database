package database

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

func (s *Service) getDsn() string {
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
		return u.String()
	case SqliteDriver:
		path := s.config.Path
		opts := s.config.Options
		if opts == "" {
			opts = "_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on&cache=shared&_txlock=immediate"
		} else if opts[0] == '?' {
			opts = opts[1:]
		}
		u := &url.URL{
			Scheme:   "file",
			Path:     path,
			RawQuery: opts,
		}
		return u.String()
	default:
		return ""
	}
}

func (s *Service) withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, time.Duration(s.config.ContextTimeout)*time.Second)
}
