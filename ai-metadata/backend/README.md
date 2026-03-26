# AI Metadata - Backend Knowledge Base

本目录包含 ModelCraft Go 后端的核心知识文档，供 AI 助手和开发者参考。

## 目录结构

```
ai-metadata/backend/
├── README.md       # 本文件 - 知识库总览
├── design/         # 设计理念 (最高优先级)
├── development/    # 开发规范
├── testing/        # 测试策略
├── deployment/     # 部署指南
└── tools/          # 工具手册
```

## 优先级说明

当文档之间存在冲突时，按以下优先级处理：

```
设计理念 > 开发规范 > 测试策略 > 部署指南 > 工具手册
```

1. **设计理念** - 不可动摇的核心原则，所有决策的基础
2. **开发规范** - 基于设计理念的实现指导
3. **测试策略** - 验证实现是否符合设计
4. **部署指南** - 将实现交付到生产环境
5. **工具手册** - 支持以上所有环节的辅助工具

## AI 助手使用指南

### 开始任务前

先阅读 @./design/README.md 了解核心约束，特别是领域模型和业务规则，避免在错误方向上实现。

当任务没有明确说明设计需求时，必须以 @./design/ 目录下的文档作为设计依据，不得自行假设或偏离已有的领域模型、业务规则和核心原则。

### 编写代码时

参考 @./development/README.md，重点关注：

- @./development/architecture.md — DDD 分层规则，明确各层依赖方向和禁止跨层
- @./development/code-style.md — 命名约定、事务模式、协程使用
- @./development/error-handling.md — 错误包体系、各层错误职责、RecordNotFound 处理
- @./development/repo-develop.md — Repository 层开发规范，RecordNotFound 两种模式的选择

### 修改数据库查询时

参考 @./development/sqlc-custom-types.md 了解 sqlc 自定义类型的使用方式，避免手动处理类型转换。

### 编写或调试日志时

参考 @./development/logging.md，确保使用 `logfacade` 而非标准库，以及 `logfacade.Stack()` 仅在接口层使用。

### 编写测试时

参考 @./testing/README.md 了解测试策略，以及 @./testing/debugging-workflow.md 了解调试流程。

### 部署相关时

参考 @./deployment/README.md 了解环境配置和部署流程。

### 使用构建工具时

参考 @./tools/justfile-guide.md 了解所有 `just` 命令的用途和参数，以及 @./tools/tools-installation.md 了解工具安装方式。

## 文档维护原则

- `design` 对应 `internal/domain/`，谨慎更新，变更需确认
- `development` 对应 `internal/` 架构，跟随代码变化更新
- `testing` 对应 `tests/`、`*_test.go`，跟随测试变化更新
- `deployment` 对应 `docker-compose*.yml`，跟随配置变化更新
- `tools` 对应 `justfile`、`scripts/`，跟随工具版本更新
