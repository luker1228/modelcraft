# Domain Model

> 本目录描述 ModelCraft 的领域模型规划。内容以代码为准——文档描述的是代码中实际存在的领域概念。

## 领域划分

| # | 领域 | 代码目录 | 说明 |
|---|------|----------|------|
| 1 | 登录 | `internal/domain/auth/` | 认证与 JWT 验证 |
| 2 | 租户隔离 | `internal/domain/organization/` | 多租户容器，资源隔离边界 |
| 3 | 项目分离 | `internal/domain/project/` | Org 内的项目隔离 |
| 4 | 用户鉴权 | `internal/domain/membership/` `internal/domain/role/` `internal/domain/permission/` | 用户-组织关系、RBAC |
| 5 | 模型 | `internal/domain/modeldesign/` `internal/domain/modelruntime/` | 模型设计 + 运行态产物（见子目录） |
| 6 | 数据库连接 | `internal/domain/cluster/` | MySQL 集群连接管理 |
| 7 | SQL 编辑器 | _(未实现)_ | 直接执行 SQL，future milestone |

## 文档目录

```
domain-model/
├── 1-auth.md               登录与认证
├── 2-tenant.md             租户隔离
├── 3-project.md            项目分离
├── 4-rbac.md               用户鉴权（RBAC）
├── 5-model/
│   ├── README.md           模型领域概览
│   ├── design.md           设计态：模型定义
│   └── artifact.md         运行态：模型产物
├── 6-database-cluster.md   数据库连接
└── 7-sql-editor.md         SQL 编辑器（规划中）
```
