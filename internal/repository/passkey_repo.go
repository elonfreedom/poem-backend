package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

usermodel "poem-backend/internal/model/user"
)

type PasskeyRepository struct {
	db *pgxpool.Pool
}

func NewPasskeyRepository(db *pgxpool.Pool) *PasskeyRepository {
	return &PasskeyRepository{db: db}
}

// Create 创建 Passkey
func (r *PasskeyRepository) Create(ctx context.Context, passkey *usermodel.Passkey) error {
	query := `
		INSERT INTO passkeys (user_id, credential_id, public_key, sign_count, device_name, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	err := r.db.QueryRow(ctx, query,
		passkey.UserID, passkey.CredentialID, passkey.PublicKey,
		passkey.SignCount, passkey.DeviceName, passkey.CreatedAt).Scan(&passkey.ID)
	return err
}

// GetByCredentialID 根据凭证 ID 获取 Passkey
func (r *PasskeyRepository) GetByCredentialID(ctx context.Context, credentialID []byte) (*usermodel.Passkey, error) {
	query := `
		SELECT id, user_id, credential_id, public_key, sign_count, device_name, created_at, last_used_at
		FROM passkeys WHERE credential_id = $1
	`
	row := r.db.QueryRow(ctx, query, credentialID)
	var p usermodel.Passkey
	var lastUsedAt *time.Time
	err := row.Scan(&p.ID, &p.UserID, &p.CredentialID, &p.PublicKey,
		&p.SignCount, &p.DeviceName, &p.CreatedAt, lastUsedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetByUserID 获取用户的所有 Passkey
func (r *PasskeyRepository) GetByUserID(ctx context.Context, userID string) ([]usermodel.Passkey, error) {
	query := `
		SELECT id, user_id, credential_id, public_key, sign_count, device_name, created_at, last_used_at
		FROM passkeys WHERE user_id = $1 ORDER BY created_at
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passkeys []usermodel.Passkey
	for rows.Next() {
		var p usermodel.Passkey
		err := rows.Scan(&p.ID, &p.UserID, &p.CredentialID, &p.PublicKey,
			&p.SignCount, &p.DeviceName, &p.CreatedAt, &p.LastUsedAt)
		if err != nil {
			return nil, err
		}
		passkeys = append(passkeys, p)
	}
	return passkeys, rows.Err()
}

// UpdateSignCount 更新签名计数器和最后使用时间
func (r *PasskeyRepository) UpdateSignCount(ctx context.Context, id int64, signCount uint32) error {
	query := `UPDATE passkeys SET sign_count = $1, last_used_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, signCount, id)
	return err
}

// Delete 删除 Passkey
func (r *PasskeyRepository) Delete(ctx context.Context, id int64, userID string) error {
	query := `DELETE FROM passkeys WHERE id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, id, userID)
	return err
}

// CountByUserID 统计用户 Passkey 数量
func (r *PasskeyRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM passkeys WHERE user_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}
