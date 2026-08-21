package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/model"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, nickname, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		user.ID, user.Nickname, user.Email, user.Role, user.CreatedAt, user.UpdatedAt)
	return err
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, nickname, email, role, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	var user model.User
	err := row.Scan(&user.ID, &user.Nickname, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, nickname, email, role, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)
	var user model.User
	err := row.Scan(&user.ID, &user.Nickname, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users SET nickname = $1, email = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, user.Nickname, user.Email, user.UpdatedAt, user.ID)
	return err
}

// UpdateEmail 更新邮箱
func (r *UserRepository) UpdateEmail(ctx context.Context, userID string, email *string) error {
	query := `UPDATE users SET email = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, email, userID)
	return err
}

// GetByEmailWithPassword 根据邮箱获取用户（包含密码哈希，用于后台登录）
func (r *UserRepository) GetByEmailWithPassword(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, nickname, email, role, password_hash, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)
	var user model.User
	err := row.Scan(&user.ID, &user.Nickname, &user.Email, &user.Role, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePassword 更新密码
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, passwordHash, userID)
	return err
}
