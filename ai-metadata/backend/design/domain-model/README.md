# Domain Model

> 本目录描述 ModelCraft 的领域模型规划。内容以代码为准——文档描述的是代码中实际存在的领域概念。

## 领域划分

| # | 领域 | 代码目录 | 说明 |
|---|------|----------|------|
| 1 | 登录与认证 | `internal/domain/auth/` `internal/domain/user/` `internal/domain/enduser/` | 自建用户体系：管理员自注册，普通用户由管理员创建 |
| 2 | 租户隔离 | `internal/domain/organization/` | 多租户容器，资源隔离边界；注册即建 Org |
| 3 | 项目分离 | `internal/domain/project/` | Org 内的项目隔离 |
| 4 | 用户鉴权 | `internal/domain/membership/` `internal/domain/role/` `internal/domain/permission/` | 用户-组织关系、RBAC |
| 5 | 模型 | `internal/domain/modeldesign/` `internal/domain/modelruntime/` | 模型设计 + 运行态产物（见子目录） |
| 6 | 数据库连接 | `internal/domain/cluster/` | MySQL 集群连接管理 |
| 7 | SQL 编辑器 | _(未实现)_ | 直接执行 SQL，future milestone |

## 双视角用户体系

```
同一 users 表，通过 user_orgs.is_admin 区分角色：

  管理员视角（is_admin = 1）
  ├── 自注册（手机号+密码）→ 自动建 Org
  ├── 管理 Org/Project/Model/Cluster
  ├── 创建普通用户并分配角色
  └── 路由前缀：/api/tenant/auth/* | /graphql/org/{orgName}/

  普通用户视角（is_admin = 0）
  ├── 由管理员创建（不可自注册）
  ├── 访问被授权的 Project 数据
  └── 路由前缀：/api/end-user/auth/* | CLI /api/cli/end-user/auth/*
```

> 一个用户可以同时拥有两个视角（双角色）。

## 文档目录

```
domain-model/
├── 1-auth.md               登录与认证（用户体系、双视角、Token 设计）
├── 2-tenant.md             租户隔离（Org 生命周期、用户与 Org 关系）
├── 3-project.md            项目分离
├── 4-rbac.md               用户鉴权（RBAC）
├── 5-model/
│   ├── README.md           模型领域概览
│   ├── design.md           设计态：模型定义
│   └── artifact.md         运行态：模型产物
├── 6-database-cluster.md   数据库连接
└── 7-sql-editor.md         SQL 编辑器（规划中）
```
