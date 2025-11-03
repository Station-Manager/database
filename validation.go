package database

import (
	"github.com/Station-Manager/errors"
	types "github.com/Station-Manager/types/database"
	"github.com/go-playground/validator/v10"
)

func validateConfig(cfg *types.Config) error {
	const op errors.Op = "database.validateConfig"
	if cfg == nil {
		return errors.New(op).Msg(errMsgNilConfig)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(cfg); err != nil {
		return errors.New(op).Err(err).Msg(errMsgConfigInvalid)
	}

	return nil
}
