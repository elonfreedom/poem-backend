package router

import (
	"context"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgxpool"

	"poem-backend/internal/config"
	"poem-backend/internal/handler/user"
	"poem-backend/internal/middleware"
	"poem-backend/internal/repository"
	userservice "poem-backend/internal/service/user"
)

// initUserDependencies 初始化用户端依赖（Repository、Service、Handler）
func initUserDependencies(server *fuego.Server, db *pgxpool.Pool, cfg *config.Config) (
	*user.AuthHandler,
	*user.UserHandler,
	*user.PoemHandler,
	*user.FavoriteHandler,
	*user.ReadingPlanHandler,
	*user.SharedPlanHandler,
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
	sessionRepo := repository.NewSessionRepository(db)
	connectionRepo := repository.NewConnectionRepository(db)
	poemRepo := repository.NewPoemRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	readingPlanRepo := repository.NewReadingPlanRepository(db)
	checkinRepo := repository.NewCheckinRepository(db)
	sharedPlanRepo := repository.NewSharedPlanRepository(db)

	// 初始化 Service
	authService := userservice.NewAuthService(userRepo, passkeyRepo, wn, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	userService := userservice.NewUserService(userRepo, passkeyRepo)
	poemService := userservice.NewPoemService(poemRepo)
	favoriteService := userservice.NewFavoriteService(favoriteRepo, poemRepo)
	readingPlanService := userservice.NewReadingPlanService(readingPlanRepo)
	sharedPlanService := userservice.NewSharedPlanService(sharedPlanRepo, poemRepo)
	checkinService := userservice.NewCheckinService(checkinRepo)

	// 初始化 Handler
	authHandler := user.NewAuthHandler(authService, sessionRepo, connectionRepo)
	userHandler := user.NewUserHandler(userService)
	poemHandler := user.NewPoemHandler(poemService)
	favoriteHandler := user.NewFavoriteHandler(favoriteService)
	readingPlanHandler := user.NewReadingPlanHandler(readingPlanService)
	sharedPlanHandler := user.NewSharedPlanHandler(sharedPlanService)
	checkinHandler := user.NewCheckinHandler(checkinService)

	return authHandler, userHandler, poemHandler, favoriteHandler, readingPlanHandler, sharedPlanHandler, checkinHandler
}

// SetupUserRoutes 注册用户端（C端）路由 - 端口 8080
func SetupUserRoutes(server *fuego.Server, db *pgxpool.Pool, cfg *config.Config) {
	authHandler, userHandler, poemHandler, favoriteHandler, readingPlanHandler, sharedPlanHandler, checkinHandler :=
		initUserDependencies(server, db, cfg)

	// 启动定期清理过期 session（每 5 分钟）
	authHandler.SessionStore.StartCleanup(context.Background(), 5*time.Minute)
	authHandler.ConnectionStore.StartCleanup(context.Background(), 5*time.Minute)

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

	// 公开：每日推荐诗歌（无需认证）
	fuego.Get(public, "/poems/daily", poemHandler.GetDaily,
		fuego.OptionSummary("每日推荐"),
		fuego.OptionOverrideDescription("获取每日推荐诗歌（公开，无需认证）"),
		fuego.OptionTags("诗歌浏览"),
	)

	// 公开：共享计划库浏览（无需认证）
	fuego.Get(public, "/shared-plans", sharedPlanHandler.ListSharedPlans,
		fuego.OptionSummary("浏览共享库"),
		fuego.OptionOverrideDescription("分页浏览共享计划库（公开，无需认证）"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Get(public, "/shared-plans/{id}", sharedPlanHandler.GetSharedPlan,
		fuego.OptionSummary("计划详情"),
		fuego.OptionOverrideDescription("获取共享计划详情（公开，无需认证）"),
		fuego.OptionTags("共享计划"),
	)

	// 公开：跨设备 Passkey（设备 B 无 token）
	fuego.Post(public, "/passkeys/add/connect", authHandler.AddDeviceConnect,
		fuego.OptionSummary("新设备连接"),
		fuego.OptionOverrideDescription("设备 B 扫码后连接，上报设备名称（公开接口）"),
		fuego.OptionTags("Passkey 认证"),
	)
	fuego.Get(public, "/passkeys/add/status", authHandler.AddDeviceStatusPublic,
		fuego.OptionSummary("查询连接状态"),
		fuego.OptionOverrideDescription("设备 B 公开轮询连接状态，无需认证（公开接口）"),
		fuego.OptionTags("Passkey 认证"),
	)
	fuego.Post(public, "/passkeys/add/finish", authHandler.AddDeviceFinish,
		fuego.OptionSummary("完成设备注册"),
		fuego.OptionOverrideDescription("设备 B 完成 Passkey 注册，获取 JWT（公开接口）"),
		fuego.OptionTags("Passkey 认证"),
	)
	fuego.Post(public, "/passkeys/add/reject", authHandler.AddDeviceReject,
		fuego.OptionSummary("设备 B 放弃绑定"),
		fuego.OptionOverrideDescription("设备 B 主动放弃创建 Passkey，通知设备 A 关闭等待弹窗（公开接口）"),
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

	// [跨设备 Passkey] 添加新设备
	fuego.Post(userGroup, "/passkeys/add/begin", authHandler.AddDeviceBegin,
		fuego.OptionSummary("开始添加设备"),
		fuego.OptionOverrideDescription("设备 A 发起添加新设备，生成连接令牌和 WebAuthn 注册选项"),
		fuego.OptionTags("Passkey 认证"),
	)
	fuego.Get(userGroup, "/passkeys/add/status", authHandler.AddDeviceStatus,
		fuego.OptionSummary("查询连接状态"),
		fuego.OptionOverrideDescription("设备 A 轮询连接状态，确认设备 B 已连接"),
		fuego.OptionTags("Passkey 认证"),
	)
	fuego.Post(userGroup, "/passkeys/add/confirm", authHandler.AddDeviceConfirm,
		fuego.OptionSummary("确认授权"),
		fuego.OptionOverrideDescription("设备 A 确认或拒绝新设备授权"),
		fuego.OptionTags("Passkey 认证"),
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

	// [共享计划] 计划共享库 — 管理我创建的计划
	fuego.Post(userGroup, "/shared-plans", sharedPlanHandler.CreateSharedPlan,
		fuego.OptionSummary("创建共享计划"),
		fuego.OptionOverrideDescription("创建并发布计划到共享库"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Get(userGroup, "/shared-plans/mine", sharedPlanHandler.GetMySharedPlans,
		fuego.OptionSummary("我的计划"),
		fuego.OptionOverrideDescription("获取我创建的共享计划列表"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Put(userGroup, "/shared-plans/{id}", sharedPlanHandler.UpdateSharedPlan,
		fuego.OptionSummary("更新计划"),
		fuego.OptionOverrideDescription("更新共享计划"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Put(userGroup, "/shared-plans/{id}/publish", sharedPlanHandler.PublishPlan,
		fuego.OptionSummary("发布计划"),
		fuego.OptionOverrideDescription("重新发布计划"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Put(userGroup, "/shared-plans/{id}/unpublish", sharedPlanHandler.UnpublishPlan,
		fuego.OptionSummary("取消发布"),
		fuego.OptionOverrideDescription("从共享库下架计划"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Delete(userGroup, "/shared-plans/{id}", sharedPlanHandler.DeleteSharedPlan,
		fuego.OptionSummary("删除计划"),
		fuego.OptionOverrideDescription("删除共享计划"),
		fuego.OptionTags("共享计划"),
	)

	// [共享计划] 订阅管理
	fuego.Post(userGroup, "/shared-plans/{id}/subscribe", sharedPlanHandler.Subscribe,
		fuego.OptionSummary("订阅计划"),
		fuego.OptionOverrideDescription("订阅共享计划"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Delete(userGroup, "/shared-plans/{id}/subscribe", sharedPlanHandler.Unsubscribe,
		fuego.OptionSummary("取消订阅"),
		fuego.OptionOverrideDescription("取消订阅共享计划"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Put(userGroup, "/subscriptions/{id}/start-date", sharedPlanHandler.SetStartDate,
		fuego.OptionSummary("设置开始日期"),
		fuego.OptionOverrideDescription("设置订阅计划的开始日期"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Get(userGroup, "/subscriptions", sharedPlanHandler.GetMySubscriptions,
		fuego.OptionSummary("我的订阅"),
		fuego.OptionOverrideDescription("获取我的订阅列表"),
		fuego.OptionTags("共享计划"),
	)

	// [共享计划] 每日诗文 & 打卡
	fuego.Get(userGroup, "/subscriptions/{id}/today", sharedPlanHandler.GetTodayPoem,
		fuego.OptionSummary("今日诗文"),
		fuego.OptionOverrideDescription("获取今日需要阅读的诗文"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Post(userGroup, "/subscriptions/{id}/checkin", sharedPlanHandler.Checkin,
		fuego.OptionSummary("打卡"),
		fuego.OptionOverrideDescription("今日打卡"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Post(userGroup, "/subscriptions/{id}/skip", sharedPlanHandler.SkipDay,
		fuego.OptionSummary("跳过天数"),
		fuego.OptionOverrideDescription("跳过当前天，返回下一首未打卡的诗文"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Get(userGroup, "/subscriptions/{id}/progress", sharedPlanHandler.GetSubscriptionProgress,
		fuego.OptionSummary("订阅进度"),
		fuego.OptionOverrideDescription("获取订阅计划的打卡进度"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Get(userGroup, "/subscriptions/{id}/checkins", sharedPlanHandler.GetCheckins,
		fuego.OptionSummary("打卡记录"),
		fuego.OptionOverrideDescription("获取订阅计划的所有打卡记录（用于热力图）"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Put(userGroup, "/subscriptions/{id}/pause", sharedPlanHandler.PauseSubscription,
		fuego.OptionSummary("暂停订阅"),
		fuego.OptionOverrideDescription("暂停订阅计划"),
		fuego.OptionTags("共享计划"),
	)
	fuego.Put(userGroup, "/subscriptions/{id}/resume", sharedPlanHandler.ResumeSubscription,
		fuego.OptionSummary("恢复订阅"),
		fuego.OptionOverrideDescription("恢复订阅计划"),
		fuego.OptionTags("共享计划"),
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
