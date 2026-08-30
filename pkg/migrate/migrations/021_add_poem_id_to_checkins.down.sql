-- 回滚：删除 poem_id 列
ALTER TABLE checkins DROP COLUMN IF EXISTS poem_id;
