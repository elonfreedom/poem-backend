-- 阅读计划表（复合主键：user_id + plan_id，plan_id 为用户级自增）
CREATE TABLE IF NOT EXISTS reading_plans (
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id         INT NOT NULL,                           -- 用户级自增
    daily_count     INT NOT NULL CHECK (daily_count > 0),
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, completed, paused
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, plan_id)
);

CREATE INDEX IF NOT EXISTS idx_reading_plans_user_status ON reading_plans(user_id, status);

-- 阅读进度表（复合主键：user_id + date）
CREATE TABLE IF NOT EXISTS reading_progress (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date        DATE NOT NULL,
    read_count  INT NOT NULL DEFAULT 0,
    poem_ids    BIGINT[] DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_reading_progress_date ON reading_progress(date);
