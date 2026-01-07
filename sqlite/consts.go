package sqlite

import "time"

const (
	errMsgNotInitialized = "Database service is not initialized."
	errMsgNilService     = "Database service is nil."
	errMsgNilConfig      = "Database config is nil."
	errMsgNotOpen        = "Database service is not open."
	errMsgAlreadyOpen    = "Database service is already open."
	errMsgConfigInvalid  = "Database configuration is invalid."
	errMsgAppConfigNil   = "Config service is nil."
	errMsgLoggerNil      = "Logging service is nil."
	errMsgPingFailed     = "Failed to ping database."
	errMsgConnFailed     = "Database connection failed."
	errMsgFailedClose    = "Failed to close database connection."
	errMsgMigrateFailed  = "Failed to run database migrations."
	errMsgEmptyPath      = "SQLite path cannot be empty."
	errMsgDsnBuildError  = "Failed to build DSN."
)

const (
	errMsgInvalidId = "Invalid ID"
)

const (
	SqliteDriver = "sqlite"
	emptyString  = ""

	// pingRetryBackoff is the delay between ping retry attempts for transient errors.
	// Kept short since SQLite is local and busy_timeout PRAGMA handles most wait scenarios.
	pingRetryBackoff = 25 * time.Millisecond
)
