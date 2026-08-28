-- 诗歌简体字段（标题、作者、正文的繁体原文自动生成简体，供前端切换阅读）
ALTER TABLE poems ADD COLUMN IF NOT EXISTS title_sc TEXT DEFAULT '';
ALTER TABLE poems ADD COLUMN IF NOT EXISTS author_sc TEXT DEFAULT '';
ALTER TABLE poems ADD COLUMN IF NOT EXISTS content_sc TEXT DEFAULT '';
