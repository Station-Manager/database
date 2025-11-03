package database

import "fmt"

func (s *Service) getDsn() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		s.config.User, s.config.Password, s.config.Host, s.config.Port, s.config.Database, s.config.SSLMode)
}
