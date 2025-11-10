PRAGMA foreign_keys = OFF;

DROP TRIGGER IF EXISTS trg_qso_set_modified_at;

DROP INDEX IF EXISTS idx_qso_date_time;
DROP INDEX IF EXISTS idx_qso_country;
DROP INDEX IF EXISTS idx_qso_band;
DROP INDEX IF EXISTS idx_qso_call;
DROP TABLE IF EXISTS qso;

DROP TABLE IF EXISTS logbook;

PRAGMA foreign_keys = ON;
