# 6. 数据库连接管理

> 代码位置：`internal/domain/cluster/`

## 概述

DatabaseCluster 代表一个可连接的 MySQL 数据库实例。它是 Project 与真实数据库之间的桥梁，也是运行态 GraphQL 执行 SQL 的连接来源。

## 核心实体

```
internal/domain/cluster/database_cluster.go

DatabaseCluster
├── ID                string
├── OrgName           string
├── ProjectSlug       string
├── Title             string
├── Description       string
├── Host              string
├── Port              int            // 默认 3306
├── Username          string
├── Password          Password       // 加密存储，见 password.go
├── ConnectionTimeout int            // 5-15 秒，默认 5
├── Status            ClusterStatus  // active | disabled
└── Version           int64
```

### Password — 密码值对象

```
internal/domain/cluster/password.go

密码在存储前加密，通过 Password 类型封装：
- NewByPlain(plain string)   从明文创建（触发加密）
- Decrypt() string           解密获取明文（用于建立连接时）
```

明文密码不在系统中流转，只在建立数据库连接时临时解密使用。

## ClusterLocator

```
ClusterLocator
├── OrgName     string
└── ProjectSlug string

GetFullPath() → "orgName.projectSlug"
```

用于在多租户环境下唯一定位一个 Cluster。

## 与 Project 的关系

```
Project (OrgName, Slug)
        │
        └── ClusterID ──▶ DatabaseCluster
                          (1:1 可选关联)
```

- 一个 Project 最多关联一个 Cluster
- Cluster 归属于 Project，不跨 Project 共享

## 连接池管理

Cluster 实体只定义连接参数，实际的连接池由 Infrastructure 层管理（`internal/infrastructure/`），Domain 层不感知连接池细节。

## 相关文件

- `internal/domain/cluster/database_cluster.go` — 实体定义
- `internal/domain/cluster/password.go` — 密码加密值对象
- `internal/domain/cluster/database_cluster_repository.go` — 仓储接口
