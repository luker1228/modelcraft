# 权限包版本快照

> **版本**: v1.0  
> **状态**: Spec（待后端实现）  
> **关联文档**: [01-permission-model.md](./01-permission-model.md)、[03-auth-flow.md](./03-auth-flow.md)

---

## 问题陈述

权限包（`EndUserPermissionBundle`）当前不保留历史，管理员每次修改权限点列表后，变更立即生效且不可追溯。在生产环境中，这带来两个风险：

1. **误操作无法回滚**：误删权限点后只能手动重新添加。
2. **变更不可审计**：无法回答"这个权限包上周是什么状态"。

---

## 设计目标

| 目标 | 说明 |
|------|------|
| **自动快照** | 每次权限点列表发生变更时，系统自动保存当前状态为只读快照 |
| **滚动保留** | 最多保留最近 5 个版本，超出时自动删除最旧的 |
| **一键回滚** | 管理员可将权限包恢复到任意历史快照，回滚本身生成新版本 |
| **不影响鉴权热路径** | 版本快照为独立存储，鉴权时始终读取当前版本，不引入 JOIN |

---

## 核心概念

### 触发快照的操作

以下操作执行成功后，系统自动为该权限包创建快照：

- `addEndUserPermissionToBundle` — 添加权限点
- `removeEndUserPermissionFromBundle` — 移除权限点

以下操作**不触发**快照（与权限列表无关）：

- `updateEndUserPermissionBundle` — 仅修改名称/描述

### 版本号规则

- 自增整数，从 `1` 开始（`v1`, `v2`, …）
- 每个权限包独立计数
- 回滚操作生成新版本号（非覆盖），内容复制自目标快照，`restoredFromVersion` 字段指向来源

### 滚动保留策略

- 最多保留 **5 个**历史版本（不含当前版本）
- 超出时，删除 `version` 最小的快照
- 回滚生成的新版本参与计数，不享有豁免

---

## 数据模型变更

### 新增表：`end_user_permission_bundle_snapshot`

```sql
CREATE TABLE end_user_permission_bundle_snapshots (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  bundle_id        UUID NOT NULL REFERENCES end_user_permission_bundles(id) ON DELETE CASCADE,
  version          INT  NOT NULL,
  permissions      JSONB NOT NULL,         -- 快照时刻的权限点 ID 数组及 sortOrder
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by       TEXT,                   -- 操作人标识，可为 system
  restored_from    INT,                    -- 若为回滚操作，指向来源版本号；否则 NULL

  UNIQUE (bundle_id, version)
);

CREATE INDEX ON end_user_permission_bundle_snapshots (bundle_id, version DESC);
```

`permissions` 字段结构（JSONB 数组）：

```json
[
  { "permissionId": "uuid-1", "sortOrder": 0 },
  { "permissionId": "uuid-2", "sortOrder": 1 }
]
```

> 快照存储 ID 而非完整数据，避免权限点改名/删除导致快照膨胀。
> 查询快照详情时，通过 ID JOIN 当前权限点获取展示信息，已删除的权限点标注 `[已删除]`。

---

## GraphQL Schema 变更

> 文件：`api/graph/project/schema/rbac.graphql`（或现有的 permission.graphql）

### 新增类型

```graphql
type EndUserPermissionBundleSnapshot {
  version:          Int!
  createdAt:        Time!
  createdBy:        String
  restoredFrom:     Int              # 若为回滚，指向来源版本号
  permissions: [EndUserPermissionSnapshotEntry!]!
}

type EndUserPermissionSnapshotEntry {
  sortOrder:    Int!
  permission:   EndUserPermission    # 已删除时为 null
  permissionId: ID!                  # 原始 ID，即使已删除也保留
}
```

### 在 `EndUserPermissionBundle` 上新增字段

```graphql
type EndUserPermissionBundle {
  # 已有字段 ...

  currentVersion: Int!               # 当前版本号（每次权限列表变更后递增）
  snapshots: [EndUserPermissionBundleSnapshot!]!  # 最近 ≤5 个历史版本，按 version DESC 排列
}
```

### 新增 Mutation

```graphql
# ── 回滚 ──────────────────────────────────────────────────────────────────────

input RestoreEndUserPermissionBundleInput {
  bundleId:      ID!
  targetVersion: Int!
}

type RestoreEndUserPermissionBundlePayload {
  bundle: EndUserPermissionBundle
  newVersion: Int!               # 回滚后生成的新版本号
  error: RestoreEndUserPermissionBundleError
}

union RestoreEndUserPermissionBundleError =
    EndUserPermissionBundleNotFound
  | EndUserPermissionBundleSnapshotNotFound
  | ProjectNotFound

extend type Mutation {
  restoreEndUserPermissionBundle(
    input: RestoreEndUserPermissionBundleInput!
  ): RestoreEndUserPermissionBundlePayload!
}
```

### 新增 Error Type

```graphql
type EndUserPermissionBundleSnapshotNotFound implements Error {
  message: String!
}
```

---

## 后端实现要点

### 快照写入时机

在 `addEndUserPermissionToBundle` / `removeEndUserPermissionFromBundle` 的 Use Case 层，操作成功后：

1. 查询该 bundle 的当前权限列表
2. 序列化为 JSONB
3. 写入 `end_user_permission_bundle_snapshots`，`version = currentVersion + 1`
4. 更新 `end_user_permission_bundles.current_version`（或从 snapshot 表 MAX 派生，无需冗余字段）
5. 删除超出 5 个的最旧快照（`DELETE … WHERE version NOT IN (SELECT version … ORDER BY version DESC LIMIT 5)`）

> 以上步骤在同一事务内执行。

### 回滚逻辑

`restoreEndUserPermissionBundle` Use Case：

1. 查询目标快照的 `permissions` JSONB
2. 以事务替换 bundle 的当前权限点关联（先 DELETE 全部，再按快照 INSERT）
3. 写入新快照（`restored_from = targetVersion`）
4. 触发滚动保留清理

### 鉴权热路径

不变。鉴权时查询 `end_user_bundle_permissions`（当前关联表），不读快照表。

---

## 前端 UI 设计要点

> 前端实现待后端 Schema 就绪后启动。

### 权限包详情页（已有）

在"关联的权限点"列表标题行右侧，显示当前版本号：

```
关联的权限点  [5]  ·  v3
```

点击 `v3` 展开版本历史抽屉（Drawer 或 Sheet），列出最近快照。

### 版本历史列表

每行：

| 版本 | 时间 | 操作人 | 变更摘要 | 操作 |
|------|------|--------|---------|------|
| v3 (当前) | 3分钟前 | admin | +2 项权限 | — |
| v2 | 昨天 | admin | -1 项权限 | 回滚 |
| v1 | 3天前 | system | 初始版本 | 回滚 |

- **变更摘要**：前端通过 diff 相邻版本的 `permissionId` 集合计算，无需后端额外字段
- **回滚**：弹出 AlertDialog 确认，调用 `restoreEndUserPermissionBundle`

### 空历史

初始创建的权限包 `currentVersion = 0`，快照列表为空，不显示版本入口。首次修改权限点后才出现。

---

## 验收场景（BDD 概要）

```gherkin
Scenario: 添加权限点后自动生成快照
  Given 权限包 P 当前有 2 个权限点，currentVersion = 1
  When 添加第 3 个权限点
  Then snapshots 中新增 version=2 的快照，包含 3 个权限点 ID
  And 旧的 version=1 快照仍存在

Scenario: 超出 5 个版本时滚动删除
  Given 权限包 P 已有 5 个历史快照（v1-v5）
  When 再次修改权限点，生成 v6
  Then v1 快照被删除
  And snapshots 中只保留 v2-v6

Scenario: 回滚到历史版本
  Given 权限包 P 当前为 v3（含权限 A, B, C），v2 快照含权限 A, B
  When 管理员回滚到 v2
  Then 权限包当前权限变为 A, B
  And 新增 v4 快照，内容为 A, B，restoredFrom=2
  And 鉴权查询反映新的权限列表

Scenario: 修改名称/描述不触发快照
  Given 权限包 P 当前 currentVersion = 3
  When 修改名称为"新名称"
  Then snapshots 数量不变，currentVersion 仍为 3
```

---

## 变更文件清单（后端）

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `api/graph/project/schema/permission.graphql` | 修改 | 新增 snapshot 类型、bundle 字段、restore mutation |
| `internal/domain/enduserrbac/bundle.go` | 修改 | Bundle 实体新增 `CurrentVersion`、`Snapshots []BundleSnapshot` |
| `internal/domain/enduserrbac/bundle_snapshot.go` | 新增 | `BundleSnapshot` 值对象 |
| `internal/domain/enduserrbac/repository.go` | 修改 | 新增 `SaveBundleSnapshot`、`ListBundleSnapshots`、`DeleteOldSnapshots` |
| `internal/usecase/enduserrbac/add_permission.go` | 修改 | 成功后写快照 |
| `internal/usecase/enduserrbac/remove_permission.go` | 修改 | 成功后写快照 |
| `internal/usecase/enduserrbac/restore_bundle.go` | 新增 | 回滚 Use Case |
| `internal/infrastructure/postgres/enduserrbac/` | 修改/新增 | snapshot 表的 sqlc 查询 + Repository 实现 |
| `db/migrations/` | 新增 | `end_user_permission_bundle_snapshots` 表 Atlas 迁移 |
| `internal/interfaces/graphql/` | 修改 | snapshot resolver、restore mutation resolver |
