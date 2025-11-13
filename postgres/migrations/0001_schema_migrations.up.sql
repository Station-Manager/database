-- Initialization schema for Station Manager (PostgreSQL)
-- High-level design: sequential internal IDs, opaque external UIDs, API keys per logbook

-- Logbook: internal PK for joins + opaque UID for external reference
CREATE TABLE IF NOT EXISTS logbook
(
    id          BIGSERIAL   PRIMARY KEY,
    uid         UUID        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified_at TIMESTAMPTZ,

    name        VARCHAR(64)  NOT NULL UNIQUE,
    callsign    VARCHAR(32)  NOT NULL,
    description VARCHAR(255),

    CONSTRAINT logbook_uid_unique UNIQUE (uid)
);

-- API keys: per-logbook keys with independent random prefix and hashed secret
CREATE TABLE IF NOT EXISTS api_keys
(
    id           BIGSERIAL   PRIMARY KEY,
    logbook_id   BIGINT      NOT NULL,

    key_name     VARCHAR(255) NOT NULL,
    key_hash     VARCHAR(128) NOT NULL,
    key_prefix   VARCHAR(16)  NOT NULL,

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
        CHECK (revoked_at IS NULL OR expires_at IS NULL OR revoked_at <= expires_at),

    -- key names are unique within a logbook (not globally)
    CONSTRAINT api_keys_name_per_logbook UNIQUE (logbook_id, key_name),

    -- explicit FK for clarity
    CONSTRAINT api_keys_logbook_fk FOREIGN KEY (logbook_id) REFERENCES logbook(id) ON DELETE CASCADE
);

-- Helpful indexes for API key lookups and status
CREATE INDEX IF NOT EXISTS idx_api_keys_logbook_prefix ON api_keys (logbook_id, key_prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys (revoked_at) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys (expires_at) WHERE expires_at IS NOT NULL;
-- Enforce only one active (non-revoked) key per logbook
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_one_active_per_logbook
  ON api_keys (logbook_id)
  WHERE revoked_at IS NULL;

-- QSO: linked to logbook; key integrity enforced at application layer
CREATE TABLE IF NOT EXISTS qso
(
    id              BIGSERIAL   PRIMARY KEY,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified_at     TIMESTAMPTZ,

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
    CONSTRAINT qso_logbook_fk FOREIGN KEY (logbook_id) REFERENCES logbook (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_qso_call ON qso (call);
CREATE INDEX IF NOT EXISTS idx_qso_band ON qso (band);
CREATE INDEX IF NOT EXISTS idx_qso_country ON qso (country);
CREATE INDEX IF NOT EXISTS idx_qso_date_time ON qso (qso_date, time_on);
CREATE INDEX IF NOT EXISTS idx_qso_additional_gin ON qso USING gin (additional_data);

-- Status view for API keys (includes per-logbook context)
CREATE OR REPLACE VIEW api_keys_status AS
SELECT id,
       logbook_id,
       key_name,
       key_prefix,
       created_at,
       last_used_at,
       expires_at,
       revoked_at,
       (revoked_at IS NULL AND (expires_at IS NULL OR expires_at > now())) AS is_active
FROM api_keys;
