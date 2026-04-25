# 🏗️ 开发规范

> **优先级: 高** - 基于设计理念的实现指导，定义代码风格和架构分层。

## 概述

开发规范定义了如何将设计理念落地为可维护的代码。包括架构分层、代码风格、命名规范等。

## 📚 文档列表

| 文档 | 说明 | 适用场景 |
|------|------|----------|
| [architecture.md](./architecture.md) | 架构分层详解 | 理解整体架构 |
| [domain-development.md](./domain-development.md) | Domain 层开发规范 ⭐ | `internal/domain/**/*.go`，定义 Repository 接口时 |
| [repo-develop.md](./repo-develop.md) | Repository 层开发规范 ⭐ | `internal/infrastructure/**/*.go` |
| [error-handling.md](./error-handling.md) | 错误处理规范 | 跨层错误处理 |
| [logging.md](./logging.md) | 日志规范 | 日志记录 |
| [type-conversion.md](./type-conversion.md) | 类型转换规范 | 类型转换 |
| [context-handling.md](./context-handling.md) | Context 处理规范 | Context 使用 |
| [tenant-scope-and-propagation.md](./tenant-scope-and-propagation.md) | 租户隔离与参数传递规范 ⭐ | 涉及 org / org+project 隔离与参数链路时 |
| [comments.md](./comments.md) | 注释规范 | 代码注释 |
| [sqlc-custom-types.md](./sqlc-custom-types.md) | sqlc 自定义类型规范 | 自定义类型实现 |
| [contract-sync.md](./contract-sync.md) | API Contract 共享规范 ⭐ | 修改 `api/` 目录时 |
| [tx-querier-rules.md](./tx-querier-rules.md) | WithTx/Querier 事务使用规范 ⭐ | 修改 `internal/app/**/*.go` 的事务逻辑时 |

## 🏛️ 架构分层

```
┌─────────────────────────────────────────┐
│           Interfaces (接口层)            │  ← HTTP/GraphQL/gRPC 入口
├─────────────────────────────────────────┤
│          Application (应用层)            │  ← 用例编排、事务管理
├─────────────────────────────────────────┤
│            Domain (领域层)               │  ← 业务逻辑、领域模型 ⭐核心
├─────────────────────────────────────────┤
│        Infrastructure (基础设施层)        │  ← 数据库、外部服务
└─────────────────────────────────────────┘
```

### 各层职责

| 层 | 职责 | 依赖 |
|----|------|------|
| Interfaces | 请求解析、响应格式化 | Application |
| Application | 用例编排、事务控制 | Domain |
| Domain | 业务逻辑、领域规则 | 无外部依赖 |
| Infrastructure | 数据持久化、外部集成 | Domain (实现接口) |

## 📁 目录结构

```
internal/
├── interfaces/          # 接口层
│   ├── http/           # HTTP API
│   └── graphql/        # GraphQL API
├── application/        # 应用层
│   ├── service/        # 应用服务
│   └── dto/            # 数据传输对象
├── domain/             # 领域层
│   ├── model/          # 领域模型
│   ├── repository/     # 仓储接口
│   └── service/        # 领域服务
└── infrastructure/     # 基础设施层
    ├── persistence/    # 数据持久化
    └── external/       # 外部服务
```

## 🎯 Repository 层开发 (重点)

> **最重要的开发规范** - 开发 `internal/infrastructure/**/*.go` 时必读!

### 核心规则速查

**RecordNotFound 处理的两种模式**:

| 场景 | 返回值 | 不存在时 | 示例方法 |
|------|--------|----------|----------|
| 必须存在的记录 | `(value, error)` | 返回 `NotFoundError` | `GetByID`, `GetByName` |
| 可能不存在的查询 | `(value, bool, error)` | 返回 `(value, false, nil)` | `FindIDByExternalID` |

**关键约束**:
- ✅ 使用 `ExecWithErrorHandling` / `QueryWithSQLErrorHandling` 包装 DB 操作
- ✅ 返回 `shared.RepositoryError`,不返回 `*bizerrors.BusinessError`
- ✅ Repository 层不打印错误日志
- ✅ 接收 `querier` 接口,支持事务和非事务
- ❌ 模式 A 不检查 `IsNotFoundError` (由 Application 层检查)
- ❌ 不在 Repository 层开启事务

**详细规范**: 见 [repo-develop.md](./repo-develop.md)

## ⚠️ 前置要求

阅读本目录前，请确保已理解 [1-design](../1-design/) 中的设计理念。

## 🔗 相关文档

- 设计原则请参考 [1-design](../1-design/)
- 测试规范请参考 [3-testing](../3-testing/)
