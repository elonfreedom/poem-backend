-- 收藏表（复合主键：user_id + poem_id）
CREATE TABLE IF NOT EXISTS favorites (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    poem_id     BIGINT NOT NULL REFERENCES poems(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, poem_id)
);

CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_poem_id ON favorites(poem_id);
