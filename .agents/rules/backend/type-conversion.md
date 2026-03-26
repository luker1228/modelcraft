---
paths:
  - "internal/**/*.go"
  - "pkg/**/*.go"
---

# 类型转换规范

进行类型转换时，必须使用 `github.com/spf13/cast` 包，禁止使用原生类型断言。

Refer to @ai-metadata/backend/development/type-conversion.md for cast usage patterns, ToXxxE vs ToXxx selection criteria, and available conversion functions.

## 触发场景

- 从 `map[string]interface{}` 或 `interface{}` 中提取值时
- 处理来自外部系统（配置、JSON、GraphQL 输入）的动态类型时
- 需要安全地将任意类型转换为具体类型时
