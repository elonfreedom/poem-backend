package router

import (
	"github.com/go-fuego/fuego"
	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/config"
	"poem-backend/internal/handler/admin"
	"poem-backend/internal/middleware"
	"poem-backend/internal/repository"
	adminservice "poem-backend/internal/service/admin"
)

// SetupAdminRoutes 注册后台管理路由 - 端口 8081
func SetupAdminRoutes(server *fuego.Server, db *pgxpool.Pool, cfg *config.Config) {
	// 初始化 Repository
	userRepo := repository.NewUserRepository(db)
	poemRepo := repository.NewPoemRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	tagRepo := repository.NewTagRepository(db)
	bannerRepo := repository.NewBannerRepository(db)
	announcementRepo := repository.NewAnnouncementRepository(db)
	configRepo := repository.NewConfigRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	checkinRepo := repository.NewCheckinRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	readingPlanRepo := repository.NewReadingPlanRepository(db)
	passkeyRepo := repository.NewPasskeyRepository(db)

	// 初始化 Service
	adminAuthService := adminservice.NewAdminAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	adminPoemService := adminservice.NewAdminPoemService(poemRepo)
	adminCategoryService := adminservice.NewAdminCategoryService(categoryRepo)
	adminTagService := adminservice.NewAdminTagService(tagRepo)
	adminStatsService := adminservice.NewAdminStatsService(statsRepo)
	adminBannerService := adminservice.NewAdminBannerService(bannerRepo)
	adminAnnouncementService := adminservice.NewAdminAnnouncementService(announcementRepo)
	adminConfigService := adminservice.NewAdminConfigService(configRepo)
	adminUserService := adminservice.NewAdminUserService(userRepo, checkinRepo, favoriteRepo, readingPlanRepo, passkeyRepo)

	// 初始化 Handler
	authHandler := admin.NewAuthHandler(adminAuthService)
	poemHandler := admin.NewPoemHandler(adminPoemService)
	categoryHandler := admin.NewCategoryHandler(adminCategoryService)
	tagHandler := admin.NewTagHandler(adminTagService)
	statsHandler := admin.NewStatsHandler(adminStatsService)
	bannerHandler := admin.NewBannerHandler(adminBannerService)
	announcementHandler := admin.NewAnnouncementHandler(adminAnnouncementService)
	configHandler := admin.NewConfigHandler(adminConfigService)
	userHandler := admin.NewUserHandler(adminUserService)

	// ========== 后台管理路由组（端口 8081，全部以 /api/admin 开头）==========
	adminGroup := fuego.Group(server, "/api/admin")

	// 公开：管理员登录（无需鉴权）
	fuego.Post(adminGroup, "/auth/login", authHandler.Login,
		fuego.OptionSummary("管理员登录"),
		fuego.OptionOverrideDescription("后台管理员使用邮箱和密码登录，返回 JWT"),
		fuego.OptionTags("后台认证"),
	)

	// 需鉴权路由（登录后访问）
	adminAuth := fuego.Group(server, "/api/admin")
	fuego.Use(adminAuth, middleware.AdminAuthMiddleware(cfg.JWT.Secret))

	// [后台认证] 用户信息（vben-admin /user/info）
	fuego.Get(adminAuth, "/user/info", authHandler.GetUserInfo,
		fuego.OptionSummary("获取用户信息"),
		fuego.OptionOverrideDescription("获取当前登录管理员的个人信息"),
		fuego.OptionTags("后台认证"),
	)

	// [后台认证] 权限码
	fuego.Get(adminAuth, "/auth/codes", authHandler.GetAccessCodes,
		fuego.OptionSummary("获取权限码"),
		fuego.OptionOverrideDescription("获取当前管理员的权限码列表"),
		fuego.OptionTags("后台认证"),
	)

	// [后台认证] 退出登录
	fuego.Post(adminAuth, "/auth/logout", authHandler.Logout,
		fuego.OptionSummary("退出登录"),
		fuego.OptionOverrideDescription("管理员退出登录"),
		fuego.OptionTags("后台认证"),
	)

	// ========== 后台管理路由 ==========
	adminMgmt := adminAuth

	// [诗歌管理]
	fuego.Post(adminMgmt, "/poems/import", poemHandler.ImportPoems,
		fuego.OptionSummary("导入诗歌"),
		fuego.OptionOverrideDescription("通过 JSON 文件批量导入诗歌数据"),
		fuego.OptionTags("诗歌管理"),
	)
	fuego.Get(adminMgmt, "/poems", poemHandler.List,
		fuego.OptionSummary("获取诗歌列表"),
		fuego.OptionOverrideDescription("分页获取诗歌列表，支持筛选"),
		fuego.OptionTags("诗歌管理"),
	)
	fuego.Post(adminMgmt, "/poems", poemHandler.Create,
		fuego.OptionSummary("创建诗歌"),
		fuego.OptionOverrideDescription("创建新诗歌"),
		fuego.OptionTags("诗歌管理"),
	)
	fuego.Get(adminMgmt, "/poems/{id}", poemHandler.GetByID,
		fuego.OptionSummary("获取诗歌详情"),
		fuego.OptionOverrideDescription("根据 ID 获取诗歌详情"),
		fuego.OptionTags("诗歌管理"),
	)
	fuego.Put(adminMgmt, "/poems/{id}", poemHandler.Update,
		fuego.OptionSummary("更新诗歌"),
		fuego.OptionOverrideDescription("更新诗歌信息"),
		fuego.OptionTags("诗歌管理"),
	)
	fuego.Delete(adminMgmt, "/poems/{id}", poemHandler.Delete,
		fuego.OptionSummary("删除诗歌"),
		fuego.OptionOverrideDescription("删除指定诗歌"),
		fuego.OptionTags("诗歌管理"),
	)
	fuego.Put(adminMgmt, "/poems/{id}/status", poemHandler.UpdateStatus,
		fuego.OptionSummary("更新诗歌状态"),
		fuego.OptionOverrideDescription("更新诗歌状态（草稿/发布/归档）"),
		fuego.OptionTags("诗歌管理"),
	)
	fuego.Put(adminMgmt, "/poems/batch/status", poemHandler.BatchUpdateStatus,
		fuego.OptionSummary("批量更新诗歌状态"),
		fuego.OptionOverrideDescription("批量更新多首诗歌的状态"),
		fuego.OptionTags("诗歌管理"),
	)
	fuego.Post(adminMgmt, "/poems/batch/convert-simplified", poemHandler.BatchConvertSimplified,
		fuego.OptionSummary("批量生成简体"),
		fuego.OptionOverrideDescription("一键为存量诗歌自动生成简体（繁体→简体），扫描 title_sc 为空的记录进行处理"),
		fuego.OptionTags("诗歌管理"),
	)

	// [分类管理]
	fuego.Get(adminMgmt, "/categories", categoryHandler.List,
		fuego.OptionSummary("获取分类列表"),
		fuego.OptionOverrideDescription("获取所有分类"),
		fuego.OptionTags("分类管理"),
	)
	fuego.Post(adminMgmt, "/categories", categoryHandler.Create,
		fuego.OptionSummary("创建分类"),
		fuego.OptionOverrideDescription("创建新分类"),
		fuego.OptionTags("分类管理"),
	)
	fuego.Put(adminMgmt, "/categories/{id}", categoryHandler.Update,
		fuego.OptionSummary("更新分类"),
		fuego.OptionOverrideDescription("更新分类信息"),
		fuego.OptionTags("分类管理"),
	)
	fuego.Delete(adminMgmt, "/categories/{id}", categoryHandler.Delete,
		fuego.OptionSummary("删除分类"),
		fuego.OptionOverrideDescription("删除指定分类"),
		fuego.OptionTags("分类管理"),
	)

	// [标签管理]
	fuego.Get(adminMgmt, "/tags", tagHandler.List,
		fuego.OptionSummary("获取标签列表"),
		fuego.OptionOverrideDescription("获取所有标签"),
		fuego.OptionTags("标签管理"),
	)
	fuego.Post(adminMgmt, "/tags", tagHandler.Create,
		fuego.OptionSummary("创建标签"),
		fuego.OptionOverrideDescription("创建新标签"),
		fuego.OptionTags("标签管理"),
	)
	fuego.Delete(adminMgmt, "/tags/{id}", tagHandler.Delete,
		fuego.OptionSummary("删除标签"),
		fuego.OptionOverrideDescription("删除指定标签"),
		fuego.OptionTags("标签管理"),
	)

	// [数据统计]
	fuego.Get(adminMgmt, "/stats/overview", statsHandler.Overview,
		fuego.OptionSummary("总览数据"),
		fuego.OptionOverrideDescription("获取平台总览统计数据"),
		fuego.OptionTags("数据统计"),
	)
	fuego.Get(adminMgmt, "/stats/daily", statsHandler.Daily,
		fuego.OptionSummary("每日统计"),
		fuego.OptionOverrideDescription("获取每日统计数据"),
		fuego.OptionTags("数据统计"),
	)
	fuego.Get(adminMgmt, "/stats/poems/hot", statsHandler.HotPoems,
		fuego.OptionSummary("热门诗歌"),
		fuego.OptionOverrideDescription("获取热门诗歌排行"),
		fuego.OptionTags("数据统计"),
	)
	fuego.Get(adminMgmt, "/stats/users/growth", statsHandler.UserGrowth,
		fuego.OptionSummary("用户增长"),
		fuego.OptionOverrideDescription("获取用户增长数据"),
		fuego.OptionTags("数据统计"),
	)

	// [Banner 管理]
	fuego.Get(adminMgmt, "/banners", bannerHandler.List,
		fuego.OptionSummary("获取 Banner 列表"),
		fuego.OptionOverrideDescription("获取所有 Banner"),
		fuego.OptionTags("Banner 管理"),
	)
	fuego.Post(adminMgmt, "/banners", bannerHandler.Create,
		fuego.OptionSummary("创建 Banner"),
		fuego.OptionOverrideDescription("创建新 Banner"),
		fuego.OptionTags("Banner 管理"),
	)
	fuego.Put(adminMgmt, "/banners/{id}", bannerHandler.Update,
		fuego.OptionSummary("更新 Banner"),
		fuego.OptionOverrideDescription("更新 Banner 信息"),
		fuego.OptionTags("Banner 管理"),
	)
	fuego.Delete(adminMgmt, "/banners/{id}", bannerHandler.Delete,
		fuego.OptionSummary("删除 Banner"),
		fuego.OptionOverrideDescription("删除指定 Banner"),
		fuego.OptionTags("Banner 管理"),
	)

	// [公告管理]
	fuego.Get(adminMgmt, "/announcements", announcementHandler.List,
		fuego.OptionSummary("获取公告列表"),
		fuego.OptionOverrideDescription("获取所有公告"),
		fuego.OptionTags("公告管理"),
	)
	fuego.Post(adminMgmt, "/announcements", announcementHandler.Create,
		fuego.OptionSummary("创建公告"),
		fuego.OptionOverrideDescription("创建新公告"),
		fuego.OptionTags("公告管理"),
	)
	fuego.Put(adminMgmt, "/announcements/{id}", announcementHandler.Update,
		fuego.OptionSummary("更新公告"),
		fuego.OptionOverrideDescription("更新公告信息"),
		fuego.OptionTags("公告管理"),
	)
	fuego.Delete(adminMgmt, "/announcements/{id}", announcementHandler.Delete,
		fuego.OptionSummary("删除公告"),
		fuego.OptionOverrideDescription("删除指定公告"),
		fuego.OptionTags("公告管理"),
	)

	// [系统配置]
	fuego.Get(adminMgmt, "/config", configHandler.List,
		fuego.OptionSummary("获取配置列表"),
		fuego.OptionOverrideDescription("获取系统配置列表"),
		fuego.OptionTags("系统配置"),
	)
	fuego.Get(adminMgmt, "/config/{key}", configHandler.GetByKey,
		fuego.OptionSummary("获取单个配置"),
		fuego.OptionOverrideDescription("根据 key 获取配置值"),
		fuego.OptionTags("系统配置"),
	)
	fuego.Put(adminMgmt, "/config", configHandler.Update,
		fuego.OptionSummary("更新配置"),
		fuego.OptionOverrideDescription("更新系统配置"),
		fuego.OptionTags("系统配置"),
	)

	// [用户管理] 前端 App 用户
	fuego.Get(adminMgmt, "/users", userHandler.List,
		fuego.OptionSummary("获取用户列表"),
		fuego.OptionOverrideDescription("分页获取前端用户列表，支持搜索和状态筛选"),
		fuego.OptionTags("用户管理"),
	)
	fuego.Get(adminMgmt, "/users/{id}", userHandler.GetByID,
		fuego.OptionSummary("获取用户详情"),
		fuego.OptionOverrideDescription("获取用户详情（含统计数据）"),
		fuego.OptionTags("用户管理"),
	)
	fuego.Put(adminMgmt, "/users/{id}/status", userHandler.UpdateStatus,
		fuego.OptionSummary("更新用户状态"),
		fuego.OptionOverrideDescription("禁用或启用用户账号"),
		fuego.OptionTags("用户管理"),
	)
}
