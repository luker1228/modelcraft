# End-User GraphQL 子集接入层方案与实施计划

## 0. 目标

在不复制业务逻辑（App/Domain/Repo）的前提下，新增独立的 end-user GraphQL 接入层，形成：

- 开发者：`/graphql/org/{orgName}/project/{projectSlug}`（现有）
- 终端用户：`/graphql/end-user/org/{orgName}/project/{projectSlug}`（新增）

并确保：

1. schema 隔离（契约稳定）
2. 错误码共用（`code` 统一）
3. 权限边界清晰（end-user 仅子集能力）

---

## 0.1 背景与问题复盘（补充上下文）

本期触发改造的真实问题：

1. end-user 页面复用了 developer GraphQL query（`GetModelsForRelation`）
- 现象：end-user 登录后在 `/u/.../data` 页面报无权限。
- 根因：`models(input)` 在 project schema 下受 developer 权限体系保护，end-user 不在该授权模型中。

2. refresh 失败（`INVALID_REFRESH_TOKEN: No refresh token`）
- 现象：登录后刷新页面，BFF `/api/bff/end-user/auth/refresh` 读取不到 cookie。
- 根因：refresh token cookie path 绑定在 `/u/{org}/{project}`，不覆盖 `/api` 路径。

3. 接口边界混杂
- 现象：end-user 数据页部分走 BFF/internal，部分走 developer GraphQL。
- 风险：权限、错误码、字段稳定性都容易漂移。

结论：
- 需要“接入层隔离”，但不需要“业务层复制”。

---

## 0.2 当前已完成改造（截至当前分支）

已落地（可运行）：

1. end-user internal data 接口
- `GET /internal/end-user/data/database-catalog`
- `GET /internal/end-user/data/model-catalog`
- 代码：`internal/interfaces/http/handlers/enduser/data_handler.go`

2. 前端 BFF 数据代理
- `/api/bff/end-user/data/database-catalog`
- `/api/bff/end-user/data/model-catalog`
- `/u/.../data` 页面已移除 `GetModelsForRelation` 调用，改走 end-user BFF 数据链路。

3. refresh cookie 作用域修复
- cookie path 改为 `/`，确保 `/api/bff/end-user/auth/refresh` 可读。

4. end-user GraphQL 契约骨架已建
- `api/graph/end_user/schema/base.graphql`
- `api/graph/end_user/schema/catalog.graphql`
- `gqlgen.end_user.yml`

说明：
- 目前 GraphQL 端点与 resolver 尚未完全接通（处于“契约先行”阶段）。

---

## 0.3 设计约束与不做项

设计约束：

1. 业务逻辑唯一
- App/Domain/Repo 只保留一套实现。

2. 错误码共享
- `PARAM_INVALID`、`UNAUTHORIZED`、`NOT_FOUND` 等 code 统一。

3. 接入层可独立演进
- end-user schema 与 developer schema 分目录、分 endpoint。

本期不做：

1. 不重构现有 developer 权限体系
2. 不将 developer schema 改成“同时服务 end-user”
3. 不在本阶段改动所有 runtime 能力，仅先收口 catalog 与数据页关键链路

---

## 1. 方案列表

## 方案 A：继续仅用 `/internal/end-user/*`（HTTP）

优点：
- 变更最小、上线最快
- 不新增 gqlgen 生成链路

缺点：
- 与现有 GraphQL 生态割裂
- 前端契约演进成本高（字段扩展需手工维护）
- 长期容易出现“REST + GraphQL 双标准”

适用：短期救火。

## 方案 B：新增独立 `api/graph/end_user`（推荐）

优点：
- 与 developer schema 解耦，避免权限错配
- 前端继续 GraphQL 工作流（query/type/mocks）
- 可逐步替代 `/internal/end-user/data/*`

缺点：
- 需要新增一套 gqlgen 配置和 resolver 目录
- 初期有 schema/generated 维护成本

适用：中长期主线。

## 方案 C：在现有 project schema 用 directive 做“子集视图”

优点：
- 复用已有 endpoint 与 schema 文件

缺点：
- 合约边界不清晰，易误开放字段
- schema 演进耦合，review 成本高
- 与“契约隔离”目标冲突

适用：不建议。

---

## 2. 推荐方案

采用 **方案 B**：

- 接口契约隔离：`api/graph/end_user/schema/*.graphql`
- 业务逻辑复用：调用已有 `modeldesign.ModelDesignAppService`、`enduser.EndUserAuthAppService`
- 错误码共享：继续使用现有 `bizerrors` code；仅在 resolver 做 end-user 文案映射
- 迁移策略：先 catalog，后 runtime 读写，最后下线 internal data 接口

---

## 3. 分阶段计划

## Phase 1（契约与路由骨架）

1. 完成 schema：
- `base.graphql`（Query/Error 基础类型）
- `catalog.graphql`（`databaseCatalog`、`modelCatalog`）

2. 完成 `gqlgen.end_user.yml` 并接入生成命令
- 新增 just 命令：`generate-gql-end-user`
- 生成目录：`internal/interfaces/graphql/enduser/generated/*`

3. 新增 endpoint：
- `POST /graphql/end-user/org/{orgName}/project/{projectSlug}`
- `GET  /graphql/end-user/org/{orgName}/project/{projectSlug}`（playground，可开关）

4. 接入中间件：
- internal token（仅 BFF 调用）
- org/project context 注入
- end-user identity 注入（来自 BFF JWT 透传 header）

验收：
- endpoint 可启动，空 query 可返回
- gqlgen 生成代码可编译

## Phase 2（Catalog 能力迁移）

1. 实现 resolver：
- `databaseCatalog(input)` -> 复用 `QueryDatabaseCatalogWithCommand`
- `modelCatalog(input)` -> 复用 `QueryModelsWithCommand`

2. 前端 BFF 改为优先调用 end-user GraphQL endpoint
- 保留 `/internal/end-user/data/*` 作为 fallback

3. `/u/.../data` 不再依赖 `/internal/end-user/data/*`

验收：
- end-user 页面无 `GetModelsForRelation` 权限错误
- database/model 列表都由 end-user GraphQL 返回
- requestId 在错误响应中可追踪

## Phase 3（记录数据能力收口）

1. 增加 end-user runtime 数据 query/mutation 子集：
- list records
- create/update/delete records（按产品开关）

2. `ModelRecordWorkspace` 去 developer 专属依赖
- 抽象 data adapter：developer/end-user 各自实现

3. 统一 capability 控制（前端按钮显隐 + 后端鉴权）

验收：
- `/u/*` 数据读写链路完全不经过 developer GraphQL endpoint

## Phase 4（清理与收敛）

1. 废弃 `/internal/end-user/data/*`（保留 auth 相关 internal 路由）
2. 文档化：
- API contract
- 错误码映射
- 变更准入 checklist

验收：
- end-user 入口唯一（GraphQL 子集）
- 无重复业务实现

---

## 4. 技术设计要点

1. 认证模型
- BFF 验证 end-user access token
- BFF 调后端时透传：`X-End-User-Id`、`X-Org-Name`、`X-Project-Slug`
- 后端 end-user GraphQL endpoint 不接受 developer JWT

2. 错误模型
- 共享 code：`PARAM_INVALID/UNAUTHORIZED/NOT_FOUND/...`
- endpoint 级别统一返回 `requestId`
- 文案按 end-user 语义做轻量包装

3. 分页规范（catalog）
- 入参：`page`、`pageSize`、`search`
- 默认值：`page=1`，`pageSize=20`
- 上限：`pageSize<=100`

4. 安全边界
- schema 不暴露建模能力（createModel/addField/enum 管理）
- 仅暴露 end-user 被授权的数据访问子集

5. 性能约束
- catalog 查询必须支持分页
- 避免一次性返回 fields/jsonSchema 等重字段
- 前端按需加载（选择数据库后再加载模型）

---

## 5. API 形态示例（用于评审）

1. `databaseCatalog`

```graphql
query EndUserDatabaseCatalog($input: DatabaseCatalogInput) {
  databaseCatalog(input: $input) {
    data {
      databases { name }
      totalCount
      page
      pageSize
    }
    error {
      __typename
      ... on InvalidInput { message }
      ... on Unauthorized { message }
      ... on ProjectNotFound { message }
    }
  }
}
```

2. `modelCatalog`

```graphql
query EndUserModelCatalog($input: ModelCatalogInput!) {
  modelCatalog(input: $input) {
    data {
      models { id name title databaseName }
      totalCount
      page
      pageSize
    }
    error {
      __typename
      ... on InvalidInput { message }
      ... on Unauthorized { message }
      ... on ProjectNotFound { message }
    }
  }
}
```

---

## 6. 代码落点（规划）

- Schema：`api/graph/end_user/schema/*.graphql`
- gqlgen config：`gqlgen.end_user.yml`
- Resolver：`internal/interfaces/graphql/enduser/*`
- Router：`internal/interfaces/http/routes.go`
- 前端 BFF：`modelcraft-front/src/app/api/bff/end-user/*`
- 前端页面：`modelcraft-front/src/app/u/[orgName]/[projectSlug]/data/page.tsx`

---

## 7. 风险与回滚

风险：

1. 两套 GraphQL 生成链路引入维护负担
2. 前端过渡期存在双链路（internal data + end-user graphql）
3. 鉴权 header 漏传导致 401
4. cookie 策略变更导致跨路径读取行为变化

控制：

1. 先从 catalog 低风险能力迁移
2. 保留 `/internal/end-user/data/*` 作为短期回退
3. 增加 BFF 集成测试：`login -> refresh -> databaseCatalog -> modelCatalog`
4. 明确 cookie/path/sameSite 策略并写回归用例

回滚：

- 前端开关切回 `/internal/end-user/data/*`
- 后端新 GraphQL endpoint 保留但不被调用
- cookie path 出问题可回退到上一版并保留 token 验证兜底

---

## 8. 验收清单（可直接打勾）

1. `/u/{org}/{project}/data` 不再出现 `GetModelsForRelation` 权限错误
2. 刷新页面后不会出现 `No refresh token`
3. end-user catalog 查询只返回轻量字段
4. requestId 在错误响应中可见
5. backend 与 frontend build 均通过
6. developer 路径行为不回归

---

## 9. 执行顺序建议

1. 先合并 Phase 1 + Phase 2（最小可用）
2. 你审核 schema 后再做 Phase 3（record 能力）
3. 最后统一清理 Phase 4

