package database

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

func (s *Service) getDsn() string {
	switch s.config.Driver {
	case "postgres":
		user := url.QueryEscape(s.config.User)
		pass := url.QueryEscape(s.config.Password)
		host := s.config.Host
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			user, pass, host, s.config.Port, s.config.Database, s.config.SSLMode)
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
