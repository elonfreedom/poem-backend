-- 添加密码字段到用户表（用于后台管理员登录）
-- 使用 IF NOT EXISTS 确保幂等，重复执行不报错
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_users_email_password ON users(email) WHERE email IS NOT NULL AND password_hash IS NOT NULL;
