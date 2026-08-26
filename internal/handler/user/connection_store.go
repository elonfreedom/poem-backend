package user

import (
	"context"
	"time"

	"poem-backend/internal/repository"
)

// connectionTTL 连接有效期
const connectionTTL = 10 * time.Minute

// Connection 跨设备连接数据
type Connection struct {
	Token           string
	UserID          string
	DeviceName      string
	Status          ConnectionStatus
	WebAuthnSession any // webauthn.SessionData
	WebAuthnOptions any // protocol.CredentialCreation
	CreatedAt       time.Time
	ExpiresAt       time.Time
}

// ConnectionStatus 连接状态
type ConnectionStatus string

const (
	ConnectionStatusWaiting   ConnectionStatus = "waiting"
	ConnectionStatusConnected ConnectionStatus = "connected"
	ConnectionStatusConfirmed ConnectionStatus = "confirmed"
	ConnectionStatusRejected  ConnectionStatus = "rejected"
	ConnectionStatusExpired   ConnectionStatus = "expired"
	ConnectionStatusCompleted ConnectionStatus = "completed"
)

// ConnectionStore 跨设备连接存储（数据库持久化）
type ConnectionStore struct {
	repo *repository.ConnectionRepository
}

// NewConnectionStore 创建新的连接存储
func NewConnectionStore(repo *repository.ConnectionRepository) *ConnectionStore {
	return &ConnectionStore{repo: repo}
}

// Store 存储连接
func (s *ConnectionStore) Store(token string, conn *Connection) {
	_ = s.repo.Store(context.Background(), &repository.Connection{
		Token:      token,
		UserID:     conn.UserID,
		DeviceName: conn.DeviceName,
		Status:     string(conn.Status),
		Session:    conn.WebAuthnSession,
		Options:    conn.WebAuthnOptions,
		ExpiresAt:  conn.ExpiresAt,
	})
}

// Get 获取连接
func (s *ConnectionStore) Get(token string) (*Connection, bool) {
	conn, ok, err := s.repo.Get(context.Background(), token)
	if err != nil || !ok {
		return nil, false
	}
	return &Connection{
		Token:           conn.Token,
		UserID:          conn.UserID,
		DeviceName:      conn.DeviceName,
		Status:          ConnectionStatus(conn.Status),
		WebAuthnSession: conn.Session,
		WebAuthnOptions: conn.Options,
		ExpiresAt:       conn.ExpiresAt,
	}, true
}

// Update 更新连接
func (s *ConnectionStore) Update(token string, updater func(*Connection)) bool {
	conn, ok := s.Get(token)
	if !ok {
		return false
	}
	updater(conn)
	s.Store(token, conn)
	return true
}

// Delete 删除连接
func (s *ConnectionStore) Delete(token string) {
	_ = s.repo.Delete(context.Background(), token)
}

// CleanupExpired 清理过期连接
func (s *ConnectionStore) CleanupExpired() {
	_, _ = s.repo.CleanupExpired(context.Background())
}

// StartCleanup 启动定期清理
func (s *ConnectionStore) StartCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				_, _ = s.repo.CleanupExpired(ctx)
			}
		}
	}()
}
