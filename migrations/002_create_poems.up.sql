-- 分类表
CREATE TABLE categories (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50) NOT NULL UNIQUE,
    sort        INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 标签表
CREATE TABLE tags (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 诗歌表
CREATE TABLE poems (
    id              BIGSERIAL PRIMARY KEY,
    title           VARCHAR(100) NOT NULL,
    author          VARCHAR(50) NOT NULL,
    dynasty         VARCHAR(20),
    content         TEXT NOT NULL,
    translation     TEXT,
    appreciation    TEXT,
    category_id     BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    tags            TEXT[] DEFAULT '{}',
    cover_url       VARCHAR(500),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft, published, archived
    created_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_poems_category_id ON poems(category_id);
CREATE INDEX idx_poems_status ON poems(status);
CREATE INDEX idx_poems_created_at ON poems(created_at DESC);

-- 诗歌标签关联表
CREATE TABLE poem_tags (
    poem_id     BIGINT NOT NULL REFERENCES poems(id) ON DELETE CASCADE,
    tag_id      BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (poem_id, tag_id)
);

CREATE INDEX idx_poem_tags_tag_id ON poem_tags(tag_id);
