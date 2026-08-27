-- 幂等删除，重复执行不报错
DROP INDEX IF EXISTS idx_connection_sessions_last_active_at;
ALTER TABLE connection_sessions DROP COLUMN IF EXISTS last_active_at;
