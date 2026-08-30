-- ============================================
-- 回滚：重建 plan_checkins 表
-- ============================================

-- 1. 重建 plan_checkins 表
CREATE TABLE IF NOT EXISTS plan_checkins (
    id              BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES plan_subscriptions(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day_number      INT NOT NULL,
    checkin_date    DATE NOT NULL,
    poem_ids        BIGINT[] NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(subscription_id, day_number)
);

-- 2. 将订阅打卡数据迁回 plan_checkins
INSERT INTO plan_checkins (subscription_id, user_id, day_number, checkin_date, poem_ids, created_at)
SELECT subscription_id, user_id, day_number, date, poem_ids, created_at
FROM checkins WHERE subscription_id IS NOT NULL;

-- 3. 删除迁移添加的索引
DROP INDEX IF EXISTS uk_checkins_sub_day;
DROP INDEX IF EXISTS uk_checkins_user_date_no_sub;
DROP INDEX IF EXISTS idx_checkins_user_sub_date;
DROP INDEX IF EXISTS idx_checkins_subscription_id;

-- 4. 删除迁移添加的列
ALTER TABLE checkins DROP COLUMN IF EXISTS poem_ids;
ALTER TABLE checkins DROP COLUMN IF EXISTS day_number;
ALTER TABLE checkins DROP COLUMN IF EXISTS subscription_id;
ALTER TABLE checkins DROP COLUMN IF EXISTS id;

-- 5. 恢复旧主键
ALTER TABLE checkins ADD PRIMARY KEY (user_id, date);
