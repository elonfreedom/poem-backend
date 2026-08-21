-- 用户表
CREATE TABLE users (
    id          UUID PRIMARY KEY,                          -- UUID v7，由应用层生成
    nickname    VARCHAR(50) NOT NULL DEFAULT '诗友',        -- 昵称，默认自动生成
    email       VARCHAR(255) UNIQUE,                       -- 邮箱（可选，用于恢复）
    role        VARCHAR(20) NOT NULL DEFAULT 'user',       -- admin, user
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_role ON users(role);
