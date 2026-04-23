---
name: schema-sync-cascade
description: >-
  当 GraphQL schema 或 SQLc 查询发生变更时使用本 skill——包括重命名类型/输入/mutation、
  添加/删除字段、修改 SQL 查询签名或 sqlc YAML 配置。触发场景：「重命名这个 GraphQL 类型」、
  「给 mutation 加一个字段」、「修改这个 SQL 查询」、「更新 schema」、「修改 input」，
  或对 api/graph/{org,project}/schema/ 或 db/queries/ 下的 .graphql / .sql 文件进行任何编辑。
  也适用于编辑 gqlgen.org.yml、gqlgen.project.yml、sqlc.yaml 或 justfile 的生成目标时。
  有疑问就用这个 skill——它能防止生成代码过期和引用断裂。
---

# Schema Sync Cascade

当 GraphQL schema 或 SQLc 查询被修改时，本 skill 确保所有引用保持同步、生成代码保持最新。
核心原则：**源 schema 的变更必须级联到每一个消费方**。

## 项目结构

```
modelcraft-backend/                           # 后端
  api/graph/{org,project}/schema/*.graphql    # GraphQL 源 schema（在此编辑）
  internal/interfaces/graphql/*/*.resolvers.go # Resolver 实现
  internal/interfaces/graphql/*/generated/*.go # gqlgen 输出（自动生成）
  db/queries/*.sql                            # SQLc 查询源文件
  internal/infrastructure/dbgen/*.go          # sqlc 输出（自动生成）
  internal/infrastructure/dbgenwrap/safe_querier_gen.go  # SafeQuerier wrapper（自动生成）
  sqlc.yaml                                   # sqlc 配置

modelcraft-front/                             # 前端
  src/graphql/mutations/*.ts                  # GraphQL mutation 文档
  src/graphql/queries/*.ts                    # GraphQL query 文档
  src/types/index.ts                          # TypeScript 类型定义
```

## 工作流

### Step 1：识别变更内容

仔细阅读用户指令，确认：
- **GraphQL**：哪些类型、input、mutation 或 query 被重命名/新增/删除？
- **SQLc**：哪些 SQL 查询在变更？列名、表名或返回类型是否受影响？
- 提取旧名称与新名称（重命名场景），或新的结构（新增场景）。

### Step 2：查找所有引用（GraphQL）

在两个代码库中并行搜索。**不要修改生成目录**——它们会被重新生成。

**后端（`modelcraft-backend/`）：**
1. `api/graph/{org,project}/schema/*.graphql` — 源 schema 本身
2. `internal/interfaces/graphql/*/*.resolvers.go` — resolver 方法名与 `generated.*` 类型引用
3. `tests/` — 引用旧名称的测试文件

**前端（`modelcraft-front/`）：**
1. `src/graphql/mutations/*.ts` — mutation 文档字符串（`mutation NameName(...) { oldName(...) }`）
2. `src/graphql/queries/*.ts` — query 文档字符串
3. `src/types/index.ts` — 与 GraphQL 类型对应的 TypeScript 接口
4. `src/hooks/*.{ts,tsx}` — 导入并调用 mutation/query 的自定义 Hook
5. `src/app/**/*.tsx` — 直接使用 mutation 或读取响应字段的页面组件

使用 Grep 查找所有出现位置，在编辑前向用户报告发现结果。

### Step 3：查找所有引用（SQLc）

**仅限后端：**
1. `db/queries/*.sql` — SQL 查询源文件
2. `internal/infrastructure/dbgen/*.go` — **不要编辑**（自动生成）
3. `internal/infrastructure/dbgenwrap/safe_querier_gen.go` — **不要编辑**（自动生成）
4. `internal/` 下调用 sqlc 生成函数的 Go 文件（通过 `SafeQuerier` 接口调用）
5. 封装 sqlc 调用的 Repository 接口与实现

### Step 4：应用变更

编辑 Step 2-3 中找到的所有非生成文件。重命名场景：
- GraphQL schema：类型/input/mutation 名称
- Resolver Go 代码：方法名 + `generated.NewName` 类型引用
- 前端 mutation 字符串：GraphQL 文档中的字段名
- 前端类型：接口名称及变量类型注解
- 前端 Hook/页面：响应数据访问路径（`data?.oldName` → `data?.newName`）
- SQL 查询名称：`-- name: OldName` → `-- name: NewName`
- Go 调用方：`q.OldName()` → `q.NewName()`

### Step 5：重新生成代码

schema 变更后必须重新生成。在 `modelcraft-backend/` 下执行：

```bash
# GraphQL schema 变更后：
just generate-gql

# SQLc 查询变更后（同时生成 sqlc 代码 + SafeQuerier wrapper）：
just generate-safe-querier

# 两者都变更：
just generate-gql && just generate-safe-querier
```

> `just generate-safe-querier` 内部会依次执行：
> 1. `sqlc generate` — 生成 `internal/infrastructure/dbgen/`
> 2. gowrap 脚本 — 生成 `internal/infrastructure/dbgenwrap/safe_querier_gen.go`
>
> **不要单独运行 `sqlc generate`**，应始终使用 `just generate-safe-querier` 以保证两层输出同步。

### Step 6：验证

重新生成后：
1. 在 `modelcraft-backend/` 下执行 `go build ./...`，检查编译错误
2. 如果构建失败，排查原因——很可能有引用遗漏，返回 Step 2
3. **不要手动编辑** `generated/` 或 `dbgen/` 下的生成文件

## 核心规则

- **永远不要编辑生成代码。** 始终修改源文件后重新生成。
- **搜索两个代码库。** GraphQL 变更影响前端；SQLc 变更仅影响后端。
- **先重新生成再验证。** 过期的生成代码会导致假编译错误。
- **响应字段路径也要更新。** 将 mutation 从 `updateModel` 重命名为 `updateModelMeta` 时，前端所有 `data?.updateModel` 都必须改为 `data?.updateModelMeta`。
