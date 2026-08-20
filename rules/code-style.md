# 代码规范

## 命名约定
- **包名**: 小写单词，如 `config`, `middleware`, `repository`
- **结构体**: PascalCase，如 `PoemResponse`, `CreatePoemRequest`
- **函数/方法**: PascalCase（导出）或 camelCase（未导出）
- **常量**: PascalCase 或 camelCase，如 `UserIDKey`
- **JSON 字段**: snake_case，如 `created_at`, `user_id`
- **DB 字段**: snake_case，与数据库列名一致

## 错误处理
- 使用 `fmt.Errorf` 包装错误，保留原始错误链
- API 错误通过 `pkg/response` 统一返回
- 不要忽略错误，除非有明确注释说明

## 代码组织
- 每个文件不超过 300 行，超过则拆分
- 导入顺序：标准库 → 第三方包 → 项目内部包，用空行分隔
- 避免循环依赖，单向依赖：handler → service → repository
