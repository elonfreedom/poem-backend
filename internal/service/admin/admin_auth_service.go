package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"poem-backend/internal/middleware"
	adminmodel "poem-backend/internal/model/admin"
	usermodel "poem-backend/internal/model/user"
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
	// 去除首尾空格（前端可能带入空格）
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	// 查找用户（vben-admin 使用 username 字段，实际为邮箱）
	user, err := s.userRepo.GetByEmailWithPassword(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.UnauthorizedError{Title: "login failed", Detail: "用户名或密码错误"}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询用户失败: %v", err)}
	}

	// 检查是否为管理员
	if user.Role != "admin" {
		return nil, fuego.ForbiddenError{Title: "access denied", Detail: "需要管理员权限"}
	}

	// 检查密码是否设置
	if user.PasswordHash == nil {
		return nil, fuego.BadRequestError{Title: "password not set", Detail: "请联系超级管理员设置密码"}
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, fuego.UnauthorizedError{Title: "login failed", Detail: "用户名或密码错误"}
	}

	// 生成 JWT
	token, err := middleware.GenerateToken(user.ID, user.Role, s.jwtSecret, s.jwtExpire)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "token error", Detail: fmt.Sprintf("生成令牌失败: %v", err)}
	}

	return &adminmodel.AdminLoginResponse{
		AccessToken: token,
		User: adminmodel.AdminUserResponse{
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "user not found", Detail: fmt.Sprintf("用户不存在: id=%s", userID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询用户失败: %v", err)}
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
		return fuego.InternalServerError{Title: "password error", Detail: fmt.Sprintf("密码哈希失败: %v", err)}
	}

	passwordHash := string(hashedBytes)
	if err := s.userRepo.UpdatePassword(ctx, userID, passwordHash); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("更新密码失败: %v", err)}
	}
	return nil
}

// CreateAdminUser 创建管理员用户
func (s *AdminAuthService) CreateAdminUser(ctx context.Context, email, password, nickname string) (*usermodel.UserResponse, error) {
	// 检查邮箱是否已存在
	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, fuego.BadRequestError{Title: "email conflict", Detail: "邮箱已被注册"}
	}

	// 哈希密码
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "password error", Detail: fmt.Sprintf("密码哈希失败: %v", err)}
	}
	passwordHash := string(hashedBytes)

	// 创建用户
	user := &usermodel.User{
		ID:           uuid.New().String(),
		Nickname:     nickname,
		Email:        &email,
		Role:         "admin",
		PasswordHash: &passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("创建管理员失败: %v", err)}
	}

	resp := user.ToResponse()
	return &resp, nil
}
