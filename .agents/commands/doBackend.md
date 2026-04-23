---
name: doBackend
description: "后端开发全流程：backend-develop / modelruntime-dev 实现代码，schema-sync-cascade 同步生成代码，test-driven-development 补充高价值逻辑测试，backend-debug 排查问题，bdd-test 流程验收"
argument-hint: "[plan file path or requirement description]"
---

执行后端开发完整流程，覆盖实现、Schema 同步、逻辑测试、调试、验收五个阶段。

## 测试分工说明

本流程采用两层测试策略，职责严格分离：

| 测试类型 | 工具 | 覆盖范围 |
|----------|------|----------|
| **高价值逻辑测试** | `test-driven-development` | Schema 生成、SQL 生成、Adapter 转换（含 Fuzz 测试） |
| **流程验收测试** | `bdd-test` | 用户级业务流程（Use Case / 集成场景） |

> `test-driven-development` **不写 Use Case 测试**，Use Case 测试统一由 `bdd-test` 负责。

## 流程

### Step 1 — 输入校验

- 检查 `$ARGUMENTS` 是否提供了计划文件路径或需求描述
- 如果是文件路径，校验文件存在且可读；缺失则立即停止并提示
- 如果未提供任何输入，提示用户补充计划文件路径或需求说明

### Step 2 — 实现

根据任务是否涉及 `internal/domain/modelruntime/` 或 `internal/app/modelruntime/`，选择对应 skill：

#### 情况 A — 涉及 modelruntime 模块

- 使用 `modelruntime-dev` skill
- 该模块有严格架构约束，**必须**遵守：
  - `graphqlModelResolver` 是无状态 Schema 构建器，不得持有任何 `context.Context` 字段
  - Resolver 闭包内必须通过 `getGraphqlRequestContext(p.Context)` 获取请求级状态
  - 关联关系 resolver 必须返回 Thunk（`func() (interface{}, error)`），不得直接返回结果
  - Schema 构建方法通过参数透传 `ctx`，禁止从 `p` 取 ctx
- 涵盖：Schema 动态生成 / Resolver / Dataloader / graphqlRequestContext / RuntimeModel

#### 情况 B — 常规后端开发

- 使用 `backend-develop` skill
- 按照 DDD 分层架构实现
- 涵盖：DB migration / Repository / App / Resolver / GraphQL Schema

- 产出：可运行的后端实现代码
- 人工确认后继续

### Step 3 — `schema-sync-cascade` 同步生成代码

- 若 Step 2 中修改了 `.graphql` 文件、`.sql` 文件、`gqlgen.*.yml` 或 `sqlc.yaml`，必须执行此步骤
- 使用 `schema-sync-cascade` skill 确保所有生成代码与 Schema 保持同步
- 涵盖：`just generate-gql` / `just generate-sqlc`，同步前后端引用
- 产出：生成代码更新完成，`go build ./...` 无编译错误
- 若 Step 2 未涉及 Schema 变更（如仅改业务逻辑），可跳过此步骤

### Step 4 — `test-driven-development` 高价值逻辑测试

针对以下高价值逻辑补充单元/集成测试（优先级由高到低）：

1. **Adapter 测试**（最高优先级）
   - 针对所有新增或修改的 Adapter / Converter 编写契约测试
   - 必须包含 **Fuzz 测试**：覆盖边界值、零值、异常输入、随机数据
   - 使用 table-driven 风格，覆盖正常路径 + 错误路径

2. **Schema 生成测试**
   - 若涉及动态 GraphQL Schema 构建（如 modelruntime），测试 Schema 结构正确性
   - 覆盖字段类型、关系字段、可选/必填约束

3. **SQL 生成测试**
   - 若涉及动态 SQL 拼接或 sqlc 查询，测试生成的 SQL 语义正确性
   - 覆盖过滤条件、排序、分页、JOIN 逻辑

> **不写的测试**：Use Case / Application Service 的业务流程测试 — 这部分由 `bdd-test` 覆盖。

- 产出：测试文件（`*_test.go`），所有测试通过
- 人工确认后继续

### Step 5 — `backend-debug` 排查（按需）

- 如果 Step 2/3/4 过程中或完成后发现错误（接口报错、日志异常、测试失败）
- 使用 `backend-debug` skill 定位根因并修复
- 产出：问题修复记录
- 修复后回到对应步骤重新验证

### Step 6 — `bdd-test` 流程验收

- 使用 `bdd-test` skill 运行相关领域的 BDD 验收测试
- 覆盖完整业务流程（Use Case 级别）
- 产出：测试报告（通过 / 失败 / 跳过）
- 如有失败，返回 Step 5 排查修复，循环直至全部通过

## 约束

- 必须按 DDD 分层架构实现，不得跨层调用
- **modelruntime 模块**：涉及 `internal/domain/modelruntime/` 或 `internal/app/modelruntime/` 时必须使用 `modelruntime-dev` skill，不得用 `backend-develop` 替代
- **modelruntime 架构红线**：`graphqlModelResolver` 禁止持有 `context.Context` 字段；关联关系 resolver 必须返回 Thunk；请求级状态只能通过 `getGraphqlRequestContext(p.Context)` 获取
- `test-driven-development` 阶段禁止写 Use Case 测试，只写逻辑测试
- Adapter 测试必须包含 Fuzz 测试，不可省略
- 禁止跳过 BDD 验收直接宣布完成
- 如果缺少必要输入文件，立即停止并提示具体缺失路径
- debug 阶段只修复问题，不引入新功能或重构
