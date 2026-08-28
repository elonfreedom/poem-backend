# TODO 需求文档

> 待开发功能列表，按优先级排序。

---

## 📋 需求列表

### 1. 简繁体转换（繁体转简体）

**优先级**：中  
**状态**：✅ 已完成  
**提出日期**：2026-08-28  
**完成日期**：2026-08-28

#### 背景
后台管理平台中的诗文大部分为繁体中文，需要提供简体版本供用户切换阅读。

#### 技术方案
- **库选型**：`github.com/liuzl/gocc`（纯 Go 实现，无 C 依赖，Docker 部署友好）
- **方案**：保留繁体原文（文化价值），新增简体字段供前端切换

#### 数据库变更
```sql
ALTER TABLE poems ADD COLUMN IF NOT EXISTS title_sc TEXT DEFAULT '';
ALTER TABLE poems ADD COLUMN IF NOT EXISTS content_sc TEXT DEFAULT '';
ALTER TABLE poems ADD COLUMN IF NOT EXISTS translation_sc TEXT DEFAULT '';
ALTER TABLE poems ADD COLUMN IF NOT EXISTS appreciation_sc TEXT DEFAULT '';
```

#### 实现步骤
1. 引入 gocc 库，封装 `pkg/convert/convert.go`
2. 数据库迁移新增简体字段
3. Model 层新增 `TitleSC`、`ContentSC`、`TranslationSC`、`AppreciationSC`
4. Service 层：创建/更新诗歌时自动生成简体
5. 批量转换脚本：`cmd/tools/convert-t2s/main.go`，处理已有数据
6. API 响应返回简体字段

#### 前端交互
- 诗文卡片加「繁体/简体」切换按钮
- 默认显示繁体，用户可切换到简体

#### 注意事项
- 转换后建议人工抽检名篇（多音字/通假字可能不准）
- 批量脚本支持 `--dry-run` 预览和分批处理

---

### 2. 诗歌拼音标注

**优先级**：高  
**状态**：🔧 实施中  
**提出日期**：2026-08-28

#### 背景
用户学习古诗时需要查看拼音和声调，尤其是生僻字。

#### 技术方案
- **库选型**：`github.com/mozillazg/go-pinyin`（成熟稳定，支持声调）
- **方案**：预存拼音到 DB，支持 admin 手动校正（解决多音字问题）

#### 数据库变更
```sql
ALTER TABLE poems ADD COLUMN IF NOT EXISTS title_pinyin TEXT DEFAULT '';
ALTER TABLE poems ADD COLUMN IF NOT EXISTS content_pinyin TEXT DEFAULT '';
ALTER TABLE poems ADD COLUMN IF NOT EXISTS author_pinyin TEXT DEFAULT '';
```

#### 实现步骤
1. 引入 go-pinyin 库，封装 `pkg/pinyin/pinyin.go`
2. 数据库迁移新增拼音字段
3. Model 层新增拼音字段（可编辑）
4. Service 层：创建/更新诗歌时自动生成拼音
5. Admin API：支持手动编辑拼音字段（校正多音字）
6. User API：返回拼音字段
7. 批量生成已有数据拼音的脚本

#### 前端展示
- 拼音放在对应文字**上方**
- 默认关闭，加「显示拼音」开关
- 符号声调（jìng），字号 0.75rem，灰色

#### 多音字处理
- 自动生成后，admin 可在管理后台编辑校正
- 拼音字段独立存储，不覆盖自动生成内容

---

## 📅 已完成

### 1. 简繁体转换（繁体转简体）
- **完成日期**：2026-08-28
- **数据库**：`poems` 表新增 `title_sc`、`content_sc` 字段（迁移 018）
- **转换库**：`github.com/liuzl/gocc`（纯 Go，无 C 依赖）
- **逻辑**：创建/更新诗歌时，若简体字段为空则从繁体自动生成；双向录入时以用户输入为准
- **批量脚本**：`cmd/tools/convert-t2s/main.go`（支持 `--dry-run`、`--sql-only`）
- **API**：admin 和 user 接口均返回 `title_sc`、`content_sc`
- **调整**：作者不参与简繁转换，不生成拼音；拼音仅保留标题和正文
