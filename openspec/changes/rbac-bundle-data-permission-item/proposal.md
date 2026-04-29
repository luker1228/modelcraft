## Why

当前 RBAC MVP 把 `PRESET` 和 `CUSTOM` 都建模成 `permission` 实体，再让 bundle 直接关联 permission。这把“模板定义”“模型绑定”“bundle 授权”三件事混在了一起，导致预设需要先落库、同一 bundle 下同一模型可被重复配置、查看链路也无法直观表达管理员真正配置的内容。

现在讨论已经收敛，可以直接用破坏性调整把模型收窄成一版更可落地的结构：`preset` 只作为模板定义存在，`custom permission` 作为模型级实体复用，bundle 内真正保存的是“某模型的数据权限 item”。

## What Changes

- **BREAKING** 将 bundle 与数据权限的关联从“直接关联 permission”改为“关联 data permission item”。
- **BREAKING** 约束同一 bundle 下同一模型最多只能存在一个数据权限 item，不能同时绑定 preset 和 custom。
- **BREAKING** 内置 `preset` 不再落库为 `permission` 实体，也不再要求先对模型执行 `applyEndUserPresetPolicy` 才能授权。
- 保留模型强耦合的 `custom data permission` 实体，用于管理员手工定义并在多个 bundle 间复用。
- 新增 bundle item 视图与写入链路：绑定 preset、绑定 custom、查看 bundle items 都围绕 item 展开，而不是围绕 permission 实体展开。
- **BREAKING** bundle 快照从记录 `permission_id` 列表改为记录 data permission item 列表。
- **BREAKING** GraphQL 类型收敛到 item 视角，明确区分“模板”“自定义权限实体”“bundle 中的实际绑定项”。

## Capabilities

### New Capabilities
- `bundle-data-permission-item`: 以 bundle-model 唯一 item 为核心管理数据权限绑定、替换和展示。

### Modified Capabilities
- `enduser-preset-policy`: 预设从“可落库的 permission 实体”调整为“按模型选择的虚拟模板”。
- `bundle-versioning`: 快照对象从 permission 列表调整为 data permission item 列表。
- `bundle-detail-real-data`: bundle 详情与添加弹窗改为消费 data permission item / preset template / custom permission 三类真实数据。

## Impact

- Schema: `modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql`、`modelcraft-backend/db/schema/mysql/14_rbac_bundle_snapshots.sql`
- GraphQL: `modelcraft-backend/api/graph/project/schema/rbac.graphql`
- Domain/App/Repo: `modelcraft-backend/internal/domain/rbac/*`、`modelcraft-backend/internal/app/rbac/*`、`modelcraft-backend/internal/infrastructure/repository/sql_end_user_permission_repository.go`
- Resolver / Adapter: `modelcraft-backend/internal/interfaces/graphql/project/*`
- Frontend bundle 管理页与详情页的查询、mutation、展示模型
