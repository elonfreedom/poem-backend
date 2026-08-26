-- WebAuthn 会话表（持久化，替代内存存储）
CREATE TABLE webauthn_sessions (
    id          TEXT PRIMARY KEY,                   -- 会话 ID（UUID）
    session_data BYTEA NOT NULL,                    -- 序列化的 webauthn.SessionData
    user_id     TEXT,                               -- 用户 ID（仅注册流程使用，登录流程为 NULL）
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL                -- 过期时间（创建后 10 分钟）
);

CREATE INDEX idx_webauthn_sessions_expires_at ON webauthn_sessions(expires_at);

-- 跨设备连接表（持久化，替代内存存储）
CREATE TABLE connection_sessions (
    token           TEXT PRIMARY KEY,               -- 连接令牌（UUID）
    user_id         TEXT NOT NULL,                  -- 设备 A 的用户 ID
    device_name     VARCHAR(100),                   -- 设备 B 的设备名称
    status          TEXT NOT NULL DEFAULT 'waiting',-- 当前状态
    session_data    BYTEA NOT NULL,                 -- 序列化的 webauthn.SessionData
    options_data    BYTEA NOT NULL,                 -- 序列化的 CredentialCreation options
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL            -- 过期时间
);

CREATE INDEX idx_connection_sessions_expires_at ON connection_sessions(expires_at);
