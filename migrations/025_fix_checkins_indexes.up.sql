-- ============================================
-- 修复打卡约束：保留部分唯一索引
-- 原因：ON CONFLICT ON CONSTRAINT 不能用于索引
-- 解决：使用 ON CONFLICT (column_list) WHERE predicate 匹配部分索引
-- 注意：不能去掉 WHERE 子句！用户同一天可能既有用户级打卡又有订阅打卡
-- ============================================

-- 无需修改索引，保持原有部分唯一索引不变：
-- uk_checkins_sub_day: (subscription_id, day_number) WHERE subscription_id IS NOT NULL
-- uk_checkins_user_date_no_sub: (user_id, date) WHERE subscription_id IS NULL

-- 此迁移文件作为占位，实际修复在代码层面（checkin_repo.go）
