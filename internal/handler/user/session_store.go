package user

import (
	"context"
	"time"

	"poem-backend/internal/repository"
)

// sessionTTL 会话有效期
const sessionTTL = 10 * time.Minute

// SessionStore WebAuthn 会话存储（数据库持久化）
type SessionStore struct {
	repo *repository.SessionRepository
}

// NewSessionStore 创建新的会话存储
func NewSessionStore(repo *repository.SessionRepository) *SessionStore {
	return &SessionStore{repo: repo}
}

// Store 存储会话
func (s *SessionStore) Store(id string, data SessionData) {
	var userID *string
	if data.UserID != "" {
		userID = &data.UserID
	}
	// 忽略错误：存储失败只会导致后续 finish 返回 invalid session
	_ = s.repo.Store(context.Background(), id, data.Session, userID, sessionTTL)
}

// Get 获取会话
func (s *SessionStore) Get(id string) (SessionData, bool) {
	session, userID, ok, err := s.repo.Get(context.Background(), id)
	if err != nil || !ok {
		return SessionData{}, false
	}
	data := SessionData{Session: *session}
	if userID != nil {
		data.UserID = *userID
	}
	return data, true
}

// Delete 删除会话
func (s *SessionStore) Delete(id string) {
	_ = s.repo.Delete(context.Background(), id)
}

// StartCleanup 启动定期清理
func (s *SessionStore) StartCleanup(ctx context.Context, interval time.Duration) {
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
