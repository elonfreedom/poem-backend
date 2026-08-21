package main

import (
	"log"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/config"
	"poem-backend/internal/router"
	"poem-backend/pkg/database"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 用户端 API Server - :8080
	userServer := fuego.NewServer(
		fuego.WithAddr(":8080"),
		fuego.WithoutAutoGroupTags(),
	)
	router.SetupUserRoutes(userServer, db, cfg)

	// 后台管理 API Server - :8081
	adminServer := fuego.NewServer(
		fuego.WithAddr(":8081"),
		fuego.WithoutAutoGroupTags(),
	)
	router.SetupAdminRoutes(adminServer, db, cfg)

	// 启动两个服务
	log.Printf("晓诗用户端 API 启动于 :8080，文档: http://localhost:8080/swagger")
	log.Printf("晓诗后台管理 API 启动于 :8081，文档: http://localhost:8081/swagger")

	go func() {
		if err := userServer.Run(); err != nil {
			log.Fatalf("用户端服务启动失败: %v", err)
		}
	}()

	// 主线程运行 admin server
	if err := adminServer.Run(); err != nil {
		log.Fatalf("后台服务启动失败: %v", err)
	}
}
