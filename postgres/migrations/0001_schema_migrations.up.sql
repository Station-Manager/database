CREATE TABLE IF NOT EXISTS api_keys
(
    id           BIGSERIAL PRIMARY KEY,
    key_name     VARCHAR(255) NOT NULL UNIQUE,       -- Human-readable identifier
    key_hash     VARCHAR(128) NOT NULL,              -- bcrypt hash (max 128 chars)
    key_prefix   VARCHAR(10)  NOT NULL,              -- First 8 chars for identification (e.g., "7qsm_abc")

    -- Access control
    scopes       TEXT[]                DEFAULT '{}', -- e.g., {'qso:write', 'qso:read'}
    allowed_ips  INET[],                             -- Optional IP whitelist

    -- Lifecycle management
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,                        -- Optional expiration
    revoked_at   TIMESTAMPTZ,                        -- Soft delete

    -- Auditing
    created_by   VARCHAR(255),                       -- Who created the key
    revoked_by   VARCHAR(255),                       -- Who revoked it
    use_count    BIGINT                DEFAULT 0,    -- Usage tracking

    -- Constraints
    CONSTRAINT check_not_revoked_and_expired
        CHECK (revoked_at IS NULL OR expires_at IS NULL OR revoked_at >= expires_at)
);

-- Indexes for performance
CREATE INDEX idx_api_keys_key_prefix ON api_keys (key_prefix);
CREATE INDEX idx_api_keys_active ON api_keys (revoked_at) WHERE revoked_at IS NULL;
CREATE INDEX idx_api_keys_expires_at ON api_keys (expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS qso
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    modified_at     TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ, -- Soft delete

    -- Frequently queried fields as columns
    call            VARCHAR(20)   NOT NULL,
    band            VARCHAR(10)   NOT NULL,
    mode            VARCHAR(10)   NOT NULL,
    freq            NUMERIC(6, 3) NOT NULL,
    qso_date        DATE          NOT NULL,
    time_on         TIME          NOT NULL,
    time_off        TIME          NOT NULL,
    rst_sent        VARCHAR(3)    NOT NULL,
    rst_rcvd        VARCHAR(3)    NOT NULL,
    country         VARCHAR(50),

    -- Everything else (name, QTH, contest data, propagation info, etc.)
    additional_data JSONB                  DEFAULT '{}'::jsonb,

    CONSTRAINT qso_data_no_duplicates CHECK (
        additional_data ? 'call' = false AND
        additional_data ? 'band' = false AND
        additional_data ? 'mode' = false AND
        additional_data ? 'freq' = false AND
        additional_data ? 'qso_date' = false AND
        additional_data ? 'time_on' = false AND
        additional_data ? 'time_off' = false AND
        additional_data ? 'rst_send' = false AND
        additional_data ? 'rst_recv' = false AND
        additional_data ? 'country' = false
        )
);

CREATE INDEX idx_qso_call ON qso (call);
CREATE INDEX idx_qso_band ON qso (band);
CREATE INDEX idx_qso_country ON qso (country);
CREATE INDEX idx_qso_date_time ON qso (qso_date, time_on);
CREATE INDEX idx_qso_additional_gin ON qso USING gin (additional_data);

