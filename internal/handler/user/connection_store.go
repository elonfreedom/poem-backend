package user

import (
	"sync"
	"time"
)

// ConnectionStatus 连接状态
type ConnectionStatus string

const (
	ConnectionStatusWaiting   ConnectionStatus = "waiting"   // 等待设备 B 连接
	ConnectionStatusConnected ConnectionStatus = "connected" // 设备 B 已连接
	ConnectionStatusConfirmed ConnectionStatus = "confirmed" // 设备 A 已确认
	ConnectionStatusRejected  ConnectionStatus = "rejected"  // 设备 A 已拒绝
	ConnectionStatusExpired   ConnectionStatus = "expired"   // 连接已过期
	ConnectionStatusCompleted ConnectionStatus = "completed" // 注册完成
)

// Connection 跨设备连接数据
type Connection struct {
	Token           string           // 连接令牌（UUID）
	UserID          string           // 设备 A 的用户 ID
	DeviceName      string           // 设备 B 的设备名称
	Status          ConnectionStatus // 当前状态
	WebAuthnSession any              // webauthn.SessionData（finish 时验证 credential）
	WebAuthnOptions any              // protocol.CredentialCreation（设备 B 创建 credential 用）
	CreatedAt       time.Time        // 创建时间
	ExpiresAt       time.Time        // 过期时间
}

// ConnectionStore 线程安全的跨设备连接存储
type ConnectionStore struct {
	mu          sync.RWMutex
	connections map[string]*Connection
}

// NewConnectionStore 创建新的连接存储
func NewConnectionStore() *ConnectionStore {
	return &ConnectionStore{
		connections: make(map[string]*Connection),
	}
}

// Store 存储连接
func (s *ConnectionStore) Store(token string, conn *Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connections[token] = conn
}

// Get 获取连接
func (s *ConnectionStore) Get(token string) (*Connection, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conn, ok := s.connections[token]
	return conn, ok
}

// Update 更新连接
func (s *ConnectionStore) Update(token string, updater func(*Connection)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if conn, ok := s.connections[token]; ok {
		updater(conn)
		return true
	}
	return false
}

// Delete 删除连接
func (s *ConnectionStore) Delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.connections, token)
}

// CleanupExpired 清理过期连接
func (s *ConnectionStore) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for token, conn := range s.connections {
		if now.After(conn.ExpiresAt) {
			delete(s.connections, token)
		}
	}
}
