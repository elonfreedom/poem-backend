-- ============================================
-- import_records 表添加 processed 字段（实时进度）
-- ============================================

ALTER TABLE import_records
    ADD COLUMN IF NOT EXISTS processed INTEGER NOT NULL DEFAULT 0;
