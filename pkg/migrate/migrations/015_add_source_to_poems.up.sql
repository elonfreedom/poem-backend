-- 幂等迁移，重复执行不报错
ALTER TABLE poems ADD COLUMN IF NOT EXISTS source VARCHAR(200) NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_poems_source ON poems(source);
