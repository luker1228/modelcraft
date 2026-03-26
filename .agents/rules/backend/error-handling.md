---
paths:
  - "internal/**/*.go"
---

# 错误处理规范

跨层传递错误时，必须遵守各层错误职责划分，正确使用 `bizerrors` 和 `RepositoryError`。

Refer to @ai-metadata/backend/development/error-handling.md for error code definitions, per-layer error responsibilities, RecordNotFound handling rules, and logging/stack trace guidelines.

## 触发场景

- 在 Repository 层处理数据库错误时
- 在 Application 层转换错误语义时
- 在 Interfaces 层将错误转换为 GraphQL 响应时
- 判断记录是否存在（RecordNotFound）时
