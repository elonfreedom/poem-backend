-- 回滚：恢复 poems.title 列长度为 VARCHAR(100)
ALTER TABLE poems ALTER COLUMN title TYPE VARCHAR(100);
