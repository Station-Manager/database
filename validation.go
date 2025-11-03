package database

import (
	"github.com/Station-Manager/errors"
	types "github.com/Station-Manager/types/database"
)

func validateConfig(cfg *types.Config) error {
	const op errors.Op = "database.validateConfig"
	if cfg == nil {
		return errors.New(op).Msg("Config parameter is nil.")
	}

	return nil
}
