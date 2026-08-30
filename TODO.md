# 后端待办事项

> 最后更新：2026-08-29
> 规范依据：`design/architecture.md` + `design/business.md`

---

## P0 — 高优先级

### #12 Passkey 绑定延迟修复：长轮询替代短轮询
- **状态**: ✅ 已完成 (2026-08-29)
- **问题**: AddDeviceStatus（设备A）和 AddDeviceStatusPublic（设备B）使用短轮询，前端定时请求造成延迟
- **方案**: 改为长轮询，服务端持有连接直到状态变化或 30s 超时
- **涉及文件**:
  - `internal/handler/user/auth_handler.go`
  - `internal/handler/user/connection_store.go`
- **改动**: ConnectionStore 增加 pub/sub 机制，Subscribe/notifySubscribers；Update 状态变化时通知订阅者；两个 status handler 改为等待 channel 或超时

### #13 修复心跳清理间隔配置错误
- **状态**: ✅ 已完成 (2026-08-29)
- **问题**: `connection_store.go` 定义 `heartbeatCleanupInterval=20s`，但 `router.go` 实际传入 5 分钟
- **方案**: 修复为使用 20s 常量
- **涉及文件**:
  - `internal/router/user_router.go`

### #14 打卡需校验活跃阅读计划
- **状态**: ✅ 已完成 (2026-08-29)
- **问题**: `CheckinService.Checkin` 未校验用户有活跃阅读计划，任何用户都能打卡
- **方案**: 打卡前调用 `readingPlanRepo.GetActiveByUserID` 检查，无活跃计划返回 BadRequestError
- **规范**: `business.md` §7.1 — "打卡需关联阅读计划，无活跃计划不可打卡"
- **涉及文件**:
  - `internal/service/user/checkin_service.go`
  - `internal/router/user_router.go`

### #15 阅读计划 completed 状态自动转换
- **状态**: ✅ 已完成 (2026-08-29)
- **问题**: 状态机 `active → completed` 转换缺失，计划永不标记完成
- **方案**: 新增 `CheckAndCompletePlan` 方法，在 `LogReading` 后自动检测：所有天数达标或已过结束日期则标记 completed
- **规范**: `business.md` §6.1 状态机定义
- **涉及文件**:
  - `internal/service/user/reading_plan_service.go`

---

## P1 — 中优先级

### #16 简繁体转换扩展 translation_sc / appreciation_sc
- **状态**: ✅ 已完成 (2026-08-29)
- **问题**: Create/Update 和批量工具只转换 title/author/content，缺 translation_sc、appreciation_sc
- **方案**: Model + Request + Response + Repository + Service 全链路扩展两个字段
- **规范**: `design/architecture.md` §3.2
- **涉及文件**:
  - `internal/model/poem.go`
  - `internal/model/admin/admin.go`
  - `internal/model/user/poem.go`
  - `internal/repository/poem_repo.go`
  - `internal/repository/pinyin_init.go`
  - `internal/service/admin/admin_poem_service.go`
  - `internal/service/user/poem_service.go`
  - `migrations/022_add_translation_sc_appreciation_sc.up.sql`

### #17 单首创建诗歌时校验唯一性
- **状态**: ✅ 已完成 (2026-08-29)
- **问题**: `AdminPoemService.Create` 未校验标题+作者+正文首句唯一性，仅 ImportPoems 校验
- **方案**: 在 Create 方法中调用 `ExistsByTitleAuthorFirstLine` 检查，重复返回 ConflictError
- **涉及文件**:
  - `internal/service/admin/admin_poem_service.go`
- **涉及文件**:
  - `internal/service/admin/admin_poem_service.go`

### #18 订阅打卡幂等性防护
- **状态**: ✅ 已完成 (已存在)
- **问题**: `CreateCheckin` 无幂等防护
- **方案**: 已存在 — `checkin_repo.go` 有 `ON CONFLICT (user_id, date) DO NOTHING`，`shared_plan_repo.go` 有 `ON CONFLICT (subscription_id, day_number) DO UPDATE SET`
- **涉及文件**: 无需修改

---

## P2 — 低优先级

### #19 收藏列表已下架诗歌标记
- **状态**: ✅ 已完成 (2026-08-29)
- **问题**: `ListFavorites` 静默跳过被删除诗歌，前端无法区分"无收藏"和"已下架"
- **方案**: `FavoriteResponse` 增加 `available` 字段，已删除诗歌返回 `available: false`
- **涉及文件**:
  - `internal/model/user/favorite.go`
  - `internal/service/user/favorite_service.go`

---

## 已完成

- ✅ #12 Passkey 长轮询替代短轮询 — 2026-08-29
- ✅ #13 修复心跳清理间隔（5min→20s）— 2026-08-29
- ✅ #14 打卡校验活跃阅读计划 — 2026-08-29
- ✅ #15 阅读计划 completed 状态自动转换 — 2026-08-29
- ✅ #16 简繁体转换扩展 translation_sc/appreciation_sc — 2026-08-29
- ✅ #17 单首创建诗歌唯一性校验 — 2026-08-29
- ✅ #18 订阅打卡幂等性防护（已存在）— 2026-08-29
- ✅ #19 收藏列表已下架诗歌标记 — 2026-08-29
