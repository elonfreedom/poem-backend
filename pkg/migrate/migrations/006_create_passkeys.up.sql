-- Passkey 通行密钥表
CREATE TABLE IF NOT EXISTS passkeys (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id   BYTEA NOT NULL UNIQUE,                  -- 凭证 ID（唯一）
    public_key      BYTEA NOT NULL,                         -- 公钥
    sign_count      INT NOT NULL DEFAULT 0,                 -- 签名计数器（防重放）
    device_name     VARCHAR(100),                           -- 设备名称
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_passkeys_user_id ON passkeys(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_passkeys_credential_id ON passkeys(credential_id);
