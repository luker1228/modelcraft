# syncModelsFromDB Design

**Date:** 2026-06-16  
**Status:** Approved  
**Scope:** 统一导入/同步模型入口，替代现有 `importModel`

---

## 背景与目标

### 现状

现有 API 有两个语义重叠的入口：

| API | 行为 | 问题 |
|-----|------|------|
| `importModel(databaseName, tableName)` | 导入新模型，若 model 已存在则报错 | 只做新建，前端调用前需判断 model 是否存在 |
| `startModelDatabaseSync(databaseId)` | 全量异步同步整个 DB 所有表（已有 + 新建） | 只能全量，不支持指定表；入参用 `databaseId` |

两者底层都是 introspect table → upsert model，但暴露的语义和入参不一致，前端被迫分两条路径处理。

### 目标

1. 提供统一的 `syncModelsFromDB` Mutation，支持按表名列表或全量同步
2. 提供 `modelSyncJob` Query 查询异步任务状态
3. 统一语义：upsert（model 不存在建新，已存在更新字段）
4. 废弃 `importModel`

### 非目标

- 不替代 `startModelDatabaseSync`（用于数据库管理页，保留）
- 不做 WebSocket / Subscription 推送
- 不处理视图、存储过程等非表对象

---

## Section 1：GraphQL API

### Mutation：syncModelsFromDB

```graphql
input SyncModelsFromDBInput {
  databaseName: String!
  tableNames:   [String!]  # 指定同步的表（与 syncAll 二选一）
  syncAll:      Boolean    # 显式传 true 才触发全量同步（与 tableNames 二选一）
}

type SyncModelsFromDBPayload {
  jobId: ID!
}

extend type Mutation {
  syncModelsFromDB(input: SyncModelsFromDBInput!): SyncModelsFromDBPayload!
}
```

**参数校验：**

| 情况 | 结果 |
|------|------|
| `tableNames` 非空 + `syncAll` 未传或 false | ✅ 同步指定表 |
| `syncAll: true` + `tableNames` 未传 | ✅ 全量同步 |
| 两者都传 | ❌ `ParamInvalid`：不能同时指定 tableNames 和 syncAll |
| 两者都未传 | ❌ `ParamInvalid`：必须指定 tableNames 或 syncAll=true |
| `tableNames` 为空数组 | ❌ `ParamInvalid`：tableNames 不能为空 |
| 同一 `databaseName` 已有 PENDING/RUNNING job | ❌ `JobAlreadyRunning` |

校验失败均通过业务错误系统返回（与现有 `startModelDatabaseSync` 保持一致，不使用 union 类型），前端通过 GraphQL errors 字段读取错误码。

立即返回 `jobId`，后台异步执行。

### Query：modelSyncJob

```graphql
extend type Query {
  modelSyncJob(jobId: ID!): ModelSyncJob
}

enum ModelSyncJobStatus {
  PENDING
  RUNNING
  SUCCEEDED
  PARTIAL_SUCCESS
  FAILED
}

type ModelSyncFailedTable {
  tableName: String!
  message:   String!
}

type ModelSyncJob {
  id:               ID!
  databaseName:     String!
  tableNames:       [String!]!  # 空数组 = 全量
  status:           ModelSyncJobStatus!
  totalTables:      Int!
  processedTables:  Int!
  createdModels:    Int!
  syncedModels:     Int!
  failedCount:      Int!
  failedTables:     [ModelSyncFailedTable!]!
  startedAt:        Time
  finishedAt:       Time
  createdAt:        Time!
  updatedAt:        Time!
}
```

Endpoint：Project GraphQL（`/graphql/org/{orgName}/project/{projectSlug}/`）

---

## Section 2：执行行为

### 每张表的 upsert 语义

1. introspect 该表的列定义
2. **model 不存在** → 创建新 model，复用现有 `ImportModel` 逻辑
3. **model 已存在** → full sync 字段（仅处理 `storageHint` 非空的字段，以 DB 列为准）：
   - DB 有对应列，model 已有该字段 → **强制覆盖**字段属性（类型、nonNull、unique 等）
   - DB 有对应列，model 无该字段 → **新增字段**
   - DB 无对应列，model 有该字段 → **删除字段**
   - `storageHint` 为空的字段（逻辑字段、关联字段等）**完全不动**
4. 单表失败 → 记入 `failedTables`，继续处理下一张表

### 字段同步范围（以 storageHint 为准）

**只处理 `storageHint` 非空的字段**，即有明确 DB 列映射的字段：

- DB 有对应列，model 已有该字段（storageHint 匹配）→ **强制覆盖**：以 DB 列当前定义为准，全量更新字段属性（类型、nonNull、unique 等），不保留旧值
- DB 有对应列，model 无该字段 → **新增字段**
- DB 无对应列，model 有该字段（storageHint 非空）→ **删除字段**
- `storageHint` 为空的字段（逻辑字段、关联字段等）→ **完全不动**

### Job 状态机

```
PENDING → RUNNING → SUCCEEDED
                  → PARTIAL_SUCCESS
                  → FAILED
```

- 无失败：`SUCCEEDED`
- 有成功也有失败：`PARTIAL_SUCCESS`
- 全部失败或任务级初始化失败：`FAILED`

### 并发控制

同一 `(orgName, projectSlug, databaseName)` 在 `PENDING/RUNNING` 状态下只允许一个 job，新请求直接返回 `JobAlreadyRunning` 错误（应用层校验）。

---

## Section 3：数据模型

### 新表：`model_sync_job`

（与现有 `model_database_sync_job` 独立，入参维度不同：这里用 databaseName，不依赖 databaseId）

```sql
CREATE TABLE model_sync_job (
  id                VARCHAR(36)   NOT NULL,
  org_name          VARCHAR(64)   NOT NULL,
  project_slug      VARCHAR(64)   NOT NULL,
  database_name     VARCHAR(128)  NOT NULL,
  table_names       JSON          NOT NULL,  -- 空数组 = 全量
  status            ENUM('pending','running','succeeded','partial_success','failed') NOT NULL,
  total_tables      INT           NOT NULL DEFAULT 0,
  processed_tables  INT           NOT NULL DEFAULT 0,
  created_models    INT           NOT NULL DEFAULT 0,
  synced_models     INT           NOT NULL DEFAULT 0,
  failed_count      INT           NOT NULL DEFAULT 0,
  failed_tables     JSON          NOT NULL,  -- [{tableName, message}]
  started_at        DATETIME(3)   NULL,
  finished_at       DATETIME(3)   NULL,
  created_at        DATETIME(3)   NOT NULL,
  updated_at        DATETIME(3)   NOT NULL,

  PRIMARY KEY (id),
  INDEX idx_org_project_db_created (org_name, project_slug, database_name, created_at),
  INDEX idx_org_project_status (org_name, project_slug, status)
);
```

---

## Section 4：前置 DB Schema 变更

`field_definitions` 表当前缺少 `storage_hint` 列，需先补充：

```sql
-- field_definitions 新增列
ALTER TABLE field_definitions
  ADD COLUMN `storage_hint` VARCHAR(128) NULL
    COMMENT '存储优化提示，通常为 DB 列名；非空表示该字段映射到实际 DB 列，参与 syncModelsFromDB 的 full sync'
    AFTER `is_deprecated`;
```

对应变更：
- `db/schema/mysql/03_model_domain.sql`：`field_definitions` 表定义加 `storage_hint` 列
- `db/queries/field.sql`：`CreateFieldDefinition` 和 `UpdateField` 加 `storage_hint` 参数
- sqlc 重新生成后，infrastructure 层的 `CreateFieldDefinitionParams` / `UpdateFieldParams` 自动包含该字段

---

## Section 5：与现有代码的关系

| 现有能力 | 复用方式 |
|---------|---------|
| `ReverseEngineerAppService.ImportModel` | 复用：model 不存在时调用 |
| `ModelDesignAppService`（字段增删逻辑） | 复用：model 已存在时执行 full sync |
| `ModelDatabaseSyncJob`（`startModelDatabaseSync`） | 保留，不影响 |
| `importModel` GraphQL mutation | 标记废弃，迁移完成后移除 |

---

## Section 6：废弃 `importModel`

- `importModel` 唯一调用方：`ImportModelDialog.tsx`
- 迁移方案：将 Dialog 改为调用 `syncModelsFromDB`（传 `tableNames: [tableName]`）
- 废弃顺序：先上 `syncModelsFromDB`，前端迁移后，再从 schema 中移除 `importModel`

---

## Section 7：测试策略

### 后端

1. **应用层单测**
   - 指定 tableNames 同步
   - syncAll=true 全量同步
   - model 不存在 → 创建
   - model 已存在 → 字段 full sync：storageHint 非空字段强制覆盖属性、增删；storageHint 为空字段不动
   - 单表失败不中断
   - 同 DB 有 active job → 报错
2. **Repository 单测**：CRUD + active job 检查
3. **GraphQL 层**：resolver 参数校验（互斥、空数组）

### 前端

1. `ImportModelDialog` 迁移到 `syncModelsFromDB`
2. 轮询 `modelSyncJob` 状态展示
3. 成功 / 部分成功 / 失败结果展示

---

## Section 8：文件变更清单

### 后端

| 文件 | 操作 |
|------|------|
| `db/schema/mysql/03_model_domain.sql` | 修改，`field_definitions` 加 `storage_hint` 列 |
| `db/queries/field.sql` | 修改，`CreateFieldDefinition` / `UpdateField` 加 `storage_hint` |
| `api/graph/project/schema/model.graphql` | 新增 `syncModelsFromDB` mutation、`ModelSyncJob` 类型、`modelSyncJob` query |
| `internal/domain/modeldatabase/` | 新增 `ModelSyncJob` domain entity 和 repository 接口 |
| `internal/app/modeldatabase/` | 新增 `SyncModelsFromDBUseCase`（触发 + 执行） |
| `internal/infrastructure/repository/` | 新增 `SqlModelSyncJobRepository` |
| `internal/interfaces/graphql/project/` | 新增 mutation/query resolver |
| `db/schema/mysql/*.sql` | 新增 `model_sync_job` 表 |
| `db/queries/*.sql` | 新增相关 sqlc 查询 |

### 前端

| 文件 | 操作 |
|------|------|
| `src/api-client/project/model-graphql-docs.ts` | 新增 `syncModelsFromDB` / `modelSyncJob` GraphQL 文档 |
| `src/web/components/features/model-editor/ImportModelDialog.tsx` | 迁移：调用 `syncModelsFromDB` 替代 `importModel` |
| `src/web/hooks/model/` | 新增 `useSyncModelsFromDB` hook |

---

## Section 9：实施顺序

1. **DB schema**：`03_model_domain.sql` 加 `storage_hint` 列 + sqlc 重新生成
2. 后端：GraphQL schema（`model.graphql`）+ codegen
3. 后端：`model_sync_job` 表 + sqlc 查询
4. 后端：domain entity + repository
5. 后端：app usecase（触发 + 执行逻辑）
6. 后端：resolver
7. 前端：contract 同步 + hooks
8. 前端：`ImportModelDialog` 迁移
9. 后续：移除 `importModel`
