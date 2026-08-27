-- 诗歌拼音字段（支持 admin 手动校正多音字）
ALTER TABLE poems ADD COLUMN IF NOT EXISTS title_pinyin TEXT DEFAULT '';
ALTER TABLE poems ADD COLUMN IF NOT EXISTS content_pinyin TEXT DEFAULT '';
ALTER TABLE poems ADD COLUMN IF NOT EXISTS author_pinyin TEXT DEFAULT '';
