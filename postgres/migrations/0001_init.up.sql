-- 0001 up: Initialization schema for Station Manager (PostgreSQL)
-- Order: users -> logbook -> api_keys -> qso -> views

-- Users must be created first (referenced by logbook)
CREATE TABLE IF NOT EXISTS users
(
    id                   BIGSERIAL PRIMARY KEY,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    modified_at          TIMESTAMPTZ,

    callsign             VARCHAR(32) NOT NULL UNIQUE,
    pass_hash            VARCHAR(255), -- Log in password hash

    -- Optional external identity (all nullable)
    issuer               TEXT,
    subject              TEXT,
    email                VARCHAR(256),
    email_confirmed      BOOLEAN               DEFAULT FALSE,

    -- External identity must be fully NULL or fully present
    CONSTRAINT users_issuer_subject_pair CHECK (
        (issuer IS NULL AND subject IS NULL) OR
        (issuer IS NOT NULL AND subject IS NOT NULL)
        ),

    -- Uniqueness for external identities (PostgreSQL treats NULLs as distinct)
    CONSTRAINT users_external_identity_unique UNIQUE (issuer, subject)
);

-- Index for external identity lookups
CREATE INDEX IF NOT EXISTS idx_users_issuer_subject
    ON users (issuer, subject);


-- Logbook references users
CREATE TABLE IF NOT EXISTS logbook
(
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified_at TIMESTAMPTZ,

    user_id     BIGINT      NOT NULL, -- FK to users
    name        VARCHAR(64) NOT NULL, -- Name of the logbook
    callsign    VARCHAR(32) NOT NULL, -- Callsign associated with the logbook
    description VARCHAR(255),

    CONSTRAINT logbook_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT logbook_user_name_unique UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_logbook_user_id ON logbook (user_id);

-- API keys reference logbook
CREATE TABLE IF NOT EXISTS api_keys
(
    id           BIGSERIAL PRIMARY KEY,
    logbook_id   BIGINT       NOT NULL, -- FK to logbook

    key_name     VARCHAR(255) NOT NULL, -- The name of the logbook associated with this key
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
    CONSTRAINT api_keys_logbook_fk FOREIGN KEY (logbook_id) REFERENCES logbook (id) ON DELETE CASCADE
);

-- Helpful indexes for API key lookups and status
CREATE INDEX IF NOT EXISTS idx_api_keys_logbook_prefix ON api_keys (logbook_id, key_prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys (revoked_at) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys (expires_at) WHERE expires_at IS NOT NULL;

-- Enforce only one active (non-revoked) key per logbook
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_one_active_per_logbook
    ON api_keys (logbook_id)
    WHERE revoked_at IS NULL;

-- QSO references logbook
CREATE TABLE IF NOT EXISTS qso
(
    id              BIGSERIAL PRIMARY KEY,
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
