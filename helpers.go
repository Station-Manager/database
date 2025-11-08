package database

import "github.com/Station-Manager/errors"

// checkService checks if the database service is not nil, has been initialized and is open.
func checkService(op errors.Op, s *Service) error {
	if s == nil {
		return errors.New(op).Msg(errMsgNilService)
	}

	if !s.isInitialized.Load() {
		return errors.New(op).Msg(errMsgNotInitialized)
	}

	if !s.isOpen.Load() {
		return errors.New(op).Msg(errMsgNotOpen)
	}
	return nil
}

func driverBasedCall(driver string, param any, sqliteCall func(any) (any, error), postgresCall func(any) (any, error)) (any, error) {
	const op errors.Op = "database.Service.driverBasedCall"
	switch driver {
	case SqliteDriver:
		return sqliteCall(param)
	case PostgresDriver:
		return postgresCall(param)
	default:
		return param, errors.New(op).Errorf("Unsupported database driver: %s", driver)
	}
}
