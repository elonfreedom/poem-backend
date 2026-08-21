package router

import (
	"github.com/go-fuego/fuego"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/config"
	"poem-backend/internal/handler/admin"
	"poem-backend/internal/handler/user"
	"poem-backend/internal/middleware"
	"poem-backend/internal/repository"
	adminservice "poem-backend/internal/service/admin"
	userservice "poem-backend/internal/service/user"
)

// initDependencies 初始化所有依赖（Repository、Service、Handler）
func initDependencies(server *fuego.Server, db *pgxpool.Pool, cfg *config.Config) (
	*user.AuthHandler,
	*user.UserHandler,
	*user.PoemHandler,
	*user.FavoriteHandler,
	*user.ReadingPlanHandler,
	*user.CheckinHandler,
) {
	// 初始化 WebAuthn
	wn, err := webauthn.New(&webauthn.Config{
		RPDisplayName: cfg.WebAuthn.RPDisplayName,
		RPID:          cfg.WebAuthn.RPID,
		RPOrigins:     []string{cfg.WebAuthn.RPOrigin},
	})
	if err != nil {
		panic("failed to initialize WebAuthn: " + err.Error())
	}

	// 初始化 Repository
	userRepo := repository.NewUserRepository(db)
	passkeyRepo := repository.NewPasskeyRepository(db)
	poemRepo := repository.NewPoemRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	readingPlanRepo := repository.NewReadingPlanRepository(db)
	checkinRepo := repository.NewCheckinRepository(db)

	// 初始化 Service
	authService := userservice.NewAuthService(userRepo, passkeyRepo, wn, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	userService := userservice.NewUserService(userRepo, passkeyRepo)
	poemService := userservice.NewPoemService(poemRepo)
	favoriteService := userservice.NewFavoriteService(favoriteRepo, poemRepo)
	readingPlanService := userservice.NewReadingPlanService(readingPlanRepo)
	checkinService := userservice.NewCheckinService(checkinRepo)

	// 初始化 Handler
	authHandler := user.NewAuthHandler(authService)
	userHandler := user.NewUserHandler(userService)
	poemHandler := user.NewPoemHandler(poemService)
	favoriteHandler := user.NewFavoriteHandler(favoriteService)
	readingPlanHandler := user.NewReadingPlanHandler(readingPlanService)
	checkinHandler := user.NewCheckinHandler(checkinService)

	return authHandler, userHandler, poemHandler, favoriteHandler, readingPlanHandler, checkinHandler
}

// SetupUserRoutes 注册用户端（C端）路由 - 端口 8080
func SetupUserRoutes(server *fuego.Server, db *pgxpool.Pool, cfg *config.Config) {
	authHandler, userHandler, poemHandler, favoriteHandler, readingPlanHandler, checkinHandler :=
		initDependencies(server, db, cfg)

	// ========== 公开路由：Passkey 认证 ==========
	public := fuego.Group(server, "/api/public")

	fuego.Post(public, "/passkey/register/begin", authHandler.BeginRegistration,
		fuego.OptionSummary("开始注册"),
		fuego.OptionOverrideDescription("使用 Passkey 开始注册流程，返回 WebAuthn 挑战"),
		fuego.OptionTags("Passkey 认证"),
	)
	fuego.Post(public, "/passkey/register/finish", authHandler.FinishRegistration,
		fuego.OptionSummary("完成注册"),
		fuego.OptionOverrideDescription("验证 Passkey 并完成注册，返回 JWT"),
		fuego.OptionTags("Passkey 认证"),
	)
	fuego.Post(public, "/passkey/login/begin", authHandler.BeginLogin,
		fuego.OptionSummary("开始登录"),
		fuego.OptionOverrideDescription("使用 Passkey 开始登录流程，返回 WebAuthn 挑战"),
		fuego.OptionTags("Passkey 认证"),
	)
	fuego.Post(public, "/passkey/login/finish", authHandler.FinishLogin,
		fuego.OptionSummary("完成登录"),
		fuego.OptionOverrideDescription("验证 Passkey 并完成登录，返回 JWT"),
		fuego.OptionTags("Passkey 认证"),
	)

	// ========== 用户路由（需认证）==========
	userGroup := fuego.Group(server, "/api/user")
	fuego.Use(userGroup, middleware.AuthMiddleware(cfg.JWT.Secret))

	// [个人信息] 用户资料
	fuego.Get(userGroup, "/profile", userHandler.GetProfile,
		fuego.OptionSummary("获取个人信息"),
		fuego.OptionOverrideDescription("获取当前登录用户的个人信息"),
		fuego.OptionTags("个人信息"),
	)
	fuego.Put(userGroup, "/profile", userHandler.UpdateProfile,
		fuego.OptionSummary("更新个人信息"),
		fuego.OptionOverrideDescription("更新当前登录用户的个人信息"),
		fuego.OptionTags("个人信息"),
	)
	fuego.Get(userGroup, "/passkeys", userHandler.GetPasskeys,
		fuego.OptionSummary("获取 Passkey 列表"),
		fuego.OptionOverrideDescription("获取当前用户的所有 Passkey 凭证"),
		fuego.OptionTags("个人信息"),
	)
	fuego.Delete(userGroup, "/passkeys/{id}", userHandler.DeletePasskey,
		fuego.OptionSummary("删除 Passkey"),
		fuego.OptionOverrideDescription("删除指定的 Passkey 凭证"),
		fuego.OptionTags("个人信息"),
	)

	// [诗歌浏览] 诗歌相关
	fuego.Get(userGroup, "/poems", poemHandler.List,
		fuego.OptionSummary("获取诗歌列表"),
		fuego.OptionOverrideDescription("分页获取诗歌列表"),
		fuego.OptionTags("诗歌浏览"),
	)
	fuego.Get(userGroup, "/poems/{id}", poemHandler.GetByID,
		fuego.OptionSummary("获取诗歌详情"),
		fuego.OptionOverrideDescription("根据 ID 获取诗歌详情"),
		fuego.OptionTags("诗歌浏览"),
	)
	fuego.Get(userGroup, "/poems/search", poemHandler.Search,
		fuego.OptionSummary("搜索诗歌"),
		fuego.OptionOverrideDescription("按标题、作者、内容搜索诗歌"),
		fuego.OptionTags("诗歌浏览"),
	)
	fuego.Get(userGroup, "/poems/daily", poemHandler.GetDaily,
		fuego.OptionSummary("每日推荐"),
		fuego.OptionOverrideDescription("获取每日推荐的诗歌"),
		fuego.OptionTags("诗歌浏览"),
	)

	// [收藏管理] 收藏相关
	fuego.Post(userGroup, "/favorites", favoriteHandler.AddFavorite,
		fuego.OptionSummary("添加收藏"),
		fuego.OptionOverrideDescription("将诗歌添加到收藏列表"),
		fuego.OptionTags("收藏管理"),
	)
	fuego.Delete(userGroup, "/favorites/{poem_id}", favoriteHandler.RemoveFavorite,
		fuego.OptionSummary("取消收藏"),
		fuego.OptionOverrideDescription("从收藏列表中移除诗歌"),
		fuego.OptionTags("收藏管理"),
	)
	fuego.Get(userGroup, "/favorites", favoriteHandler.ListFavorites,
		fuego.OptionSummary("获取收藏列表"),
		fuego.OptionOverrideDescription("获取当前用户的收藏列表"),
		fuego.OptionTags("收藏管理"),
	)

	// [阅读计划] 阅读计划相关
	fuego.Post(userGroup, "/reading-plans", readingPlanHandler.CreatePlan,
		fuego.OptionSummary("创建阅读计划"),
		fuego.OptionOverrideDescription("创建新的阅读计划"),
		fuego.OptionTags("阅读计划"),
	)
	fuego.Get(userGroup, "/reading-plans/current", readingPlanHandler.GetCurrentPlan,
		fuego.OptionSummary("获取当前计划"),
		fuego.OptionOverrideDescription("获取当前进行中的阅读计划"),
		fuego.OptionTags("阅读计划"),
	)
	fuego.Put(userGroup, "/reading-plans/{id}/pause", readingPlanHandler.PausePlan,
		fuego.OptionSummary("暂停计划"),
		fuego.OptionOverrideDescription("暂停阅读计划"),
		fuego.OptionTags("阅读计划"),
	)
	fuego.Put(userGroup, "/reading-plans/{id}/resume", readingPlanHandler.ResumePlan,
		fuego.OptionSummary("恢复计划"),
		fuego.OptionOverrideDescription("恢复已暂停的阅读计划"),
		fuego.OptionTags("阅读计划"),
	)
	fuego.Get(userGroup, "/reading-plans/{id}/progress", readingPlanHandler.GetPlanProgress,
		fuego.OptionSummary("获取计划进度"),
		fuego.OptionOverrideDescription("获取阅读计划的进度详情"),
		fuego.OptionTags("阅读计划"),
	)
	fuego.Post(userGroup, "/reading-plans/log", readingPlanHandler.LogReading,
		fuego.OptionSummary("记录阅读"),
		fuego.OptionOverrideDescription("记录今日阅读的诗歌"),
		fuego.OptionTags("阅读计划"),
	)

	// [打卡系统] 打卡相关
	fuego.Post(userGroup, "/checkins", checkinHandler.Checkin,
		fuego.OptionSummary("打卡"),
		fuego.OptionOverrideDescription("今日打卡"),
		fuego.OptionTags("打卡系统"),
	)
	fuego.Get(userGroup, "/checkins", checkinHandler.GetCheckinList,
		fuego.OptionSummary("获取打卡记录"),
		fuego.OptionOverrideDescription("获取打卡历史记录"),
		fuego.OptionTags("打卡系统"),
	)
	fuego.Get(userGroup, "/checkins/stats", checkinHandler.GetStats,
		fuego.OptionSummary("获取打卡统计"),
		fuego.OptionOverrideDescription("获取打卡统计数据"),
		fuego.OptionTags("打卡系统"),
	)
	fuego.Get(userGroup, "/checkins/calendar", checkinHandler.GetCalendar,
		fuego.OptionSummary("获取打卡日历"),
		fuego.OptionOverrideDescription("获取月度打卡日历"),
		fuego.OptionTags("打卡系统"),
	)
	fuego.Get(userGroup, "/checkins/ranking", checkinHandler.GetRanking,
		fuego.OptionSummary("获取排行榜"),
		fuego.OptionOverrideDescription("获取打卡排行榜"),
		fuego.OptionTags("打卡系统"),
	)
}

// SetupAdminRoutes 注册后台管理路由 - 端口 8081
func SetupAdminRoutes(server *fuego.Server, db *pgxpool.Pool, cfg *config.Config) {
	// 初始化依赖
	userRepo := repository.NewUserRepository(db)
	poemRepo := repository.NewPoemRepository(db)
	adminAuthService := adminservice.NewAdminAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHour)

	// 初始化 Handler
	authHandler := admin.NewAuthHandler(adminAuthService)
	poemHandler := admin.NewPoemHandler(poemRepo)
	categoryHandler := admin.NewCategoryHandler()
	tagHandler := admin.NewTagHandler()
	statsHandler := admin.NewStatsHandler()
	bannerHandler := admin.NewBannerHandler()
	announcementHandler := admin.NewAnnouncementHandler()
	configHandler := admin.NewConfigHandler()

	// ========== 公开路由：登录 ==========
	public := fuego.Group(server, "/api")

	fuego.Post(public, "/auth/login", authHandler.Login,
		fuego.OptionSummary("管理员登录"),
		fuego.OptionOverrideDescription("后台管理员使用邮箱和密码登录，返回 JWT"),
		fuego.OptionTags("后台认证"),
	)

	// ========== 需认证路由 ==========
	adminGroup := fuego.Group(server, "/api")
	fuego.Use(adminGroup, middleware.AdminAuthMiddleware(cfg.JWT.Secret))

	// [后台认证] 用户信息
	fuego.Get(adminGroup, "/user/info", authHandler.GetUserInfo,
		fuego.OptionSummary("获取用户信息"),
		fuego.OptionOverrideDescription("获取当前登录管理员的个人信息"),
		fuego.OptionTags("后台认证"),
	)

	// [后台认证] 权限码
	fuego.Get(adminGroup, "/auth/codes", authHandler.GetAccessCodes,
		fuego.OptionSummary("获取权限码"),
		fuego.OptionOverrideDescription("获取当前管理员的权限码列表"),
		fuego.OptionTags("后台认证"),
	)

	// [后台认证] 退出登录
	fuego.Post(adminGroup, "/auth/logout", authHandler.Logout,
		fuego.OptionSummary("退出登录"),
		fuego.OptionOverrideDescription("管理员退出登录"),
		fuego.OptionTags("后台认证"),
	)

	// ========== 后台管理路由 ==========
	adminMgmt := fuego.Group(server, "/api/admin")
	fuego.Use(adminMgmt, middleware.AdminAuthMiddleware(cfg.JWT.Secret))

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
}
