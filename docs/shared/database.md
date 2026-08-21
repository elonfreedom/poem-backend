# 数据库设计

## 表结构

### 用户相关

| 表名 | 说明 | 主要字段 |
|-----|------|---------|
| users | 用户表 | id, phone, nickname, avatar_url, role, wechat_openid, apple_user_id, created_at, updated_at |

### 诗歌相关

| 表名 | 说明 | 主要字段 |
|-----|------|---------|
| poems | 诗歌表 | id, title, author, dynasty, content, translation, appreciation, category_id, cover_url, status, created_by, created_at, updated_at |
| categories | 分类表 | id, name, sort, created_at, updated_at |
| tags | 标签表 | id, name, created_at |
| poem_tags | 诗歌标签关联 | poem_id, tag_id |

### 收藏相关

| 表名 | 说明 | 主要字段 |
|-----|------|---------|
| favorites | 收藏表 | id, user_id, poem_id, created_at |

### 阅读计划相关

| 表名 | 说明 | 主要字段 |
|-----|------|---------|
| reading_plans | 阅读计划 | id, user_id, daily_count, start_date, end_date, status, created_at, updated_at |
| reading_progress | 阅读进度 | id, plan_id, user_id, date, read_count, poem_ids, created_at |

### 打卡相关

| 表名 | 说明 | 主要字段 |
|-----|------|---------|
| checkins | 打卡记录 | id, user_id, date, consecutive_day, created_at |
| checkin_stats | 打卡统计 | user_id, total_days, consecutive_day, max_consecutive, last_check_in |

### 系统相关

| 表名 | 说明 | 主要字段 |
|-----|------|---------|
| banners | 轮播图 | id, title, image_url, link_type, link_value, sort, status, created_at, updated_at |
| announcements | 公告 | id, title, content, status, created_at, updated_at |
| system_configs | 系统配置 | id, key, value, remark, updated_at |
| poem_views | 浏览记录 | id, poem_id, user_id, created_at |

## 实体关系

```
users (1) ─── (N) favorites ─── (1) poems
users (1) ─── (N) reading_plans ─── (N) reading_progress
users (1) ─── (N) checkins
users (1) ─── (1) checkin_stats
poems (N) ─── (1) categories
poems (N) ─── (N) tags (通过 poem_tags)
poems (1) ─── (N) poem_views
```

## 索引设计

```sql
-- 用户表
CREATE UNIQUE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_wechat_openid ON users(wechat_openid);
CREATE INDEX idx_users_role ON users(role);

-- 诗歌表
CREATE INDEX idx_poems_category_id ON poems(category_id);
CREATE INDEX idx_poems_status ON poems(status);
CREATE INDEX idx_poems_created_at ON poems(created_at);
CREATE INDEX idx_poems_title_gin ON poems USING gin(to_tsvector('chinese', title));
CREATE INDEX idx_poems_content_gin ON poems USING gin(to_tsvector('chinese', content));

-- 收藏表
CREATE UNIQUE INDEX idx_favorites_user_poem ON favorites(user_id, poem_id);
CREATE INDEX idx_favorites_user_id ON favorites(user_id);

-- 打卡表
CREATE UNIQUE INDEX idx_checkins_user_date ON checkins(user_id, date);
CREATE INDEX idx_checkins_date ON checkins(date);

-- 阅读进度
CREATE INDEX idx_reading_progress_plan_id ON reading_progress(plan_id);
CREATE INDEX idx_reading_progress_user_date ON reading_progress(user_id, date);

-- 浏览记录
CREATE INDEX idx_poem_views_poem_id ON poem_views(poem_id);
CREATE INDEX idx_poem_views_created_at ON poem_views(created_at);
```

## 数据字典

### 状态枚举

| 字段 | 枚举值 | 说明 |
|-----|--------|------|
| poems.status | draft / published / archived | 草稿 / 已发布 / 已归档 |
| reading_plans.status | active / completed / paused | 进行中 / 已完成 / 已暂停 |
| banners.status | active / inactive | 启用 / 禁用 |
| announcements.status | draft / published | 草稿 / 已发布 |
| users.role | admin / user | 管理员 / 普通用户 |

### 通用字段

| 字段 | 类型 | 说明 |
|-----|------|------|
| id | bigint | 主键，自增 |
| created_at | timestamp | 创建时间 |
| updated_at | timestamp | 更新时间 |
