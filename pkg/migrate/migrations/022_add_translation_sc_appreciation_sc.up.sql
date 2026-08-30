-- 为 poems 表添加 translation_sc 和 appreciation_sc 列（简繁体转换扩展）
ALTER TABLE poems ADD COLUMN IF NOT EXISTS translation_sc TEXT DEFAULT '';
ALTER TABLE poems ADD COLUMN IF NOT EXISTS appreciation_sc TEXT DEFAULT '';
