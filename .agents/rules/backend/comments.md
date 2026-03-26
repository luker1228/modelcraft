---
paths:
  - "internal/**/*.go"
  - "pkg/**/*.go"
---

# 注释规范

编写 Go 代码时，必须为所有导出标识符添加符合规范的文档注释。

Refer to @ai-metadata/backend/development/comments.md for exported identifier comment requirements, comment style guidelines, and prohibited patterns (trailing comments, missing docs, etc.).

## 触发场景

- 创建或修改导出的类型、函数、方法、常量、变量时
- 编写包级别的 package 注释时
- 编写接口方法注释时
