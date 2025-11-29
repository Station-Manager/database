PRAGMA foreign_keys = OFF;

-- Drop triggers
DROP TRIGGER IF EXISTS trg_qso_set_modified_at;

-- Drop indexes created on qso
DROP INDEX IF EXISTS idx_qso_active_date_time;
DROP INDEX IF EXISTS idx_qso_active_call;
DROP INDEX IF EXISTS idx_qso_logbook_id;
DROP INDEX IF EXISTS idx_qso_date_time;
DROP INDEX IF EXISTS idx_qso_country;
DROP INDEX IF EXISTS idx_qso_band;
DROP INDEX IF EXISTS idx_qso_call;

-- Drop indexes created on logbook
DROP INDEX IF EXISTS uq_logbook_api_key;

-- Drop indexes created on contacted_station
DROP INDEX IF EXISTS idx_contacted_station_callsign;

-- Drop tables
DROP TABLE IF EXISTS contacted_station;
DROP TABLE IF EXISTS qso;
DROP TABLE IF EXISTS logbook;
DROP TABLE IF EXISTS session;

PRAGMA foreign_keys = ON;
