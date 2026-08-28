-- 作者表
CREATE TABLE IF NOT EXISTS authors (
    id                BIGSERIAL PRIMARY KEY,
    name              VARCHAR(100) NOT NULL,
    name_traditional  VARCHAR(100) NOT NULL DEFAULT '',
    dynasty           VARCHAR(20) NOT NULL DEFAULT '未知',
    biography         TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_authors_name ON authors(name);
CREATE INDEX IF NOT EXISTS idx_authors_dynasty ON authors(dynasty);

-- 为 poems 表添加 author_id 外键
ALTER TABLE poems ADD COLUMN IF NOT EXISTS author_id BIGINT REFERENCES authors(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_poems_author_id ON poems(author_id);
