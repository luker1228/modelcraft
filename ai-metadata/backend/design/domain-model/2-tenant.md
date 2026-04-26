# 2. 租户隔离

> 代码位置：`internal/domain/organization/`

## 概述

Organization 是 ModelCraft 的多租户顶层容器。所有资源（Project、Cluster、Model、User）都归属于某个 Org，Org 是最顶层的隔离边界。

## 核心实体

```
internal/domain/organization/organization.go

Organization
├── ID           string      // UUID
├── Name         string      // 唯一标识符，来自 AuthProvider 组织名
│                            // 格式：2-64 字符，小写字母开头，允许数字/下划线/连字符
├── DisplayName  string      // UI 显示名称（可选）
├── OwnerID      string      // 创建者用户 ID
└── Status       OrgStatus   // active | suspended | deleted
```

## 隔离机制

Org 的 `Name` 是系统中所有资源路由的根节点：

```
URL 路由示例：
  /org/{orgName}/design/graphql          ← 设计态入口
  /{orgName}/{projectSlug}/{db}/{model}   ← 运行态入口
```

请求进入时，从 URL 或 JWT 中提取 `orgName`，注入 context，后续所有查询均以此为隔离键。

## 生命周期

```
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

## 与 AuthProvider 的关系

- Org 的 `Name` 与 AuthProvider 中的 Organization Name 保持一致
- ModelCraft 不自管 Org 创建流程，由 AuthProvider webhook 触发同步

## 相关文件

- `internal/domain/organization/organization.go` — 实体定义
- `internal/domain/organization/repository.go` — 仓储接口
