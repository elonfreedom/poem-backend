-- 回退订阅计划排队机制

-- 删除新增索引
DROP INDEX IF EXISTS idx_one_current_per_user;
DROP INDEX IF EXISTS idx_unique_user_plan_subscribed;
DROP INDEX IF EXISTS idx_subs_user_queue;

-- 恢复唯一约束
ALTER TABLE plan_subscriptions
    ADD CONSTRAINT plan_subscriptions_user_id_shared_plan_id_key
    UNIQUE (user_id, shared_plan_id);

-- 恢复 start_date 非空
ALTER TABLE plan_subscriptions ALTER COLUMN start_date SET NOT NULL;

-- 删除新增字段
ALTER TABLE plan_subscriptions DROP COLUMN IF EXISTS is_current;
ALTER TABLE plan_subscriptions DROP COLUMN IF EXISTS queue_order;
