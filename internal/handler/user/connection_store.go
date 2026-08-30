package user

import (
	"context"
	"sync"
	"time"

	"poem-backend/internal/repository"
)

// ConnectionStatusTimeout 长轮询超时状态（内部使用）
const ConnectionStatusTimeout ConnectionStatus = "_timeout"

// connectionTTL 连接有效期
const connectionTTL = 10 * time.Minute

// heartbeatTimeout 心跳超时时间（设备 B 超过此时间未活跃则视为离线）
const heartbeatTimeout = 60 * time.Second

// heartbeatCleanupInterval 心跳清理间隔
const heartbeatCleanupInterval = 20 * time.Second

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
	LastActiveAt    time.Time // 设备 B 最后活跃时间
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

// ConnectionStore 跨设备连接存储（数据库持久化 + 内存订阅）
type ConnectionStore struct {
	mu          sync.Mutex
	repo        *repository.ConnectionRepository
	subscribers map[string][]chan ConnectionStatus // token -> 等待状态变化的订阅者
}

// connectionLongPollTimeout 长轮询超时时间
const connectionLongPollTimeout = 30 * time.Second

// NewConnectionStore 创建新的连接存储
func NewConnectionStore(repo *repository.ConnectionRepository) *ConnectionStore {
	return &ConnectionStore{
		repo:        repo,
		subscribers: make(map[string][]chan ConnectionStatus),
	}
}

// Subscribe 订阅连接状态变化，返回状态通道和清理函数
// 当状态变化时，新状态会通过 channel 推送；超时则推送当前状态
func (s *ConnectionStore) Subscribe(token string, timeout time.Duration) (<-chan ConnectionStatus, func()) {
	ch := make(chan ConnectionStatus, 1)

	s.mu.Lock()
	s.subscribers[token] = append(s.subscribers[token], ch)
	s.mu.Unlock()

	cleanup := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, c := range s.subscribers[token] {
			if c == ch {
				s.subscribers[token] = append(s.subscribers[token][:i], s.subscribers[token][i+1:]...)
				break
			}
		}
		if len(s.subscribers[token]) == 0 {
			delete(s.subscribers, token)
		}
		close(ch)
	}

	// 超时兜底：超时后推送当前状态
	go func() {
		time.Sleep(timeout)
		s.mu.Lock()
		defer s.mu.Unlock()
		// 检查是否已取消订阅
		for _, c := range s.subscribers[token] {
			if c == ch {
				select {
				case ch <- ConnectionStatusTimeout:
				default:
				}
				break
			}
		}
	}()

	return ch, cleanup
}

// notifySubscribers 通知所有订阅者状态变化
func (s *ConnectionStore) notifySubscribers(token string, status ConnectionStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.subscribers[token] {
		select {
		case ch <- status:
		default:
		}
	}
	delete(s.subscribers, token)
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
		LastActiveAt:    conn.LastActiveAt,
	}, true
}

// Update 更新连接（状态变化时通知订阅者）
func (s *ConnectionStore) Update(token string, updater func(*Connection)) bool {
	conn, ok := s.Get(token)
	if !ok {
		return false
	}
	oldStatus := conn.Status
	updater(conn)
	s.Store(token, conn)
	// 状态变化时通知长轮询订阅者
	if conn.Status != oldStatus {
		s.notifySubscribers(token, conn.Status)
	}
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

// StartCleanup 启动定期清理（过期连接 + 心跳超时）
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
				_, _ = s.repo.CleanupInactive(ctx, heartbeatTimeout)
			}
		}
	}()
}

// UpdateHeartbeat 更新设备 B 心跳
func (s *ConnectionStore) UpdateHeartbeat(token string) {
	_ = s.repo.UpdateHeartbeat(context.Background(), token)
}
