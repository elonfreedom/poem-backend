-- 添加状态字段到用户表（用于禁用/启用前端用户）
ALTER TABLE users ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';

CREATE INDEX idx_users_status ON users(status);
