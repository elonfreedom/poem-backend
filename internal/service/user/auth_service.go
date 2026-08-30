package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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
func (s *AuthService) BeginRegistration(ctx context.Context, deviceName string) (*protocol.CredentialCreation, *webauthn.SessionData, string, string, error) {
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
		return nil, nil, "", "", fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("创建用户失败: %v", err)}
	}

	// 开始 WebAuthn 注册
	options, session, err := s.webauthn.BeginRegistration(user)
	if err != nil {
		return nil, nil, "", "", fuego.InternalServerError{Title: "webauthn error", Detail: fmt.Sprintf("初始化注册失败: %v", err)}
	}

	// 生成会话 ID
	sessionID := uuid.New().String()

	return options, session, user.ID, sessionID, nil
}

// FinishRegistration 完成注册
func (s *AuthService) FinishRegistration(ctx context.Context, userID string, session webauthn.SessionData, r *http.Request) (*usermodel.LoginResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "user not found", Detail: fmt.Sprintf("用户不存在: id=%s", userID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询用户失败: %v", err)}
	}

	// 完成 WebAuthn 注册
	credential, err := s.webauthn.FinishRegistration(user, session, r)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "webauthn failed", Detail: fmt.Sprintf("注册验证失败: %v", err)}
	}

	// 保存 Passkey
	passkey := &usermodel.Passkey{
		UserID:       user.ID,
		CredentialID: credential.ID,
		PublicKey:    credential.PublicKey,
		SignCount:    credential.Authenticator.SignCount,
		Flags:        credential.Flags.MsgpByte(),
		DeviceName:   "Unknown Device", // 可以从请求中获取
		CreatedAt:    time.Now(),
	}
	if err := s.passkeyRepo.Create(ctx, passkey); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("保存 Passkey 失败: %v", err)}
	}

	// 生成 JWT
	token, err := middleware.GenerateToken(user.ID, user.Role, s.jwtSecret, s.jwtExpire)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "token error", Detail: fmt.Sprintf("生成令牌失败: %v", err)}
	}

	return &usermodel.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// BeginLogin 开始登录
func (s *AuthService) BeginLogin(ctx context.Context) (*protocol.CredentialAssertion, *webauthn.SessionData, string, error) {
	// 发现式登录（无需用户名）
	options, session, err := s.webauthn.BeginDiscoverableLogin()
	if err != nil {
		return nil, nil, "", fuego.InternalServerError{Title: "webauthn error", Detail: fmt.Sprintf("初始化登录失败: %v", err)}
	}

	// 生成会话 ID
	sessionID := uuid.New().String()

	return options, session, sessionID, nil
}

// FinishLogin 完成登录
func (s *AuthService) FinishLogin(ctx context.Context, session webauthn.SessionData, r *http.Request) (*usermodel.LoginResponse, error) {
	// 完成 WebAuthn 登录（发现式）
	credential, err := s.webauthn.FinishDiscoverableLogin(s.findUserHandler(ctx), session, r)
	if err != nil {
		return nil, fuego.UnauthorizedError{Title: "login failed", Detail: fmt.Sprintf("登录验证失败: %v", err)}
	}

	// 根据 credential ID 查找 Passkey
	passkey, err := s.passkeyRepo.GetByCredentialID(ctx, credential.ID)
	if err != nil {
		return nil, fuego.UnauthorizedError{Title: "login failed", Detail: "凭证不存在"}
	}

	// 获取用户
	user, err := s.userRepo.GetByID(ctx, passkey.UserID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询用户失败: %v", err)}
	}

	// 更新签名计数器
	if credential.Authenticator.SignCount != passkey.SignCount {
		_ = s.passkeyRepo.UpdateSignCount(ctx, passkey.ID, credential.Authenticator.SignCount)
	}

	// 生成 JWT
	token, err := middleware.GenerateToken(user.ID, user.Role, s.jwtSecret, s.jwtExpire)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "token error", Detail: fmt.Sprintf("生成令牌失败: %v", err)}
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
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		passkeys, err := s.passkeyRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		user.Credentials = make([]webauthn.Credential, 0, len(passkeys))
		for _, pk := range passkeys {
			flags := webauthn.CredentialFlagsFromMsgpByte(pk.Flags)
			user.Credentials = append(user.Credentials, webauthn.Credential{
				ID:              pk.CredentialID,
				PublicKey:       pk.PublicKey,
				AttestationType: "none",
				Flags:           flags,
				Authenticator: webauthn.Authenticator{
					SignCount: pk.SignCount,
				},
			})
		}
		return user, nil
	}
}

// generateNickname 生成默认昵称
func generateNickname() string {
	// 简单实现：诗友 + 4 位随机数字
	return fmt.Sprintf("诗友%04d", time.Now().UnixNano()%10000)
}

// ==================== 跨设备 Passkey ====================

// BeginAddDevice 开始添加新设备
// 返回：连接令牌、WebAuthn 注册选项、会话数据、过期时间
func (s *AuthService) BeginAddDevice(ctx context.Context, userID string, deviceName string) (string, *protocol.CredentialCreation, *webauthn.SessionData, time.Time, error) {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, nil, time.Time{}, fuego.NotFoundError{Title: "user not found", Detail: fmt.Sprintf("用户不存在: id=%s", userID)}
		}
		return "", nil, nil, time.Time{}, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询用户失败: %v", err)}
	}

	// 查询用户已有的 Passkey 凭证，用于 excludeCredentials
	// 避免 iCloud Keychain 等同步场景下浏览器返回 NotAllowedError
	excludeList, err := s.buildExcludeList(ctx, userID)
	if err != nil {
		return "", nil, nil, time.Time{}, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询已有凭证失败: %v", err)}
	}

	// 开始 WebAuthn 注册（为现有用户添加新 credential，排除已有凭证）
	options, session, err := s.webauthn.BeginRegistration(user, webauthn.WithExclusions(excludeList))
	if err != nil {
		return "", nil, nil, time.Time{}, fuego.InternalServerError{Title: "webauthn error", Detail: fmt.Sprintf("初始化注册失败: %v", err)}
	}

	// 生成连接令牌（10分钟有效）
	token := uuid.New().String()
	expiresAt := time.Now().Add(10 * time.Minute)

	return token, options, session, expiresAt, nil
}

// buildExcludeList 构建 excludeCredentials 列表
// 将用户已有的 Passkey 凭证 ID 转换为 WebAuthn CredentialDescriptor
func (s *AuthService) buildExcludeList(ctx context.Context, userID string) ([]protocol.CredentialDescriptor, error) {
	passkeys, err := s.passkeyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询 Passkey 失败: %v", err)}
	}

	excludeList := make([]protocol.CredentialDescriptor, 0, len(passkeys))
	for _, pk := range passkeys {
		excludeList = append(excludeList, protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: pk.CredentialID,
		})
	}
	return excludeList, nil
}

// FinishAddDevice 完成新设备注册
// 使用 WebAuthn 库验证 credential 并提取公钥
func (s *AuthService) FinishAddDevice(ctx context.Context, userID string, session webauthn.SessionData, r *http.Request, deviceName string) (*usermodel.LoginResponse, error) {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "user not found", Detail: fmt.Sprintf("用户不存在: id=%s", userID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询用户失败: %v", err)}
	}

	// 完成 WebAuthn 注册（验证 credential 并提取公钥）
	credential, err := s.webauthn.FinishRegistration(user, session, r)
	if err != nil {
		return nil, fuego.BadRequestError{Title: "webauthn failed", Detail: fmt.Sprintf("凭证验证失败: %v", err)}
	}

	// 保存 Passkey
	passkey := &usermodel.Passkey{
		UserID:       user.ID,
		CredentialID: credential.ID,
		PublicKey:    credential.PublicKey,
		SignCount:    credential.Authenticator.SignCount,
		Flags:        credential.Flags.MsgpByte(),
		DeviceName:   deviceName,
		CreatedAt:    time.Now(),
	}
	if err := s.passkeyRepo.Create(ctx, passkey); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("保存 Passkey 失败: %v", err)}
	}

	// 生成 JWT
	token, err := middleware.GenerateToken(user.ID, user.Role, s.jwtSecret, s.jwtExpire)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "token error", Detail: fmt.Sprintf("生成令牌失败: %v", err)}
	}

	return &usermodel.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}
