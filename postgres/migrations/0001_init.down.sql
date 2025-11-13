-- Rollback initialization schema for Station Manager (PostgreSQL)

-- Drop view first (depends on api_keys)
DROP VIEW IF EXISTS api_keys_status;

-- Drop QSO indexes then table (QSO depends on logbook)
DROP INDEX IF EXISTS idx_qso_call;
DROP INDEX IF EXISTS idx_qso_band;
DROP INDEX IF EXISTS idx_qso_country;
DROP INDEX IF EXISTS idx_qso_date_time;
DROP INDEX IF EXISTS idx_qso_additional_gin;
DROP TABLE IF EXISTS qso;

-- Drop API key indexes then table (api_keys depends on logbook)
DROP INDEX IF EXISTS idx_api_keys_one_active_per_logbook;
DROP INDEX IF EXISTS idx_api_keys_logbook_prefix;
DROP INDEX IF EXISTS idx_api_keys_active;
DROP INDEX IF EXISTS idx_api_keys_expires_at;
DROP TABLE IF EXISTS api_keys;

-- Drop logbook indexes then table (logbook depends on users)
DROP INDEX IF EXISTS idx_logbook_user_id;
DROP TABLE IF EXISTS logbook;

-- Drop users indexes then table
DROP INDEX IF EXISTS idx_users_bootstrap_active;
DROP INDEX IF EXISTS idx_users_issuer_subject;
DROP TABLE IF EXISTS users;
