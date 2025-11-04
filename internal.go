package database

import (
	"context"
	"fmt"
	"time"
)

func (s *Service) getDsn() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		s.config.User, s.config.Password, s.config.Host, s.config.Port, s.config.Database, s.config.SSLMode)
}

func (s *Service) withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, time.Duration(s.config.ContextTimeout)*time.Second)
}
