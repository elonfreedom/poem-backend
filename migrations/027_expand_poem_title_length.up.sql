-- 扩大 poems.title 列长度（VARCHAR(100) → VARCHAR(200)）
-- 原因：部分诗歌标题超过 100 字符，导致导入失败
ALTER TABLE poems ALTER COLUMN title TYPE VARCHAR(200);
