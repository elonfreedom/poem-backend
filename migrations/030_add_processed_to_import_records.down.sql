-- 回滚：删除 processed 字段
ALTER TABLE import_records DROP COLUMN IF EXISTS processed;
