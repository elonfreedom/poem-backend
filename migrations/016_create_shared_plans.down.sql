-- 回退共享计划表
ALTER TABLE reading_plans DROP COLUMN IF EXISTS source;
ALTER TABLE reading_plans DROP COLUMN IF EXISTS shared_plan_id;
ALTER TABLE reading_plans DROP COLUMN IF EXISTS poem_ids;

DROP TABLE IF EXISTS plan_checkins;
DROP TABLE IF EXISTS plan_subscriptions;
DROP TABLE IF EXISTS shared_plans;
