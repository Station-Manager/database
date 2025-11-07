CREATE TABLE IF NOT EXISTS qso
(
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at      DATETIME    NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    modified_at     DATETIME,
    deleted_at      DATETIME,

    call            VARCHAR(20) NOT NULL,
    band            VARCHAR(10) NOT NULL,
    mode            VARCHAR(10) NOT NULL,
    freq            REAL        NOT NULL CHECK (freq >= 0 AND freq <= 999.999 AND round(freq * 1000) = freq * 1000),
    /* DATETIME here prompts SQLBoiler to use the "time.Time" type */
    qso_date        DATETIME    NOT NULL CHECK (
        length(qso_date) = 8 AND
        substr(qso_date, 1, 4) BETWEEN '0000' AND '9999' AND
        substr(qso_date, 5, 2) BETWEEN '01' AND '12' AND
        substr(qso_date, 7, 2) BETWEEN '01' AND '31' AND
        date(substr(qso_date, 1, 4) || '-' || substr(qso_date, 5, 2) || '-' || substr(qso_date, 7, 2)) IS NOT NULL
        ),
    /* DATETIME here prompts SQLBoiler to use the "time.Time" type */
    time_on         DATETIME    NOT NULL CHECK (
        (length(time_on) = 4 AND substr(time_on, 1, 2) BETWEEN '00' AND '23' AND
         substr(time_on, 3, 2) BETWEEN '00' AND '59')
        ),
    /* DATETIME here prompts SQLBoiler to use the "time.Time" type */
    time_off        DATETIME    NOT NULL CHECK (
        (length(time_off) = 4 AND substr(time_off, 1, 2) BETWEEN '00' AND '23' AND
         substr(time_off, 3, 2) BETWEEN '00' AND '59')
        ),
    rst_sent        VARCHAR(3)  NOT NULL,
    rst_rcvd        VARCHAR(3)  NOT NULL,
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
