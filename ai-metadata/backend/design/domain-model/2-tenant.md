# 2. 租户隔离

> 代码位置：`internal/domain/organization/`

## 概述

Organization 是 ModelCraft 的多租户顶层容器。所有资源（Project、Cluster、Model、User）都归属于某个 Org，Org 是最顶层的隔离边界。

## 核心实体

```
internal/domain/organization/organization.go

Organization
├── ID           string      // UUID
├── Name         string      // 唯一标识符，格式：2-64 字符，小写字母开头，允许数字/下划线/连字符
├── DisplayName  string      // UI 显示名称（可选）
├── OwnerID      string      // 创建者用户 ID（→ users.id）
└── Status       OrgStatus   // active | suspended | deleted
```

## 隔离机制

Org 的 `Name` 是系统中所有资源路由的根节点：

```
URL 路由示例：
  /graphql/org/{orgName}/                    ← 设计态 Org GraphQL
  /graphql/org/{orgName}/project/{slug}/     ← 设计态 Project GraphQL
  /org/{orgName}/{projectSlug}/{db}/{model}  ← 运行态入口
```

请求进入时，从 URL 或 JWT 中提取 `orgName`，注入 context，后续所有查询均以此为隔离键。

## 生命周期

```
注册管理员（Register）
      │ 自动触发
      ▼
NewOrganization()
      │
      ▼
   active  ──Suspend()──▶  suspended  ──Activate()──▶  active
      │
  MarkDeleted()
      │
      ▼
   deleted（软删除，不可恢复）
```

**注册即建 Org**：管理员注册时系统自动为其创建一个个人 Org，注册者成为该 Org 的 Owner 兼管理员（`user_orgs.is_admin = 1`）。

## 用户与 Org 的关系

```
users (全局用户表)
    │
    │  user_orgs
    ├── is_admin = 1  →  管理员：可管理 Org 下所有资源，可创建普通用户
    └── is_admin = 0  →  普通用户：只能访问被授权的 Project 数据
```

- 每个用户只能属于一个 Org（`uk_user_orgs_user` 唯一约束）
- 一个用户可以同时是管理员和普通用户（`is_admin` 字段控制）
- **普通用户不能自注册**，只能由管理员在 Org 内创建

## 相关文件

- `internal/domain/organization/organization.go` — 实体定义
- `internal/domain/organization/repository.go` — 仓储接口
- `internal/app/organization/create_organization_service.go` — 注册时自动建 Org
- `db/schema/mysql/05_organizations.sql` — 数据库 Schema
- `db/schema/mysql/06_users.sql` — users + user_orgs 表
