package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"log"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UserRoleKey contextKey = "user_role"

type JWTClaims struct {
	UserID string `json:"user_id"` // UUID v7
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT Token
func GenerateToken(userID string, role, secret string, expireHour int) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

// AuthMiddleware 认证中间件
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("[Auth] Missing Authorization header, path: %s", r.URL.Path)
				http.Error(w, `{"code":401,"message":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("[Auth] Invalid Authorization format, path: %s, header prefix: %s", r.URL.Path, truncateToken(authHeader))
				http.Error(w, `{"code":401,"message":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := ParseToken(parts[1], secret)
			if err != nil {
				log.Printf("[Auth] Token parse error: %v, path: %s, token prefix: %s", err, r.URL.Path, truncateToken(parts[1]))
				http.Error(w, `{"code":401,"message":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"code":401,"message":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"code":401,"message":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := ParseToken(parts[1], secret)
			if err != nil {
				http.Error(w, `{"code":401,"message":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			if claims.Role != "admin" {
				http.Error(w, `{"code":403,"message":"admin role required"}`, http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext 从 context 获取用户 ID
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// GetUserRoleFromContext 从 context 获取用户角色
func GetUserRoleFromContext(ctx context.Context) string {
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role
	}
	return ""
}

// truncateToken 截断 token 用于安全日志输出（只显示前 10 个字符）
func truncateToken(token string) string {
	if len(token) <= 10 {
		return token
	}
	return token[:10]
}
