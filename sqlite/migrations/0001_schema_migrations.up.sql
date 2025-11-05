CREATE TABLE IF NOT EXISTS qso
(
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at      DATETIME    NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    modified_at     DATETIME,
    deleted_at      DATETIME,

    call            VARCHAR(20) NOT NULL,
    band            VARCHAR(10) NOT NULL,
    mode            VARCHAR(10) NOT NULL,
    freq            INTEGER     NOT NULL,
    qso_date        DATE        NOT NULL,
    time_on         TIME        NOT NULL,
    time_off        TIME        NOT NULL,
    rst_send        VARCHAR(3)  NOT NULL,
    rst_recv        VARCHAR(3)  NOT NULL,
    country         VARCHAR(50),
    additional_data JSON                 DEFAULT ('{}'),

    CONSTRAINT qso_data_no_duplicates CHECK (
        json_extract(additional_data, '$.call') IS NULL AND
        json_extract(additional_data, '$.band') IS NULL AND
        json_extract(additional_data, '$.mode') IS NULL AND
        json_extract(additional_data, '$.freq') IS NULL AND
        json_extract(additional_data, '$.qso_date') IS NULL AND
        json_extract(additional_data, '$.time_on') IS NULL AND
        json_extract(additional_data, '$.time_off') IS NULL AND
        json_extract(additional_data, '$.rst_send') IS NULL AND
        json_extract(additional_data, '$.rst_recv') IS NULL AND
        json_extract(additional_data, '$.country') IS NULL
        )
);

CREATE INDEX IF NOT EXISTS idx_qso_call ON qso (call);
CREATE INDEX IF NOT EXISTS idx_qso_band ON qso (band);
CREATE INDEX IF NOT EXISTS idx_qso_country ON qso (country);
CREATE INDEX IF NOT EXISTS idx_qso_date_time ON qso (qso_date, time_on);
