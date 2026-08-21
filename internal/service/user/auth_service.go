package user

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	"poem-backend/internal/middleware"
	usermodel "poem-backend/internal/model/user"
	"poem-backend/internal/repository"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	passkeyRepo *repository.PasskeyRepository
	webauthn    *webauthn.WebAuthn
	jwtSecret   string
	jwtExpire   int
}

func NewAuthService(
	userRepo *repository.UserRepository,
	passkeyRepo *repository.PasskeyRepository,
	webauthn *webauthn.WebAuthn,
	jwtSecret string,
	jwtExpire int,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		passkeyRepo: passkeyRepo,
		webauthn:    webauthn,
		jwtSecret:   jwtSecret,
		jwtExpire:   jwtExpire,
	}
}

// BeginRegistration 开始注册
func (s *AuthService) BeginRegistration(ctx context.Context, deviceName string) (*protocol.CredentialCreation, *webauthn.SessionData, string, error) {
	// 生成 UUID v7 用户 ID
	userID := uuid.New().String()
	nickname := generateNickname()

	user := &usermodel.User{
		ID:        userID,
		Nickname:  nickname,
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 创建用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// 开始 WebAuthn 注册
	options, session, err := s.webauthn.BeginRegistration(user)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to begin registration: %w", err)
	}

	return options, session, user.ID, nil
}

// FinishRegistration 完成注册
func (s *AuthService) FinishRegistration(ctx context.Context, userID string, session webauthn.SessionData, r *http.Request) (*usermodel.LoginResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 完成 WebAuthn 注册
	credential, err := s.webauthn.FinishRegistration(user, session, r)
	if err != nil {
		return nil, fmt.Errorf("failed to finish registration: %w", err)
	}

	// 保存 Passkey
	passkey := &usermodel.Passkey{
		UserID:       user.ID,
		CredentialID: credential.ID,
		PublicKey:    credential.PublicKey,
		SignCount:    credential.Authenticator.SignCount,
		DeviceName:   "Unknown Device", // 可以从请求中获取
		CreatedAt:    time.Now(),
	}
	if err := s.passkeyRepo.Create(ctx, passkey); err != nil {
		return nil, fmt.Errorf("failed to save passkey: %w", err)
	}

	// 生成 JWT
	token, err := middleware.GenerateToken(user.ID, user.Role, s.jwtSecret, s.jwtExpire)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &usermodel.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// BeginLogin 开始登录
func (s *AuthService) BeginLogin(ctx context.Context) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	// 发现式登录（无需用户名）
	options, session, err := s.webauthn.BeginDiscoverableLogin()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin login: %w", err)
	}

	return options, session, nil
}

// FinishLogin 完成登录
func (s *AuthService) FinishLogin(ctx context.Context, session webauthn.SessionData, r *http.Request) (*usermodel.LoginResponse, error) {
	// 完成 WebAuthn 登录（发现式）
	credential, err := s.webauthn.FinishDiscoverableLogin(s.findUserHandler(ctx), session, r)
	if err != nil {
		return nil, fmt.Errorf("failed to finish login: %w", err)
	}

	// 根据 credential ID 查找 Passkey
	passkey, err := s.passkeyRepo.GetByCredentialID(ctx, credential.ID)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}

	// 获取用户
	user, err := s.userRepo.GetByID(ctx, passkey.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 更新签名计数器
	if credential.Authenticator.SignCount != passkey.SignCount {
		_ = s.passkeyRepo.UpdateSignCount(ctx, passkey.ID, credential.Authenticator.SignCount)
	}

	// 生成 JWT
	token, err := middleware.GenerateToken(user.ID, user.Role, s.jwtSecret, s.jwtExpire)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &usermodel.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// findUserHandler 用于发现式登录的回调
func (s *AuthService) findUserHandler(ctx context.Context) webauthn.DiscoverableUserHandler {
	return func(rawID, userHandle []byte) (webauthn.User, error) {
		userID := string(userHandle)
		return s.userRepo.GetByID(ctx, userID)
	}
}

// generateNickname 生成默认昵称
func generateNickname() string {
	// 简单实现：诗友 + 4 位随机数字
	return fmt.Sprintf("诗友%04d", time.Now().UnixNano()%10000)
}
