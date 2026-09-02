-- ============================================
-- import_records 表添加 updated_at 字段
-- ============================================

ALTER TABLE import_records
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
