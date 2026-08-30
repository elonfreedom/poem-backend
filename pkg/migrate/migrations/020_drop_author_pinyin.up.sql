-- 删除未使用的 author_pinyin 字段（作者不参与拼音标注）
ALTER TABLE poems DROP COLUMN IF EXISTS author_pinyin;
