package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connection 跨设备连接数据
type Connection struct {
	Token        string
	UserID       string
	DeviceName   string
	Status       string
	Session      any // webauthn.SessionData 序列化前
	Options      any // CredentialCreation 序列化前
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastActiveAt time.Time // 设备 B 最后活跃时间（心跳超时检测）
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
		INSERT INTO connection_sessions (token, user_id, device_name, status, session_data, options_data, expires_at, last_active_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (token) DO UPDATE SET
			status = EXCLUDED.status,
			device_name = EXCLUDED.device_name,
			last_active_at = EXCLUDED.last_active_at
	`
	_, err = r.db.Exec(ctx, query,
		conn.Token, conn.UserID, conn.DeviceName, conn.Status,
		sessionData, optionsData, conn.ExpiresAt, conn.LastActiveAt)
	return err
}

// Get 获取连接
func (r *ConnectionRepository) Get(ctx context.Context, token string) (*Connection, bool, error) {
	query := `SELECT token, user_id, device_name, status, session_data, options_data, expires_at, last_active_at FROM connection_sessions WHERE token = $1 AND expires_at > NOW()`
	row := r.db.QueryRow(ctx, query, token)
	var conn Connection
	var sessionData, optionsData []byte
	err := row.Scan(&conn.Token, &conn.UserID, &conn.DeviceName, &conn.Status, &sessionData, &optionsData, &conn.ExpiresAt, &conn.LastActiveAt)
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

// UpdateHeartbeat 更新设备 B 心跳时间
func (r *ConnectionRepository) UpdateHeartbeat(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx, `UPDATE connection_sessions SET last_active_at = NOW() WHERE token = $1`, token)
	return err
}

// CleanupInactive 清理超时不活跃的连接（设备 B 断网/关闭浏览器）
// timeout: 超过此时间未活跃则标记为 expired
// 只处理 connected 和 confirmed 状态的连接
// waiting 状态不参与心跳超时检测（设备 B 还没扫码，只使用 QR 码过期时间）
func (r *ConnectionRepository) CleanupInactive(ctx context.Context, timeout time.Duration) (int64, error) {
	cutoff := time.Now().Add(-timeout)
	tag, err := r.db.Exec(ctx, `
		UPDATE connection_sessions
		SET status = 'expired'
		WHERE status IN ('connected', 'confirmed')
		  AND last_active_at < $1
	`, cutoff)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
