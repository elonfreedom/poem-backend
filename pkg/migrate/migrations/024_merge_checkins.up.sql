-- ============================================
-- 合并打卡系统：将 plan_checkins 数据迁移到 checkins
-- ============================================

-- 1. 添加新列到 checkins 表
ALTER TABLE checkins ADD COLUMN IF NOT EXISTS id BIGSERIAL;
ALTER TABLE checkins ADD COLUMN IF NOT EXISTS subscription_id BIGINT REFERENCES plan_subscriptions(id) ON DELETE CASCADE;
ALTER TABLE checkins ADD COLUMN IF NOT EXISTS day_number INT;
ALTER TABLE checkins ADD COLUMN IF NOT EXISTS poem_ids BIGINT[] DEFAULT '{}';

-- 2. 添加辅助索引
CREATE INDEX IF NOT EXISTS idx_checkins_subscription_id ON checkins(subscription_id) WHERE subscription_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_checkins_user_sub_date ON checkins(user_id, subscription_id, date);

-- 3. 迁移 plan_checkins 数据到 checkins
INSERT INTO checkins (user_id, date, consecutive_day, poem_id, created_at, subscription_id, day_number, poem_ids)
SELECT
    pc.user_id,
    pc.checkin_date AS date,
    1 AS consecutive_day,
    CASE WHEN array_length(pc.poem_ids, 1) > 0 THEN pc.poem_ids[1] ELSE NULL END AS poem_id,
    pc.created_at,
    pc.subscription_id,
    pc.day_number,
    pc.poem_ids
FROM plan_checkins pc
ON CONFLICT DO NOTHING;

-- 4. 删除旧的 (user_id, date) 主键约束，改为使用 id 作为主键
ALTER TABLE checkins DROP CONSTRAINT IF EXISTS checkins_pkey;
ALTER TABLE checkins ADD PRIMARY KEY (id);

-- 5. 添加部分唯一约束（保持旧系统的一天一卡 + 订阅系统的一天一卡）
CREATE UNIQUE INDEX IF NOT EXISTS uk_checkins_user_date_no_sub ON checkins (user_id, date) WHERE subscription_id IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uk_checkins_sub_day ON checkins (subscription_id, day_number) WHERE subscription_id IS NOT NULL;

-- 6. 删除 plan_checkins 表
DROP TABLE IF EXISTS plan_checkins;
