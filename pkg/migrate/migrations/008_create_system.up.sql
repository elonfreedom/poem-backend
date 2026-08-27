-- Banner 轮播图表
CREATE TABLE IF NOT EXISTS banners (
    id          BIGSERIAL PRIMARY KEY,
    title       VARCHAR(100) NOT NULL,
    image_url   VARCHAR(500) NOT NULL,
    link_type   VARCHAR(20) NOT NULL DEFAULT 'url',        -- poem, url
    link_value  VARCHAR(500) NOT NULL,
    sort        INT NOT NULL DEFAULT 0,
    status      VARCHAR(20) NOT NULL DEFAULT 'active',      -- active, inactive
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_banners_status_sort ON banners(status, sort);

-- 公告表
CREATE TABLE IF NOT EXISTS announcements (
    id          BIGSERIAL PRIMARY KEY,
    title       VARCHAR(200) NOT NULL,
    content     TEXT NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'draft',       -- draft, published
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_announcements_status ON announcements(status);

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id          BIGSERIAL PRIMARY KEY,
    key         VARCHAR(100) NOT NULL UNIQUE,
    value       TEXT NOT NULL,
    remark      VARCHAR(255),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 诗歌浏览记录表
CREATE TABLE IF NOT EXISTS poem_views (
    id          BIGSERIAL PRIMARY KEY,
    poem_id     BIGINT NOT NULL REFERENCES poems(id) ON DELETE CASCADE,
    user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_poem_views_poem_id ON poem_views(poem_id);
CREATE INDEX IF NOT EXISTS idx_poem_views_created_at ON poem_views(created_at);
