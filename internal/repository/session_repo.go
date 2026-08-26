package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionRepository WebAuthn 会话持久化
type SessionRepository struct {
	db *pgxpool.Pool
}

// NewSessionRepository 创建会话仓库
func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

// Store 存储 WebAuthn 会话
func (r *SessionRepository) Store(ctx context.Context, id string, session webauthn.SessionData, userID *string, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO webauthn_sessions (id, session_data, user_id, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			session_data = EXCLUDED.session_data,
			user_id = EXCLUDED.user_id,
			expires_at = EXCLUDED.expires_at
	`
	_, err = r.db.Exec(ctx, query, id, data, userID, time.Now().Add(ttl))
	return err
}

// Get 获取 WebAuthn 会话
func (r *SessionRepository) Get(ctx context.Context, id string) (*webauthn.SessionData, *string, bool, error) {
	query := `SELECT session_data, user_id FROM webauthn_sessions WHERE id = $1 AND expires_at > NOW()`
	row := r.db.QueryRow(ctx, query, id)
	var data []byte
	var userID *string
	err := row.Scan(&data, &userID)
	if err != nil {
		return nil, nil, false, err
	}
	var session webauthn.SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, nil, false, err
	}
	return &session, userID, true, nil
}

// Delete 删除 WebAuthn 会话
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM webauthn_sessions WHERE id = $1`, id)
	return err
}

// CleanupExpired 清理过期会话
func (r *SessionRepository) CleanupExpired(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM webauthn_sessions WHERE expires_at <= NOW()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
