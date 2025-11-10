CREATE TABLE IF NOT EXISTS api_keys
(
    id           BIGSERIAL PRIMARY KEY,
    key_name     VARCHAR(255) NOT NULL UNIQUE,
    key_hash     VARCHAR(128) NOT NULL,
    key_prefix   VARCHAR(10)  NOT NULL,

    scopes       TEXT[]                DEFAULT '{}',
    allowed_ips  INET[],

    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ,

    created_by   VARCHAR(255),
    revoked_by   VARCHAR(255),
    use_count    BIGINT                DEFAULT 0,

    CONSTRAINT api_keys_revoked_before_or_at_expires
        CHECK (revoked_at IS NULL OR expires_at IS NULL OR revoked_at <= expires_at)
);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_prefix ON api_keys (key_prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys (revoked_at) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys (expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS logbook
(
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified_at TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,

    name VARCHAR(20) NOT NULL UNIQUE,
    callsign VARCHAR(20) NOT NULL,
    description VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS qso
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified_at     TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,

    call            VARCHAR(20) NOT NULL,
    band            VARCHAR(10) NOT NULL,
    mode            VARCHAR(10) NOT NULL,
    freq            BIGINT      NOT NULL CHECK (freq >= 0 AND freq <= 99999999),
    qso_date        DATE        NOT NULL,
    time_on         TIME        NOT NULL,
    time_off        TIME        NOT NULL,
    rst_sent        VARCHAR(3)  NOT NULL,
    rst_rcvd        VARCHAR(3)  NOT NULL,
    country         VARCHAR(50),
    additional_data JSONB       NOT NULL DEFAULT '{}'::jsonb,

    logbook_id      BIGINT      NOT NULL,

    -- Ensure additional_data does not duplicate main columns
    CONSTRAINT qso_additional_no_overlap CHECK (
        NOT (additional_data ?| array [
            'call','band','mode','freq','qso_date',
            'time_on','time_off','rst_sent','rst_rcvd','country'
            ])
        ),
    CONSTRAINT qso_logbook_fk FOREIGN KEY (logbook_id) REFERENCES logbook (id)
);

CREATE INDEX IF NOT EXISTS idx_qso_call ON qso (call);
CREATE INDEX IF NOT EXISTS idx_qso_band ON qso (band);
CREATE INDEX IF NOT EXISTS idx_qso_country ON qso (country);
CREATE INDEX IF NOT EXISTS idx_qso_date_time ON qso (qso_date, time_on);
CREATE INDEX IF NOT EXISTS idx_qso_additional_gin ON qso USING gin (additional_data);

CREATE OR REPLACE VIEW api_keys_status AS
SELECT id,
       key_name,
       key_prefix,
       created_at,
       last_used_at,
       expires_at,
       revoked_at,
       (revoked_at IS NULL AND (expires_at IS NULL OR expires_at > now())) AS is_active
FROM api_keys;
