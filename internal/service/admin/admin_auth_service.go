package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"poem-backend/internal/middleware"
	"poem-backend/internal/model"
	adminmodel "poem-backend/internal/model/admin"
	"poem-backend/internal/repository"
)

type AdminAuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
	jwtExpire int
}

func NewAdminAuthService(userRepo *repository.UserRepository, jwtSecret string, jwtExpire int) *AdminAuthService {
	return &AdminAuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpire: jwtExpire,
	}
}

// Login 后台管理员登录（适配 vben-admin，username 即邮箱）
func (s *AdminAuthService) Login(ctx context.Context, username, password string) (*adminmodel.AdminLoginResponse, error) {
	// 查找用户（vben-admin 使用 username 字段，实际为邮箱）
	user, err := s.userRepo.GetByEmailWithPassword(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 检查是否为管理员
	if user.Role != "admin" {
		return nil, fmt.Errorf("admin role required")
	}

	// 检查密码是否设置
	if user.PasswordHash == nil {
		return nil, fmt.Errorf("password not set, please contact super admin")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 生成 JWT
	token, err := middleware.GenerateToken(user.ID, user.Role, s.jwtSecret, s.jwtExpire)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &adminmodel.AdminLoginResponse{
		AccessToken: token,
		User:        adminmodel.AdminUserResponse{
			ID:        user.ID,
			Nickname:  user.Nickname,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

// GetUserInfo 获取管理员用户信息（适配 vben-admin /user/info）
func (s *AdminAuthService) GetUserInfo(ctx context.Context, userID string) (*adminmodel.AdminUserInfoResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return &adminmodel.AdminUserInfoResponse{
		UserId:   user.ID,
		Username: user.Nickname,
		RealName: user.Nickname,
		Avatar:   "",
		Desc:     user.Role,
		HomePath: "/dashboard/workspace",
		Roles: []adminmodel.AdminRoleInfo{
			{RoleName: user.Role, Value: user.Role},
		},
	}, nil
}

// SetPassword 设置密码（用于初始化管理员密码）
func (s *AdminAuthService) SetPassword(ctx context.Context, userID, password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	passwordHash := string(hashedBytes)
	return s.userRepo.UpdatePassword(ctx, userID, passwordHash)
}

// CreateAdminUser 创建管理员用户
func (s *AdminAuthService) CreateAdminUser(ctx context.Context, email, password, nickname string) (*model.UserResponse, error) {
	// 检查邮箱是否已存在
	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// 哈希密码
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	passwordHash := string(hashedBytes)

	// 创建用户
	user := &model.User{
		ID:           uuid.New().String(),
		Nickname:     nickname,
		Email:        &email,
		Role:         "admin",
		PasswordHash: &passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	resp := user.ToResponse()
	return &resp, nil
}
