---
paths:
  - "internal/infrastructure/dbgen/**/*.go"
  - "internal/infrastructure/**/*.go"
---

# sqlc 自定义类型规范

实现 sqlc 自定义类型时，必须遵守 `sql.Scanner`/`driver.Valuer` 接口实现标准。

Refer to @ai-metadata/backend/development/sqlc-custom-types.md for custom type implementation templates, db tag requirements, Scan method patterns, and common sqlc pitfalls.

## 触发场景

- 在 `internal/infrastructure/dbgen/` 中新增自定义类型时
- 处理 JSON 字段（StringSlice、JSONMap 等）时
- 遇到 sqlc 类型扫描异常时
