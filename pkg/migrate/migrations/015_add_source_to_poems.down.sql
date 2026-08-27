-- 幂等删除，重复执行不报错
DROP INDEX IF EXISTS idx_poems_source;
ALTER TABLE poems DROP COLUMN IF EXISTS source;
