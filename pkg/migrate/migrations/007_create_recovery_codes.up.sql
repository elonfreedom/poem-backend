-- 恢复码表（用于账号恢复）
CREATE TABLE IF NOT EXISTS recovery_codes (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash   VARCHAR(255) NOT NULL,                      -- bcrypt 哈希
    used        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_recovery_codes_user_id ON recovery_codes(user_id);
