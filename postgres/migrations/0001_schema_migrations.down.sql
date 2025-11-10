-- Drop view first
DROP VIEW IF EXISTS api_keys_status;

-- Drop api_keys indexes then table
DROP INDEX IF EXISTS idx_api_keys_key_prefix;
DROP INDEX IF EXISTS idx_api_keys_active;
DROP INDEX IF EXISTS idx_api_keys_expires_at;
DROP TABLE IF EXISTS api_keys;

-- Drop qso indexes then table (qso has FK to logbook, so drop it first)
DROP INDEX IF EXISTS idx_qso_call;
DROP INDEX IF EXISTS idx_qso_band;
DROP INDEX IF EXISTS idx_qso_country;
DROP INDEX IF EXISTS idx_qso_date_time;
DROP INDEX IF EXISTS idx_qso_additional_gin;
DROP TABLE IF EXISTS qso;

-- Drop logbook table last
DROP TABLE IF EXISTS logbook;