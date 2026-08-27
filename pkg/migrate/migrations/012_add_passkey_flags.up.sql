-- 添加 passkey flags 字段（用于存储 WebAuthn 凭证标志位）
-- 使用 IF NOT EXISTS 确保幂等，重复执行不报错
ALTER TABLE passkeys ADD COLUMN IF NOT EXISTS flags SMALLINT NOT NULL DEFAULT 0;
