-- 添加状态字段到用户表（用于禁用/启用前端用户）
-- 使用 IF NOT EXISTS 确保幂等，重复执行不报错
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';

CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
