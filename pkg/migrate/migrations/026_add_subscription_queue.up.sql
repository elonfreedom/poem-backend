-- ============================================
-- 订阅计划排队机制
-- ============================================

-- 新增字段
ALTER TABLE plan_subscriptions
    ADD COLUMN IF NOT EXISTS is_current BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS queue_order INT NOT NULL DEFAULT 0;

-- 存量数据迁移：active → subscribed + is_current=true
UPDATE plan_subscriptions SET status = 'subscribed', is_current = true WHERE status = 'active';

-- 存量数据迁移：paused → cancelled
UPDATE plan_subscriptions SET status = 'cancelled' WHERE status = 'paused';

-- completed 保持不变

-- start_date 改为可空（仅当前计划有值）
ALTER TABLE plan_subscriptions ALTER COLUMN start_date DROP NOT NULL;

-- 替换唯一约束：原为 UNIQUE(user_id, shared_plan_id)
ALTER TABLE plan_subscriptions
    DROP CONSTRAINT IF EXISTS plan_subscriptions_user_id_shared_plan_id_key;

-- 每用户最多一个当前计划（防并发）
CREATE UNIQUE INDEX idx_one_current_per_user
    ON plan_subscriptions(user_id) WHERE is_current = true;

-- 每用户每计划最多一个 subscribed（允许取消后重订）
CREATE UNIQUE INDEX idx_unique_user_plan_subscribed
    ON plan_subscriptions(user_id, shared_plan_id) WHERE status = 'subscribed';

-- 排队订阅索引（按 queue_order 排序查询）
CREATE INDEX IF NOT EXISTS idx_subs_user_queue
    ON plan_subscriptions(user_id, queue_order) WHERE status = 'subscribed';

-- 更新状态注释
COMMENT ON COLUMN plan_subscriptions.status IS '状态: subscribed, completed, cancelled';
