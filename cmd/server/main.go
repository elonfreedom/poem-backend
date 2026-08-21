package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/config"
	"poem-backend/internal/router"
	"poem-backend/pkg/database"
)

// vbenErrorSerializer 自定义错误序列化（适配 vben-admin {code, message} 格式）
func vbenErrorSerializer(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	status := http.StatusInternalServerError
	message := err.Error()

	// 尝试提取 Fuego HTTPError 的状态码和标题
	type httpError interface {
		StatusCode() int
		ErrorTitle() string
	}
	if he, ok := err.(httpError); ok {
		status = he.StatusCode()
		message = he.ErrorTitle()
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"code":    status,
		"message": message,
	})
}

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
		fuego.WithErrorSerializer(vbenErrorSerializer),
	)
	router.SetupUserRoutes(userServer, db, cfg)

	// 后台管理 API Server - :8081
	adminServer := fuego.NewServer(
		fuego.WithAddr(":8081"),
		fuego.WithoutAutoGroupTags(),
		fuego.WithErrorSerializer(vbenErrorSerializer),
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
