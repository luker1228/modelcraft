# GraphQL Schema 规范

> **优先级: 中** - 定义后端 GraphQL Schema 的组织结构和代码生成工作流。

## 概述

后端 `api/graph/` 目录是 GraphQL Schema 的**唯一真相源**，通过 `gqlgen` 生成 Go 代码。所有 Schema 修改必须先编辑 `.graphql` 文件，再运行代码生成。

---

## Schema 目录结构

```
api/graph/
├── org/schema/              # Org 域 GraphQL Schema
│   ├── base.graphql         # 基础类型（Node、PageInfo 等）
│   ├── end_user.graphql     # EndUser 相关类型（Org 级，如 accessible projects）
│   ├── permission.graphql   # 权限相关类型
│   ├── profile.graphql      # 用户 Profile 类型
│   ├── project.graphql      # 项目 CRUD
│   ├── schema.graphql       # 根 Query/Mutation
│   └── user_management.graphql
└── project/schema/          # Project 域 GraphQL Schema（tenant + end-user 共用）
    ├── base.graphql
    ├── cluster.graphql      # 数据库集群、ModelDatabaseCatalog（错误类型：InvalidInput | ResourceNotFound）
    ├── end_user.graphql     # EndUser Project 级类型（含 end_user_role_users 相关）
    ├── enum.graphql
    ├── field.graphql
    ├── logical_foreign_key.graphql
    ├── model.graphql        # 模型 CRUD，ModelConnection（edges/node relay 分页），ModelQueryInput（databaseName/offset/limit）
    ├── rbac.graphql
    ├── rls.graphql
    └── schema.graphql       # 根 Query/Mutation
```

---

## 两套独立 Schema

后端有**两套独立的 GraphQL Schema**，分别服务在不同 endpoint：

| Schema | 目录 | 服务 URL | 包含内容 |
|--------|------|----------|----------|
| Org GraphQL | `api/graph/org/schema/` | `/graphql/org/{orgName}/` | 项目/集群/用户/角色管理 |
| Project GraphQL | `api/graph/project/schema/` | `/graphql/org/{orgName}/project/{projectSlug}/` | 模型/字段/枚举/外键/分组 |

对应的 gqlgen 配置文件：

| Schema | 配置文件 |
|--------|---------|
| Org GraphQL | `gqlgen.org.yml` |
| Project GraphQL | `gqlgen.project.yml` |

---

## 代码生成工作流

### 修改 Schema 后

```bash
# 1. 编辑 .graphql 文件（禁止直接编辑 generated/ 目录）
vim api/graph/project/schema/model.graphql

# 2. 运行代码生成
just generate-gql

# 3. 实现新增的 resolver 方法（生成后会提示未实现的接口）
```

### 注意事项

- **禁止运行 `just clean-gql`** — 会删除已实现的 resolver 代码
- **禁止直接编辑 `internal/interfaces/graphql/generated/`** — 该目录为自动生成，手动修改会被覆盖
- 每次 `just generate-gql` 只更新生成代码，不影响已实现的 resolver

---

## 关键规则

1. **Schema 优先** — 先修改 `.graphql` 文件，再运行 `just generate-gql`，最后实现 resolver
2. **生成代码只读** — `internal/interfaces/graphql/generated/` 禁止手动编辑
3. **业务域隔离** — Org 和 Project 两套 Schema 独立，不跨 Schema 引用类型
4. **错误类型以 Schema 为准** — 前端 inline fragment `... on XxxError` 中的类型名必须与 Schema 中 union 定义一致；`ModelDatabaseCatalogError = InvalidInput | ResourceNotFound`，不存在 `ProjectNotFound` / `Unauthorized`
5. **models query 用 relay 分页** — `models(input: ModelQueryInput)` 返回 `ModelConnection`（`edges/node`），入参用 `offset/limit`，不是 `page/pageSize`

---

## 参考索引

| 主题 | 文件 |
|------|------|
| Org GraphQL Schema | `api/graph/org/schema/` |
| Project GraphQL Schema | `api/graph/project/schema/` |
| gqlgen Org 配置 | `gqlgen.org.yml` |
| gqlgen Project 配置 | `gqlgen.project.yml` |
| 生成的 Go 代码 | `internal/interfaces/graphql/generated/` |
| Resolver 实现目录 | `internal/interfaces/graphql/` |
