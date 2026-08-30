package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/jackc/pgx/v5"

	usermodel "poem-backend/internal/model/user"
	"poem-backend/internal/repository"
)

type UserService struct {
	userRepo    *repository.UserRepository
	passkeyRepo *repository.PasskeyRepository
}

func NewUserService(
	userRepo *repository.UserRepository,
	passkeyRepo *repository.PasskeyRepository,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		passkeyRepo: passkeyRepo,
	}
}

// GetProfile 获取个人信息
func (s *UserService) GetProfile(ctx context.Context, userID string) (*usermodel.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "user not found", Detail: fmt.Sprintf("用户不存在: id=%s", userID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询用户失败: %v", err)}
	}

	resp := user.ToResponse()
	return &resp, nil
}

// UpdateProfile 更新个人信息
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *usermodel.UpdateProfileRequest) (*usermodel.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "user not found", Detail: fmt.Sprintf("用户不存在: id=%s", userID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询用户失败: %v", err)}
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
		user.UpdatedAt = time.Now()
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("更新用户失败: %v", err)}
	}

	resp := user.ToResponse()
	return &resp, nil
}

// BindEmail 绑定邮箱
func (s *UserService) BindEmail(ctx context.Context, userID string, email string) error {
	// 检查邮箱是否已被绑定
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil && existing.ID != userID {
		return fuego.BadRequestError{Title: "email conflict", Detail: "邮箱已被其他账户绑定"}
	}

	if err := s.userRepo.UpdateEmail(ctx, userID, &email); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("绑定邮箱失败: %v", err)}
	}
	return nil
}

// GetPasskeys 获取 Passkey 列表
func (s *UserService) GetPasskeys(ctx context.Context, userID string) ([]usermodel.PasskeyResponse, error) {
	passkeys, err := s.passkeyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询 Passkey 失败: %v", err)}
	}

	var responses []usermodel.PasskeyResponse
	for _, p := range passkeys {
		responses = append(responses, p.ToResponse())
	}
	return responses, nil
}

// DeletePasskey 删除 Passkey
func (s *UserService) DeletePasskey(ctx context.Context, userID string, passkeyID int64) error {
	// 检查是否至少保留一个 Passkey
	count, err := s.passkeyRepo.CountByUserID(ctx, userID)
	if err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询 Passkey 数量失败: %v", err)}
	}
	if count <= 1 {
		return fuego.BadRequestError{Title: "last passkey", Detail: "至少保留一个 Passkey"}
	}

	if err := s.passkeyRepo.Delete(ctx, passkeyID, userID); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("删除 Passkey 失败: %v", err)}
	}
	return nil
}
