package main

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"

	"poem-backend/internal/config"
	"poem-backend/pkg/database"
)

func main() {
	cfg := config.Load()
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}

	email := "admin@xiaoshi.app"
	password := "admin123456"

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		db.Close()
		log.Fatalf("Hash failed: %v", err)
	}

	tag, err := db.Exec(context.Background(),
		"UPDATE users SET password_hash=$1, updated_at=NOW() WHERE email=$2 AND role='admin'",
		string(hashed), email)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}

	if tag.RowsAffected() == 0 {
		fmt.Println("❌ 未找到 admin@xiaoshi.app 管理员账号")
		return
	}

	fmt.Printf("✅ 超管账号密码已重置\n  邮箱: %s\n  密码: %s\n", email, password)
}
