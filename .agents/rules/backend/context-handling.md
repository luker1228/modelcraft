---
paths:
  - "internal/**/*.go"
---

# Context 处理规范

使用 Context 时，必须通过 `pkg/ctxutils` 提取值，Application/Domain 层使用显式参数。

Refer to @ai-metadata/backend/development/context-handling.md for ctxutils usage, per-layer context extraction rules, explicit parameter passing patterns, and available context getter functions.

## 触发场景

- 在 Interfaces 层从 context 中获取 orgName、userID 等信息时
- 设计 Application 层函数签名时
- 编写中间件或请求处理器时
