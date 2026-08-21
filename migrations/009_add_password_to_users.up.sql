-- 添加密码字段到用户表（用于后台管理员登录）
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255);

CREATE INDEX idx_users_email_password ON users(email) WHERE email IS NOT NULL AND password_hash IS NOT NULL;
