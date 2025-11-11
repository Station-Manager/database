PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS api_keys
(
    id            INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,
    key_name      TEXT     NOT NULL UNIQUE,
    key_hash      TEXT     NOT NULL,
    key_prefix    TEXT     NOT NULL CHECK (length(key_prefix) <= 10),

    scopes        TEXT,
    allowed_ips   TEXT,

    created_at    DATETIME NOT NULL DEFAULT (datetime('now', 'localtime')),
    last_used_at  DATETIME,
    expires_at    DATETIME,
    revoked_at    DATETIME,

    created_by    TEXT,
    revoked_by    TEXT,
    use_count     INTEGER  NOT NULL DEFAULT (0)
);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_prefix ON api_keys (key_prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys (revoked_at);
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys (expires_at);
