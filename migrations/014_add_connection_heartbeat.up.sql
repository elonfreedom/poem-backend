-- 添加心跳超时机制：检测设备 B 是否离线（关闭浏览器/断网）
-- 使用 IF NOT EXISTS 确保幂等，重复执行不报错
ALTER TABLE connection_sessions ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_connection_sessions_last_active_at ON connection_sessions(last_active_at);
