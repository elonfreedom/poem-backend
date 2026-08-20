# 开发流程

## 新功能开发
1. 从 `develop` 创建 `feature/*` 分支
2. 编写代码，遵循代码规范
3. 添加测试用例
4. 更新相关文档
5. 提交 PR 到 `develop`
6. Code Review 通过后合并

## Bug 修复
1. 从 `main` 或 `develop` 创建 `fix/*` 分支
2. 编写测试复现 bug
3. 修复问题
4. 验证修复
5. 提交 PR

## 测试规范
- 单元测试覆盖率目标：80%
- 测试文件与源文件同目录，`_test.go` 后缀
- 使用表驱动测试（Table-Driven Tests）
- Mock 外部依赖
