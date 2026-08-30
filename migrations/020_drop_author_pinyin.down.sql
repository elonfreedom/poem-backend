-- 恢复 author_pinyin 字段（如需回滚）
ALTER TABLE poems ADD COLUMN IF NOT EXISTS author_pinyin TEXT DEFAULT '';
