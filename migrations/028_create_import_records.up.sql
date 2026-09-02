-- ============================================
-- 导入记录表：记录每次批量导入的元数据
-- ============================================

CREATE TABLE IF NOT EXISTS import_records (
    id              BIGSERIAL PRIMARY KEY,
    file_name       VARCHAR(500) NOT NULL DEFAULT '',     -- 上传文件名（可为空）
    source          VARCHAR(200) NOT NULL DEFAULT '',     -- 来源标注
    total           INT NOT NULL DEFAULT 0,               -- 总记录数
    success         INT NOT NULL DEFAULT 0,               -- 成功数
    failed          INT NOT NULL DEFAULT 0,               -- 失败数
    status          VARCHAR(20) NOT NULL DEFAULT 'success', -- success/partial/failed
    errors          JSONB NOT NULL DEFAULT '[]',          -- 错误详情数组
    created_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_import_records_created_at ON import_records(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_import_records_status ON import_records(status);
CREATE INDEX IF NOT EXISTS idx_import_records_created_by ON import_records(created_by);
