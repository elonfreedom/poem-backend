-- ============================================
-- 阅读计划共享库
-- ============================================

-- 共享计划表（社区发布的计划模板）
CREATE TABLE IF NOT EXISTS shared_plans (
    id              BIGSERIAL PRIMARY KEY,
    creator_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           VARCHAR(100) NOT NULL,
    description     TEXT DEFAULT '',
    tags            TEXT[] DEFAULT '{}',
    poem_ids        BIGINT[] NOT NULL,
    daily_count     INT NOT NULL DEFAULT 1 CHECK (daily_count > 0),
    total_days      INT NOT NULL CHECK (total_days > 0),
    subscribe_count INT NOT NULL DEFAULT 0,
    is_published    BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shared_plans_published ON shared_plans(is_published, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shared_plans_subscribe ON shared_plans(is_published, subscribe_count DESC);
CREATE INDEX IF NOT EXISTS idx_shared_plans_tags ON shared_plans USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_shared_plans_creator ON shared_plans(creator_id);

-- 用户订阅关系表
CREATE TABLE IF NOT EXISTS plan_subscriptions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_plan_id  BIGINT NOT NULL REFERENCES shared_plans(id) ON DELETE CASCADE,
    start_date      DATE NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'active', -- active, completed, paused
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, shared_plan_id)
);

CREATE INDEX IF NOT EXISTS idx_subs_user ON plan_subscriptions(user_id, status);
CREATE INDEX IF NOT EXISTS idx_subs_plan ON plan_subscriptions(shared_plan_id);

-- 打卡记录表（订阅计划的每日打卡）
CREATE TABLE IF NOT EXISTS plan_checkins (
    id              BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES plan_subscriptions(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_number      INT NOT NULL,                            -- 第几天
    checkin_date    DATE NOT NULL,                          -- 打卡日期
    poem_ids        BIGINT[] NOT NULL,                      -- 当天打卡的诗歌
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(subscription_id, day_number)
);

CREATE INDEX IF NOT EXISTS idx_checkins_sub ON plan_checkins(subscription_id, day_number);
CREATE INDEX IF NOT EXISTS idx_checkins_user ON plan_checkins(user_id, checkin_date);

-- 现有 reading_plans 表改造：支持订阅来源
ALTER TABLE reading_plans ADD COLUMN IF NOT EXISTS source VARCHAR(20) NOT NULL DEFAULT 'custom';
ALTER TABLE reading_plans ADD COLUMN IF NOT EXISTS shared_plan_id BIGINT REFERENCES shared_plans(id) ON DELETE SET NULL;
ALTER TABLE reading_plans ADD COLUMN IF NOT EXISTS poem_ids BIGINT[] DEFAULT '{}';
