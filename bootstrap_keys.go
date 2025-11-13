package database

import (
	"database/sql"
	stderr "errors"
	"github.com/Station-Manager/database/postgres/models"
	"github.com/Station-Manager/errors"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"time"
)

func (s *Service) StoreBootstrap(userID int64, hash string, expires time.Time) error {
	const op errors.Op = "database.Service.StoreBootstrap"
	if err := checkService(op, s); err != nil {
		return errors.New(op).Err(err)
	}
	if s.DatabaseConfig.Driver != PostgresDriver {
		s.Logger.WarnWith().Msg("Bootstrap keys are only supported on Postgres")
		return nil // Not supported on the client only the server
	}

	if userID < 1 {
		return errors.New(op).Msg("UserID is required")
	}

	if hash == emptyString {
		return errors.New(op).Msg("Hash is required")
	}

	if expires.IsZero() {
		return errors.New(op).Msg("Expires is required")
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()
	if h == nil || !isOpen {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	ctx, cancel := s.withDefaultTimeout(nil)
	defer cancel()

	model, err := models.FindUser(ctx, h, userID)
	if err != nil && stderr.Is(err, sql.ErrNoRows) {
		return errors.New(op).Err(err)
	}

	if err != nil || model == nil {
		localErr := errors.New(op).Err(err)
		s.Logger.ErrorWith().Err(err).Int64("userID", userID).Msg("Failed to find user for bootstrap key")
		return localErr.Msg("Internal error") // Error to report back to the client - no info leakage.
	}

	model.BootstrapHash = null.StringFrom(hash)
	model.BootstrapExpiresAt = null.TimeFrom(expires)

	_, err = model.Update(ctx, h, boil.Whitelist(models.UserColumns.BootstrapHash, models.UserColumns.BootstrapExpiresAt))
	if err != nil {
		return errors.New(op).Err(err)
	}

	return nil
}
