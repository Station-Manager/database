BEGIN;

-- drop view first
DROP VIEW IF EXISTS api_keys_status;

-- drop qso indexes then table
DROP INDEX IF EXISTS idx_qso_additional_gin;
DROP INDEX IF EXISTS idx_qso_date_time;
DROP INDEX IF EXISTS idx_qso_country;
DROP INDEX IF EXISTS idx_qso_band;
DROP INDEX IF EXISTS idx_qso_call;
DROP TABLE IF EXISTS qso;

-- drop api_keys indexes then table
DROP INDEX IF EXISTS idx_api_keys_expires_at;
DROP INDEX IF EXISTS idx_api_keys_active;
DROP INDEX IF EXISTS idx_api_keys_key_prefix;
DROP TABLE IF EXISTS api_keys;

COMMIT;
