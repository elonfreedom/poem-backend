-- 回滚：删除 updated_at 字段
ALTER TABLE import_records DROP COLUMN IF EXISTS updated_at;
