PRAGMA foreign_keys = OFF;

DROP INDEX IF EXISTS idx_api_keys_expires_at;
DROP INDEX IF EXISTS idx_api_keys_active;
DROP INDEX IF EXISTS idx_api_keys_key_prefix;
DROP TABLE IF EXISTS api_keys;

PRAGMA foreign_keys = ON;
