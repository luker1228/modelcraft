---
paths:
  - "internal/infrastructure/**/*.go"
---

# Repository 层开发规范

开发或修改 `internal/infrastructure/**/*.go` 时，必须遵守 Repository 层开发规范。

Refer to @ai-metadata/backend/development/repo-develop.md for RecordNotFound handling patterns (Model A vs Model B), error wrapping with helper functions, querier interface usage, Null type conversion helpers, and transaction rules.

## 触发场景

- 开发或修改 `internal/infrastructure/` 目录下任何文件时
- 实现仓储接口时
- 处理 SQL 查询结果和错误时
- 使用 sqlc 生成代码时
