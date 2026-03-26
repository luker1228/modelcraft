---
paths:
  - "internal/**/*.go"
---

# 日志规范

记录日志时，必须使用 `pkg/logfacade`，遵守字段 key 常量和堆栈跟踪使用规则。

Refer to @ai-metadata/backend/development/logging.md for log field key constants, Stack() usage rules per layer, structured error logging patterns, and complex object logging.

## 触发场景

- 在任何层打印日志时
- 在 Interfaces 层转换错误时（必须同时记录 `Err()` 和 `Stack()`）
- 记录 SQL、HTTP 请求等结构化信息时
