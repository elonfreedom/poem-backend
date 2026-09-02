-- 回滚：删除 import_records 表
DROP INDEX IF EXISTS idx_import_records_created_by;
DROP INDEX IF EXISTS idx_import_records_status;
DROP INDEX IF EXISTS idx_import_records_created_at;
DROP TABLE IF EXISTS import_records;
