package database

import (
	"context"
	"github.com/Station-Manager/errors"
	"github.com/Station-Manager/types"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func (s *Service) InsertUser(user types.User) (types.User, error) {
	const op errors.Op = "database.Service.InsertUser"
	if err := checkService(op, s); err != nil {
		return user, errors.New(op).Err(err)
	}
	ctx := context.Background()
	return s.InsertUserContext(ctx, user)
}

func (s *Service) InsertUserContext(ctx context.Context, user types.User) (types.User, error) {
	const op errors.Op = "database.Service.InsertUserContext"
	if err := checkService(op, s); err != nil {
		return user, errors.New(op).Err(err)
	}

	s.mu.RLock()
	h := s.handle
	isOpen := s.isOpen.Load()
	s.mu.RUnlock()
	if h == nil || !isOpen {
		return user, errors.New(op).Msg(errMsgNotOpen)
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = s.withDefaultTimeout(ctx)
		defer cancel()
	}

	s.initAdapters()
	model, err := s.AdaptTypeToPostgresModelUser(user)
	if err != nil {
		return user, errors.New(op).Err(err)
	}

	if err = model.Insert(ctx, h, boil.Infer()); err != nil {
		return user, errors.New(op).Err(err)
	}

	user.ID = model.ID
	return user, nil
}
