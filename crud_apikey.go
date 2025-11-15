package database

import (
	"context"
	"database/sql"
	stderr "errors"
	"github.com/Station-Manager/adapters"
	"github.com/Station-Manager/database/postgres/models"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func (s *Service) InsertAPIKey(name, prefix, hash string, logbookID int64) error {
	const op errors.Op = "database.Service.InsertAPIKey"
	if err := checkService(op, s); err != nil {
		return errors.New(op).Err(err)
	}
	ctx := context.Background()
	return s.InsertAPIKeyContext(ctx, name, prefix, hash, logbookID)
}

func (s *Service) InsertAPIKeyContext(ctx context.Context, name, prefix, hash string, logbookID int64) error {
	const op errors.Op = "database.Service.InsertAPIKeyContext"
	if name == "" || prefix == "" || hash == "" {
		return errors.New(op).Msg("Name, prefix, and hash are required")
	}

	if logbookID < 1 {
		return errors.New(op).Msg("Logbook ID not set")
	}
	model := models.APIKey{
		LogbookID: logbookID,
		KeyName:   name,
		KeyHash:   hash,
		KeyPrefix: prefix,
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()
	if h == nil || !isOpen {
		return errors.New(op).Msg(errMsgNotOpen)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	if err := model.Insert(ctx, h, boil.Infer()); err != nil {
		return errors.New(op).Err(err)
	}

	return nil
}

func (s *Service) FetchAPIKeyByPrefix(prefix string) (types.ApiKey, error) {
	const op errors.Op = "database.Service.FetchAPIKeyByPrefix"
	ctx := context.Background()
	return s.FetchAPIKeyByPrefixContext(ctx, prefix)
}

func (s *Service) FetchAPIKeyByPrefixContext(ctx context.Context, prefix string) (types.ApiKey, error) {
	const op errors.Op = "database.Service.FetchAPIKeyByPrefix"
	emptyRetVal := types.ApiKey{}
	if err := checkService(op, s); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()
	if h == nil || !isOpen {
		return emptyRetVal, errors.New(op).Msg(errMsgNotOpen)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	model, err := models.APIKeys(models.APIKeyWhere.KeyPrefix.EQ(prefix)).One(ctx, h)
	if err != nil && !stderr.Is(err, sql.ErrNoRows) {
		return emptyRetVal, errors.New(op).Err(err)
	}
	if model == nil || err != nil {
		return emptyRetVal, errors.New(op).Err(err).Errorf("prefix not found: %s", prefix)
	}

	var apiKey types.ApiKey
	adapter := adapters.New()
	if err = adapter.Into(&apiKey, model); err != nil {
		return emptyRetVal, errors.New(op).Err(err)
	}

	return apiKey, nil
}
