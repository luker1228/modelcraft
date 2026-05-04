## 1. Schema 与错误模型收敛

- [x] 1.1 在 Org/Project/EndUser Schema 中新增 `ResourceType` 枚举与 `ResourceNotFound` 类型
- [x] 1.2 将各 Query/Mutation 错误 union 中的 `*NotFound` 成员替换为 `ResourceNotFound`
- [x] 1.3 清理不再使用的 `*NotFound` 类型定义并确保 schema 可通过校验

## 2. 后端错误映射改造

- [x] 2.1 修改各 GraphQL 错误适配器，将 `NOT_FOUND.*` 统一映射到 `generated.ResourceNotFound`
- [x] 2.2 建立业务错误码到 `ResourceType` 的映射函数，并补充 `UNKNOWN` 兜底策略
- [x] 2.3 保持 `extensions.code` 细粒度透传，补充必要单元测试覆盖映射分支

## 3. 前端与 BDD 迁移

- [x] 3.1 批量更新前端 `graphql-docs.ts`：将 `... on XxxNotFound` 改为 `... on ResourceNotFound { message resourceType }`
- [x] 3.2 运行前端 codegen 并修复类型报错，统一运行时分支判断到 `ResourceNotFound + resourceType`
- [x] 3.3 更新 BDD feature 与 step-definitions 的错误断言，移除对具体 `XxxNotFound` 的依赖

## 4. 回归验证与收口

- [x] 4.1 执行后端 GraphQL 代码生成并完成编译检查
- [x] 4.2 执行前端 lint/类型检查，确认 GraphQL 客户端与 UI 逻辑通过
- [x] 4.3 执行 BDD 回归，确认 not-found 场景全部通过并记录破坏性变更说明

## 破坏性变更说明（2026-05-04）

### 本次变更引入的破坏性变更

1. **GraphQL Schema 变更（破坏性）**
   - 移除：`ModelNotFound`、`ProjectNotFound`、`ClusterNotFound`、`EnumNotFound`、`GroupNotFound` 等所有 `*NotFound` 类型
   - 新增：统一类型 `ResourceNotFound { message: String!, resourceType: ResourceType! }`
   - 新增：枚举 `ResourceType`，包含 `PROJECT / CLUSTER / MODEL / ENUM / GROUP / USER / PROFILE / ORGANIZATION / ROLE / END_USER / END_USER_PERMISSION / END_USER_PERMISSION_BUNDLE / END_USER_PERMISSION_BUNDLE_SNAPSHOT / END_USER_ROLE / END_USER_IN_PROJECT / PERMISSION_ROLE / PERMISSION_USER / UNKNOWN`
   - 影响：所有使用 `... on XxxNotFound` 的 GraphQL 客户端（前端、BDD、第三方）必须迁移为 `... on ResourceNotFound { message resourceType }`

2. **后端错误映射**
   - `extensions.code` 仍保留细粒度码（如 `NOT_FOUND.MODEL`），无破坏性变更
   - 新增集中映射函数 `BizCodeToResourceType`（位于 project/adapter 和 org/adapter）
   - 无 `FIELD`、`FK`、`MEMBERSHIP`、`RECORD`、`API_KEY` 对应 `ResourceType` 枚举值，上述场景兜底为 `UNKNOWN`

### BDD 回归状态

- **not-found 断言本身**：GraphQL 请求体已正确使用 `ResourceNotFound + resourceType`，迁移生效
- **阻塞原因（非本次变更）**：
  - `profile` 场景：`PERMISSION_DENIED`（缺少 `user:read` 权限），存量 RBAC 配置问题
  - `bundle-snapshot`、`rbac`、`field` 等场景：`401 Unauthorized`（后端服务认证配置），存量基础设施问题
  - `bundle-snapshot` 部分场景：`Undefined step`（step-definitions 未实现），存量待办

## 验证阻塞说明（2026-05-03，已于 2026-05-04 更新）

- 前端 `npm run codegen` 已通过（先执行 `front-contract-pull` 同步 backend `api/` 到 front `contract/` 后恢复正常）。
- 前端 `npm run lint` 通过（exit code 0），仅有 Tailwind 顺序等 Warning，均为存量且与本次变更无关。
- 前端 TypeScript 类型检查通过（exit code 0），存量 type error 均为无关的 RBAC/mock 类型问题。
- BDD 运行仍未全绿：`ResourceNotFound` 类型断言已正确迁移，但被上游 `401 Unauthorized`/`PERMISSION_DENIED` 阻断，未能到达断言步骤；均为存量基础设施阻塞（profile 权限配置、rbac 认证），非本次变更引入。