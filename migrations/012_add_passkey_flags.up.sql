-- 添加 passkey flags 字段（用于存储 WebAuthn 凭证标志位）
ALTER TABLE passkeys ADD COLUMN flags SMALLINT NOT NULL DEFAULT 0;
