package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/jackc/pgx/v5"

	adminmodel "poem-backend/internal/model/adminmodel"
	"poem-backend/internal/repository"
)

type AdminUserService struct {
	userRepo        *repository.UserRepository
	checkinRepo     *repository.CheckinRepository
	favoriteRepo    *repository.FavoriteRepository
	readingPlanRepo *repository.ReadingPlanRepository
	passkeyRepo     *repository.PasskeyRepository
}

func NewAdminUserService(
	userRepo *repository.UserRepository,
	checkinRepo *repository.CheckinRepository,
	favoriteRepo *repository.FavoriteRepository,
	readingPlanRepo *repository.ReadingPlanRepository,
	passkeyRepo *repository.PasskeyRepository,
) *AdminUserService {
	return &AdminUserService{
		userRepo:        userRepo,
		checkinRepo:     checkinRepo,
		favoriteRepo:    favoriteRepo,
		readingPlanRepo: readingPlanRepo,
		passkeyRepo:     passkeyRepo,
	}
}

// ListUsers 获取前端用户列表
func (s *AdminUserService) ListUsers(ctx context.Context, page, pageSize int, keyword, statusFilter string) ([]adminmodel.AdminUserListItem, int64, error) {
	users, total, err := s.userRepo.ListUsers(ctx, page, pageSize, keyword, statusFilter)
	if err != nil {
		return nil, 0, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询用户列表失败: %v", err)}
	}

	items := make([]adminmodel.AdminUserListItem, 0, len(users))
	for _, u := range users {
		item := adminmodel.AdminUserListItem{
			ID:        u.ID,
			Nickname:  u.Nickname,
			Role:      u.Role,
			Status:    u.Status,
			CreatedAt: u.CreatedAt,
		}
		if u.Email != nil {
			item.Email = maskEmail(*u.Email)
		}
		items = append(items, item)
	}
	return items, total, nil
}

// GetUserDetail 获取用户详情（含统计数据）
func (s *AdminUserService) GetUserDetail(ctx context.Context, userID string) (*adminmodel.AdminUserDetailResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fuego.NotFoundError{Title: "user not found", Detail: fmt.Sprintf("用户不存在: id=%s", userID)}
		}
		return nil, fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("查询用户失败: id=%s, error=%v", userID, err)}
	}

	detail := &adminmodel.AdminUserDetailResponse{
		ID:        user.ID,
		Nickname:  user.Nickname,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	if user.Email != nil {
		detail.Email = maskEmail(*user.Email)
	}

	// 打卡统计
	if stats, err := s.checkinRepo.GetStats(ctx, userID); err == nil && stats != nil {
		detail.TotalCheckinDays = stats.TotalDays
		detail.ConsecutiveDays = stats.ConsecutiveDay
	}

	// 收藏数量
	if count, err := s.favoriteRepo.CountByUserID(ctx, userID); err == nil {
		detail.FavoriteCount = count
	}

	// 阅读计划数量
	if count, err := s.readingPlanRepo.CountByUserID(ctx, userID); err == nil {
		detail.ReadingPlanCount = count
	}

	// Passkey 数量
	if count, err := s.passkeyRepo.CountByUserID(ctx, userID); err == nil {
		detail.PasskeyCount = count
	}

	return detail, nil
}

// UpdateUserStatus 更新用户状态
func (s *AdminUserService) UpdateUserStatus(ctx context.Context, userID string, status string) error {
	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		return fuego.InternalServerError{Title: "database error", Detail: fmt.Sprintf("更新用户状态失败: id=%s, error=%v", userID, err)}
	}
	return nil
}

// maskEmail 脱敏邮箱
func maskEmail(email string) string {
	at := -1
	for i, c := range email {
		if c == '@' {
			at = i
			break
		}
	}
	if at <= 3 {
		return "***" + email[at:]
	}
	return email[:3] + "***" + email[at:]
}
