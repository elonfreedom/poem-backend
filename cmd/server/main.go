package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-fuego/fuego"

	"poem-backend/internal/config"
	"poem-backend/internal/router"
	"poem-backend/pkg/database"
	"poem-backend/pkg/response"
)

// 业务错误码映射
const (
	CodeOK            = 0
	CodeBadRequest    = 400
	CodeUnauthorized  = 401
	CodeForbidden     = 403
	CodeNotFound      = 404
	CodeInternalError = 500
)

// vbenErrorSerializer 自定义错误序列化（适配 vben-admin）
// 所有业务错误统一返回 HTTP 200，错误码放在 body 里：
// {"code": 401, "data": null, "error": "未登录", "message": "未登录"}
func vbenErrorSerializer(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	code := CodeBadRequest
	message := err.Error()

	// 提取 Fuego HTTPError 的状态码和标题
	type httpError interface {
		StatusCode() int
		ErrorTitle() string
	}
	if he, ok := err.(httpError); ok {
		code = he.StatusCode()
		message = he.ErrorTitle()
	}

	// 限制在已知错误码范围内
	switch code {
	case CodeBadRequest, CodeUnauthorized, CodeForbidden, CodeNotFound, CodeInternalError:
		// ok
	default:
		code = CodeInternalError
	}

	_ = json.NewEncoder(w).Encode(response.APIResponse[any]{
		Code:    code,
		Message: message,
		Error:   message,
	})
}

// standardErrorSerializer 标准错误序列化（用户端 C 端）
// 返回真正的 HTTP 状态码，前端通过状态码判断请求是否成功
// 成功：HTTP 200，失败：HTTP 400/401/403/404/500
func standardErrorSerializer(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	code := CodeBadRequest
	message := err.Error()

	// 提取 Fuego HTTPError 的状态码和标题
	type httpError interface {
		StatusCode() int
		ErrorTitle() string
	}
	if he, ok := err.(httpError); ok {
		code = he.StatusCode()
		message = he.ErrorTitle()
	}

	// 限制在已知错误码范围内，并映射到 HTTP 状态码
	httpStatus := http.StatusOK
	switch code {
	case CodeBadRequest:
		httpStatus = http.StatusBadRequest
	case CodeUnauthorized:
		httpStatus = http.StatusUnauthorized
	case CodeForbidden:
		httpStatus = http.StatusForbidden
	case CodeNotFound:
		httpStatus = http.StatusNotFound
	default:
		httpStatus = http.StatusInternalServerError
	}

	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(response.APIResponse[any]{
		Code:    code,
		Message: message,
		Error:   message,
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

	// 用户端 API Server - :8080（使用标准 HTTP 状态码）
	userServer := fuego.NewServer(
		fuego.WithAddr(":8080"),
		fuego.WithoutAutoGroupTags(),
		fuego.WithErrorSerializer(standardErrorSerializer),
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

	// 主线程运行 admin server，退出时关闭数据库
	if err := adminServer.Run(); err != nil {
		log.Fatalf("后台服务启动失败: %v", err)
	}
	db.Close()
}
