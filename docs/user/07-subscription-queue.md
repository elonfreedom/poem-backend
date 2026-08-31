# 订阅计划排队机制

## 功能概述

订阅计划排队机制实现「同一时间段只能有一个活跃阅读计划」的核心规则。用户订阅多个计划时，第一个计划为当前计划（`is_current=true`），后续计划按订阅时间排队等待。当前计划完成后，用户手动确认激活下一个计划。

## 设计哲学

> **打卡必须先有计划**。计划是打卡的前提。
>
> 自己做计划麻烦（选诗难、排序难）→ 有了「诗林」订阅他人计划
> 用户可能同时订阅多个计划 → 但每天只能打卡一次 → 所以有了「排队」

## 用户故事

- 作为用户，我想订阅多个阅读计划，但同一时间只有一个计划在进行
- 作为用户，我想知道我的排队计划什么时候能开始
- 作为用户，当前计划完成后，我想选择是否开始下一个计划
- 作为用户，我不想重复阅读已经读过的诗文

## 状态模型

### 订阅状态

| 状态 | 含义 | 触发条件 |
|------|------|----------|
| `subscribed` | 已订阅，排队中或进行中 | 订阅时 |
| `completed` | 已完成 | 全部诗文打卡完毕 |
| `cancelled` | 已取消 | 用户主动取消订阅 |

### 当前计划标记

| 字段 | 类型 | 说明 |
|------|------|------|
| `is_current` | boolean | 标记当前活跃计划，每用户仅一个 true |
| `queue_order` | int | 队列排序（0=当前，1=下一个，2=再下一个...） |

### 状态迁移

```
                    ┌─────────────────────────────────┐
                    │         subscribed              │
                    │  (is_current=true, queue_order=0)│
                    └───────────┬─────────────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
                ▼               ▼               ▼
        ┌──────────┐    ┌──────────┐    ┌──────────┐
        │completed │    │cancelled │    │cancelled │
        │(全部完成) │    │(用户取消) │    │(用户取消) │
        └──────────┘    └──────────┘    └──────────┘
```

## 功能详情

### 1. 订阅计划

**接口**：`POST /api/user/shared-plans/{id}/subscribe`

**业务规则**：
1. 检查计划是否存在且已发布
2. 检查用户是否已订阅该计划（status=subscribed）
3. 查找用户当前计划（is_current=true）
   - **无当前计划** → 新订阅 `is_current=true, queue_order=0, start_date=今天`
   - **有当前计划** → 新订阅 `is_current=false, queue_order=Max+1, start_date=null`
4. 订阅数 +1
5. 返回订阅结果（含 is_current、queue_order、estimated_start_month）

**响应示例**：
```json
{
  "id": 123,
  "status": "subscribed",
  "is_current": false,
  "queue_order": 1,
  "estimated_start_month": "2026-12"
}
```

### 2. 取消订阅

**接口**：`DELETE /api/user/shared-plans/{id}/subscribe`

**业务规则**：
1. 加载订阅，校验用户归属
2. 软删除：`status → cancelled`（不删除记录）
3. 如果取消的是当前计划（is_current=true）：
   - `is_current → false`
   - 后续排队计划的 `queue_order` 前移（-1）
   - **不自动激活**下一个计划（前端弹窗让用户确认）
4. 如果取消的是排队计划（is_current=false）：
   - 该计划之后的订阅 `queue_order` 前移
5. 订阅数 -1

### 3. 激活排队计划

**接口**：`POST /api/user/subscriptions/{id}/activate`

**请求体**：
```json
{
  "start_date": "2026-12-01"
}
```

**业务规则**：
1. 加载订阅，校验 status=subscribed
2. 将现有 is_current=true 的订阅 → is_current=false
3. 目标订阅：`is_current=true, queue_order=0, start_date=请求值`
4. 重新计算其他订阅的 queue_order

### 4. 打卡

**接口**：`POST /api/user/subscriptions/{id}/checkin`

**业务规则**：
1. 创建打卡记录（现有逻辑不变）
2. 计算完成率 = 已打卡诗文数 / actual_total
3. 如果 actual_total 全部完成：
   - `status → completed, is_current → false`
   - 后续排队计划的 `queue_order` 前移
   - **不自动激活**下一个计划

### 5. 获取今日诗文（重复过滤）

**接口**：`GET /api/user/subscriptions/{id}/today`

**业务规则**：
1. 获取计划的 poem_ids
2. 查询用户全局已打卡诗文（checkin_history）
3. 过滤：`effective_poems = poem_ids - checked_in_globally`
4. 基于 start_date 计算 day_number
5. 从 effective_poems 中取当天诗文返回

**响应新增字段**：
- `actual_total`：实际需打卡数 = 总数 - 全局已读数

### 6. 获取订阅列表

**接口**：`GET /api/user/subscriptions`

**业务规则**：
1. 返回用户所有订阅（按 queue_order 排序）
2. 每个订阅包含：completed_poems、total_poems、actual_total
3. 排队中的计划包含 estimated_start_month

**响应示例**：
```json
{
  "subscriptions": [
    {
      "id": 1,
      "shared_plan_id": 10,
      "title": "唐诗三百首",
      "status": "subscribed",
      "is_current": true,
      "queue_order": 0,
      "start_date": "2026-10-01",
      "completed_poems": [1, 2, 3],
      "total_poems": 30,
      "actual_total": 25
    },
    {
      "id": 2,
      "shared_plan_id": 20,
      "title": "宋词精选",
      "status": "subscribed",
      "is_current": false,
      "queue_order": 1,
      "start_date": null,
      "completed_poems": [],
      "total_poems": 25,
      "actual_total": 20,
      "estimated_start_month": "2026-12"
    }
  ]
}
```

## 激活规则

| 事件 | 结果 |
|------|------|
| 用户订阅，无当前计划 | 新计划 → `is_current=true, queue_order=0` |
| 用户订阅，有当前计划 | 新计划 → `is_current=false, queue_order=Max+1` |
| 当前计划完成 | 该计划 → completed，后续 queue_order 前移 |
| 当前计划取消 | 该计划 → cancelled，后续 queue_order 前移 |
| 排队计划取消 | 该计划 → cancelled，后续 queue_order 前移 |
| 用户激活排队计划 | 目标 → `is_current=true, queue_order=0` |

## expected_start_month 计算规则

```
当前计划的 start_date + 当前计划的 total_days = 预计结束日期
estimated_start_month = 预计结束日期所在月份
```

格式：`"2026-12"`（YYYY-MM）

示例：
- 当前计划：唐诗（30天），start_date: 2026-10-01
- 预计结束：2026-10-31
- 排队计划 estimated_start_month：`"2026-11"`

## 重复诗文过滤

**设计原则**：不修改计划定义的 poem_ids，在获取今日诗文时实时过滤。

**逻辑**：
- `total_poems`：计划定义的诗文总数
- `completed_poems`：本计划已打卡的诗文 ID 列表
- `global_checked`：用户全局已打卡诗文（跨所有计划）
- `actual_total` = `total_poems` - `global_checked` 数量
- 获取今日诗文时，从 poem_ids 中排除 global_checked

## 边界场景

| 场景 | 处理 |
|------|------|
| 取消排队中的计划 | 该计划 → cancelled，后续计划 queue_order 前移 |
| 并发订阅（快速连续点击） | 数据库唯一约束保证只有一个 is_current=true |
| 当前计划被管理员删除 | 等同于完成，触发 queue_order 前移 |
| 用户无当前计划时获取今日诗文 | 返回错误（无活跃计划） |

## 数据模型变更

### plan_subscriptions 表

**新增字段**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `is_current` | BOOLEAN | 是否为当前计划 |
| `queue_order` | INT | 队列排序 |

**修改字段**：

| 字段 | 变更 | 说明 |
|------|------|------|
| `status` | 枚举改为：subscribed/completed/cancelled | 移除 active/pending/paused |
| `start_date` | 改为可空 | 仅当前计划有值 |

**新增约束**：

```sql
-- 每用户最多一个当前计划
CREATE UNIQUE INDEX idx_one_current_per_user 
    ON plan_subscriptions(user_id) WHERE is_current = true;

-- 每用户每计划最多一个 subscribed 订阅（允许取消后重订）
CREATE UNIQUE INDEX idx_unique_user_plan_subscribed 
    ON plan_subscriptions(user_id, shared_plan_id) WHERE status = 'subscribed';
```

### 存量数据迁移

| 现有 status | 迁移后 status | is_current |
|-------------|---------------|------------|
| active | subscribed | true |
| paused | cancelled | false |
| completed | completed | false |

## 验收标准

- [ ] 无当前计划时订阅，新计划直接激活（is_current=true, queue_order=0）
- [ ] 有当前计划时订阅，新计划进入队列（is_current=false）
- [ ] 当前计划完成/取消后，后续排队计划 queue_order 前移
- [ ] 排队计划显示正确的 estimated_start_month
- [ ] 取消排队计划不影响其他排队计划的位置
- [ ] 用户可通过 activate 接口激活排队计划
- [ ] 获取今日诗文时过滤全局已读诗文
- [ ] 并发订阅不会产生多个 is_current=true

## API 变更清单

| 方法 | 路径 | 变更 |
|------|------|------|
| `POST` | `/shared-plans/{id}/subscribe` | 响应新增 is_current、queue_order、estimated_start_month |
| `DELETE` | `/shared-plans/{id}/subscribe` | 改为软删除（status→cancelled） |
| `POST` | `/subscriptions/{id}/activate` | **新增** 激活排队计划 |
| `GET` | `/subscriptions` | 响应新增 is_current、queue_order、completed_poems、actual_total |
| `POST` | `/subscriptions/{id}/checkin` | 完成后 is_current→false，不自动激活下一个 |
| `GET` | `/subscriptions/{id}/today` | 新增重复诗文过滤逻辑 |
