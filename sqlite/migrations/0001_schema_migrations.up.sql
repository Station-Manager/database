PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS logbook
(
    id          INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at  DATETIME NOT NULL DEFAULT (datetime('now', 'localtime')),
    modified_at DATETIME,
    deleted_at  DATETIME,

    name        TEXT     NOT NULL UNIQUE CHECK (length(name) <= 20),
    /*
        The callsign associated with the logbook. This should be treated as the 'station_callsign' according to the
        ADIF standard (the one used 'on the air').
    */
    callsign    TEXT     NOT NULL CHECK (length(callsign) <= 20),
    api_key     TEXT     NOT NULL UNIQUE CHECK (length(api_key) <= 20),
    description TEXT
);

-- Seed a default logbook so newly initialized databases have a usable logbook.
-- Use INSERT OR IGNORE so migrations are idempotent.
INSERT OR IGNORE INTO logbook (name, callsign, description) VALUES ('default', 'NOCALL', 'Default logbook created by migrations');

CREATE TABLE IF NOT EXISTS qso
(
    id              INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,
    created_at      DATETIME NOT NULL DEFAULT (datetime('now', 'localtime')),
    modified_at     DATETIME,
    deleted_at      DATETIME,

    call            TEXT     NOT NULL CHECK (length(call) <= 20),
    band            TEXT     NOT NULL CHECK (length(band) <= 10),
    mode            TEXT     NOT NULL CHECK (length(mode) <= 10),
    /* freq is in kHz, thus app MUST multiply by 1000 for MHz */
    freq            INTEGER  NOT NULL CHECK (freq >= 0 AND freq <= 99999999),
    /* DATETIME here prompts SQLBoiler to use the "time.Time" type */
    qso_date        TEXT     NOT NULL CHECK (
        length(qso_date) = 8 AND
        substr(qso_date, 1, 4) BETWEEN '0000' AND '9999' AND
        substr(qso_date, 5, 2) BETWEEN '01' AND '12' AND
        substr(qso_date, 7, 2) BETWEEN '01' AND '31' AND
        date(substr(qso_date, 1, 4) || '-' || substr(qso_date, 5, 2) || '-' || substr(qso_date, 7, 2)) IS NOT NULL
        ),
    /* DATETIME here prompts SQLBoiler to use the "time.Time" type */
    time_on         TEXT     NOT NULL CHECK (
        (length(time_on) = 4 AND substr(time_on, 1, 2) BETWEEN '00' AND '23' AND
         substr(time_on, 3, 2) BETWEEN '00' AND '59')
        ),
    /* DATETIME here prompts SQLBoiler to use the "time.Time" type */
    time_off        TEXT     NOT NULL CHECK (
        (length(time_off) = 4 AND substr(time_off, 1, 2) BETWEEN '00' AND '23' AND
         substr(time_off, 3, 2) BETWEEN '00' AND '59')
        ),
    rst_sent        TEXT     NOT NULL CHECK (length(rst_sent) <= 3),
    rst_rcvd        TEXT     NOT NULL CHECK (length(rst_rcvd) <= 3),
    country         TEXT CHECK (length(country) <= 50),
    additional_data JSON     NOT NULL DEFAULT ('{}') CHECK (json_valid(additional_data)),

    logbook_id      INTEGER  NOT NULL,

    CONSTRAINT qso_data_no_duplicates CHECK (
        json_extract(additional_data, '$.call') IS NULL AND
        json_extract(additional_data, '$.band') IS NULL AND
        json_extract(additional_data, '$.mode') IS NULL AND
        json_extract(additional_data, '$.freq') IS NULL AND
        json_extract(additional_data, '$.qso_date') IS NULL AND
        json_extract(additional_data, '$.time_on') IS NULL AND
        json_extract(additional_data, '$.time_off') IS NULL AND
        json_extract(additional_data, '$.rst_sent') IS NULL AND
        json_extract(additional_data, '$.rst_rcvd') IS NULL AND
        json_extract(additional_data, '$.country') IS NULL
        ),
    -- Explicit foreign key behavior: prevent deleting a logbook that still has QSOs. Change to ON DELETE CASCADE if you prefer automatic cleanup.
    CONSTRAINT fk_qso_logbook FOREIGN KEY (logbook_id) REFERENCES logbook (id) ON DELETE RESTRICT ON UPDATE NO ACTION
);

CREATE INDEX IF NOT EXISTS idx_qso_call ON qso (call);
CREATE INDEX IF NOT EXISTS idx_qso_band ON qso (band);
CREATE INDEX IF NOT EXISTS idx_qso_country ON qso (country);
CREATE INDEX IF NOT EXISTS idx_qso_date_time ON qso (qso_date, time_on);
-- Index on the FK column for joins/deletes
CREATE INDEX IF NOT EXISTS idx_qso_logbook_id ON qso (logbook_id);

-- Optional: index freq if you query by frequency often
/*
CREATE INDEX IF NOT EXISTS idx_qso_freq ON qso (freq);
*/

-- Optional: partial indexes to speed queries that ignore soft-deleted rows
/*
CREATE INDEX IF NOT EXISTS idx_qso_active_call ON qso (call) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_qso_active_date_time ON qso (qso_date, time_on) WHERE deleted_at IS NULL;
*/

-- Optional: enforce uniqueness for active QSOs (example — adjust columns to your deduplication rules)
-- CREATE UNIQUE INDEX IF NOT EXISTS uq_qso_active_unique ON qso (call, qso_date, time_on, freq) WHERE deleted_at IS NULL;

-- Trigger to set modified_at on updates (safe pattern: update the row after the user's update)
CREATE TRIGGER IF NOT EXISTS trg_qso_set_modified_at
    AFTER UPDATE
    ON qso
    FOR EACH ROW
    WHEN NEW.modified_at IS NULL OR NEW.modified_at = OLD.modified_at
BEGIN
    UPDATE qso SET modified_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
