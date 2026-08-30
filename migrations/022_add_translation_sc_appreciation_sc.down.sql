-- 回滚：删除 translation_sc 和 appreciation_sc 列
ALTER TABLE poems DROP COLUMN IF EXISTS translation_sc;
ALTER TABLE poems DROP COLUMN IF EXISTS appreciation_sc;
