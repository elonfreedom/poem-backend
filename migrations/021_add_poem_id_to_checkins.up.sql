-- 为 checkins 表添加 poem_id 列（关联诗歌，用于热力图 hover 显示）
ALTER TABLE checkins ADD COLUMN IF NOT EXISTS poem_id BIGINT REFERENCES poems(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_checkins_poem_id ON checkins(poem_id);
