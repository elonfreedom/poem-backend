package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connection 跨设备连接数据
type Connection struct {
	Token      string
	UserID     string
	DeviceName string
	Status     string
	Session    any // webauthn.SessionData 序列化前
	Options    any // CredentialCreation 序列化前
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

// ConnectionRepository 跨设备连接持久化
type ConnectionRepository struct {
	db *pgxpool.Pool
}

// NewConnectionRepository 创建连接仓库
func NewConnectionRepository(db *pgxpool.Pool) *ConnectionRepository {
	return &ConnectionRepository{db: db}
}

// Store 存储连接
func (r *ConnectionRepository) Store(ctx context.Context, conn *Connection) error {
	sessionData, err := json.Marshal(conn.Session)
	if err != nil {
		return err
	}
	optionsData, err := json.Marshal(conn.Options)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO connection_sessions (token, user_id, device_name, status, session_data, options_data, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (token) DO UPDATE SET
			status = EXCLUDED.status,
			device_name = EXCLUDED.device_name
	`
	_, err = r.db.Exec(ctx, query,
		conn.Token, conn.UserID, conn.DeviceName, conn.Status,
		sessionData, optionsData, conn.ExpiresAt)
	return err
}

// Get 获取连接
func (r *ConnectionRepository) Get(ctx context.Context, token string) (*Connection, bool, error) {
	query := `SELECT token, user_id, device_name, status, session_data, options_data, expires_at FROM connection_sessions WHERE token = $1 AND expires_at > NOW()`
	row := r.db.QueryRow(ctx, query, token)
	var conn Connection
	var sessionData, optionsData []byte
	err := row.Scan(&conn.Token, &conn.UserID, &conn.DeviceName, &conn.Status, &sessionData, &optionsData, &conn.ExpiresAt)
	if err != nil {
		return nil, false, err
	}
	if err := json.Unmarshal(sessionData, &conn.Session); err != nil {
		return nil, false, err
	}
	if err := json.Unmarshal(optionsData, &conn.Options); err != nil {
		return nil, false, err
	}
	return &conn, true, nil
}

// Update 更新连接状态
func (r *ConnectionRepository) Update(ctx context.Context, token string, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE connection_sessions SET status = $1 WHERE token = $2`, status, token)
	return err
}

// Delete 删除连接
func (r *ConnectionRepository) Delete(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM connection_sessions WHERE token = $1`, token)
	return err
}

// CleanupExpired 清理过期连接
func (r *ConnectionRepository) CleanupExpired(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM connection_sessions WHERE expires_at <= NOW()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
