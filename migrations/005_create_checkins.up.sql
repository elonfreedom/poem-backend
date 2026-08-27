-- 打卡记录表（复合主键：user_id + date）
CREATE TABLE IF NOT EXISTS checkins (
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date                DATE NOT NULL,
    consecutive_day     INT NOT NULL DEFAULT 1,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_checkins_date ON checkins(date);

-- 打卡统计表（主键：user_id）
CREATE TABLE IF NOT EXISTS checkin_stats (
    user_id             UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_days          INT NOT NULL DEFAULT 0,
    consecutive_day     INT NOT NULL DEFAULT 0,
    max_consecutive     INT NOT NULL DEFAULT 0,
    last_check_in       DATE
);
