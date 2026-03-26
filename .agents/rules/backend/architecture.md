---
paths:
  - "internal/**/*.go"
---

# 架构分层规范

开发任何 `internal/` 目录下的代码时，必须遵守 DDD 分层依赖规则。

Refer to @ai-metadata/backend/development/architecture.md for detailed layer responsibilities, dependency rules, directory mappings, and code examples.

## 触发场景

- 新增或修改 `internal/` 目录下任何文件时
- 添加跨层 import 语句时
- 设计新模块、服务或仓储时
