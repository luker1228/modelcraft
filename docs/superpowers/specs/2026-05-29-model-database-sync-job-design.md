# Model Database Sync Job Design

**Date:** 2026-05-29  
**Status:** Approved  
**Scope:** v1 — 异步批量同步数据库为模型

---

## 背景与目标

当前数据库管理页已经支持接管 database，也已有“单表导入为模型”和“单模型同步 schema”的能力，但缺少“按数据库批量同步”的入口。

当一个已接管数据库下已经存在大量表时，用户需要逐表导入，成本高；如果把整库同步塞进单次 GraphQL 请求，又容易在表数量较多时遇到超时、网关中断和结果不可追踪的问题。

### 目标

1. 在数据库管理页为每个已接管 database 提供“同步数据库”入口
2. 同步行为为异步任务，不阻塞用户当前页面操作
3. 对数据库中的每张表执行：
   - 若模型不存在：导入为新模型
   - 若模型已存在：同步该模型 schema
4. 单表失败不中断整体任务，最终返回完整汇总
5. 新建模型统一放入固定分组“数据库导入”

### 非目标

1. 不做 WebSocket / GraphQL Subscription 推送
2. 不做全局任务中心
3. 不处理视图、存储过程等非表对象
4. 不处理模型分组自定义选择

---

## Section 1：用户流程

### 发起同步

在 `/org/[orgName]/project/[projectSlug]/databases` 的数据库列表中，为每一行增加“同步数据库”操作。

点击后弹出确认框，明确说明：

1. 数据库中的新表会导入为模型
2. 已存在的模型会执行 schema 同步
3. 新建模型会进入固定分组“数据库导入”
4. 同步为后台任务，页面会轮询展示进度

用户确认后，前端调用 `startModelDatabaseSync`，服务端只创建任务并立即返回 `jobId`。

### 查看进度

前端拿到 `jobId` 后按固定间隔轮询 `modelDatabaseSyncJob(jobId)`。

列表行内展示当前状态：

| 状态 | 展示 |
|------|------|
| `PENDING` | 排队中 |
| `RUNNING` | 同步中 `processedTables/totalTables` |
| `SUCCEEDED` | 同步完成 |
| `PARTIAL_SUCCESS` | 部分成功 |
| `FAILED` | 同步失败 |

点击状态文案或结果入口，打开详情面板/弹窗，查看本次任务汇总。

### 结果展示

详情中至少展示：

1. 扫描表数 `totalTables`
2. 已处理表数 `processedTables`
3. 新建模型数 `createdModels`
4. 已同步模型数 `syncedModels`
5. 失败表数 `failedCount`
6. 失败明细列表（`tableName` + `message`）

---

## Section 2：后端 API

### GraphQL Schema

在 project schema 下新增异步任务相关类型与查询/变更：

```graphql
enum ModelDatabaseSyncJobStatus {
  PENDING
  RUNNING
  SUCCEEDED
  PARTIAL_SUCCESS
  FAILED
}

type ModelDatabaseSyncFailedTable {
  tableName: String!
  message: String!
}

type ModelDatabaseSyncJob {
  id: ID!
  databaseId: ID!
  status: ModelDatabaseSyncJobStatus!
  totalTables: Int!
  processedTables: Int!
  createdModels: Int!
  syncedModels: Int!
  failedCount: Int!
  failedTables: [ModelDatabaseSyncFailedTable!]!
  startedAt: Time
  finishedAt: Time
  createdAt: Time!
  updatedAt: Time!
}

type StartModelDatabaseSyncPayload {
  job: ModelDatabaseSyncJob!
}

extend type Mutation {
  startModelDatabaseSync(databaseId: ID!): StartModelDatabaseSyncPayload!
}

extend type Query {
  modelDatabaseSyncJob(jobId: ID!): ModelDatabaseSyncJob
}
```

### API 语义

`startModelDatabaseSync`：

1. 校验 `databaseId` 对应注册记录存在，且属于当前 `org + project`
2. 校验当前 database 没有处于 `PENDING/RUNNING` 的任务
3. 创建任务记录
4. 异步投递执行
5. 立即返回任务对象

`modelDatabaseSyncJob`：

1. 按 `jobId + orgName + projectSlug` 查询任务
2. 返回当前最新状态和汇总

---

## Section 3：任务数据模型

新增持久化表：`model_database_sync_job`

```sql
CREATE TABLE model_database_sync_job (
  id                VARCHAR(36)   NOT NULL,
  org_name          VARCHAR(64)   NOT NULL,
  project_slug      VARCHAR(64)   NOT NULL,
  database_id       VARCHAR(36)   NOT NULL,
  status            ENUM('pending','running','succeeded','partial_success','failed') NOT NULL,
  total_tables      INT           NOT NULL DEFAULT 0,
  processed_tables  INT           NOT NULL DEFAULT 0,
  created_models    INT           NOT NULL DEFAULT 0,
  synced_models     INT           NOT NULL DEFAULT 0,
  failed_count      INT           NOT NULL DEFAULT 0,
  failed_tables     JSON          NOT NULL,
  started_at        DATETIME(3)   NULL,
  finished_at       DATETIME(3)   NULL,
  created_at        DATETIME(3)   NOT NULL,
  updated_at        DATETIME(3)   NOT NULL,

  PRIMARY KEY (id),
  INDEX idx_project_database_created_at (org_name, project_slug, database_id, created_at),
  INDEX idx_project_status (org_name, project_slug, status)
);
```

### 字段说明

| 字段 | 说明 |
|------|------|
| `database_id` | 对应 `model_database.id` |
| `status` | 当前任务状态 |
| `failed_tables` | JSON 数组，元素格式 `{tableName, message}` |
| `started_at` | worker 真正开始处理的时间 |
| `finished_at` | 任务结束时间 |

### 并发约束

同一个 `(org_name, project_slug, database_id)` 在 `PENDING/RUNNING` 状态下只允许存在一个任务。

v1 不强依赖数据库唯一索引做状态互斥，先在应用层校验并在创建任务时落库；如果后续需要更强保证，再补充锁或唯一约束方案。

---

## Section 4：执行模型

### 执行入口

`startModelDatabaseSync` 完成任务创建后，后台通过现有异步执行机制启动一个 goroutine / worker 处理任务。

如果项目内已有统一异步任务基础设施，优先复用；如果没有，v1 可在 mutation 成功后通过 `bizutils.GoWithCtx` 启动后台执行，并在执行时重新构造最小上下文。

### 执行步骤

1. 把任务状态更新为 `RUNNING`，写入 `startedAt`
2. 根据 `databaseId` 取出注册数据库
3. 列出该数据库中的所有业务表，沿用现有导入模型使用的“表过滤规则”
4. 查询当前 project 下该 database 已存在的全部模型
5. 确保固定分组“数据库导入”存在，不存在则自动创建
6. 逐表处理：
   - 若表无对应模型：调用现有导入模型逻辑，创建到“数据库导入”
   - 若表已有对应模型：调用现有 schema 同步逻辑
   - 若失败：记录失败明细，`failedCount +1`，继续下一张表
   - 每处理一张表都更新 `processedTables`
7. 全部结束后收敛最终状态：
   - 无失败：`SUCCEEDED`
   - 有成功也有失败：`PARTIAL_SUCCESS`
   - 全部失败或初始化失败：`FAILED`
8. 写入 `finishedAt`

### 关键复用原则

不要重新发明“导入单表”和“同步单模型”的内部逻辑。批量任务只做编排，单表行为应复用当前已稳定的应用服务能力，避免两套 schema 映射规则分叉。

---

## Section 5：失败处理与日志

### 失败粒度

任务失败分两层：

1. **任务级失败**
   - 找不到 database
   - 无法建立数据库连接
   - 无法列出表
   - 任务初始化阶段发生不可继续错误

2. **单表级失败**
   - 单表导入失败
   - 单模型 schema 同步失败
   - 模型/字段转换失败

任务级失败会直接把整个任务置为 `FAILED`。  
单表级失败只写入 `failedTables`，整体继续执行。

### 日志要求

至少记录以下日志字段：

1. `job_id`
2. `database_id`
3. `database_name`
4. `table_name`
5. `action`：`import` 或 `sync_schema`
6. `status`
7. `error`

这样后续排查可以从任务维度和单表维度两条路径追踪。

---

## Section 6：前端页面变更

### 数据库列表页

目标页面：

```text
/org/[orgName]/project/[projectSlug]/databases
```

在每行数据库的更多菜单中新增：

1. `同步数据库`
2. `查看同步结果`（有最近任务时显示）

### 交互细节

1. 发起中：按钮 loading，避免重复点击
2. 轮询中：列表行显示当前任务状态和进度
3. 页面刷新后：如果列表项带有最近任务信息，前端可继续轮询未完成任务
4. 任务完成后：自动停止轮询

### 数据获取建议

v1 建议数据库列表查询补充“最近一次同步任务摘要”，避免页面刷新后丢失状态入口。若当前列表 API 不适合扩展，则前端可在发起任务后将 `jobId` 保存在本页状态中，先完成最小闭环。

推荐优先级：

1. 最小版：发起后本页内轮询当前 `jobId`
2. 增强版：列表接口补最近任务摘要，实现刷新可恢复

v1 允许先落最小版。

---

## Section 7：边界与约束

1. 只处理已接管的 database
2. 只处理普通表，不处理系统库和非表对象
3. 已有模型不重建，只做 schema 同步
4. 新建模型固定进入“数据库导入”分组
5. 同一 database 同时只允许一个运行中的同步任务
6. 不保证任务执行过程中的数据库强一致快照，按“扫描时刻到处理时刻的最终可见状态”执行

---

## Section 8：测试策略

### 后端

1. 应用层单测
   - 新表导入
   - 已有模型走 schema 同步
   - 单表失败不中断
   - 全量失败 / 部分成功 / 全量成功状态收敛
   - 并发重复发起被拒绝

2. Repository 单测
   - 创建任务
   - 更新任务进度
   - 查询任务详情
   - 运行中任务存在性检查

3. GraphQL 层测试
   - `startModelDatabaseSync`
   - `modelDatabaseSyncJob`

### 前端

1. 数据库列表行发起同步
2. 轮询状态更新
3. 成功 / 部分成功 / 失败结果展示
4. 运行中禁止重复发起

---

## 文件变更清单

### 后端

| 文件 | 操作 |
|------|------|
| `api/graph/project/schema/database.graphql` | 修改，新增 sync job schema |
| `internal/domain/modeldatabase/` | 修改或新增任务实体、仓储接口 |
| `internal/app/modeldatabase/` | 新增启动任务与执行任务用例 |
| `internal/infrastructure/repository/` | 新增 sync job repository |
| `internal/interfaces/graphql/project/` | 新增 mutation/query resolver |
| `db/schema/mysql/*.sql` | 新增任务表 schema |
| `db/queries/*.sql` | 新增任务查询与更新 SQL |

### 前端

| 文件 | 操作 |
|------|------|
| `src/api-client/project/model-database-graphql-docs.ts` | 新增 start/query job 文档 |
| `src/web/hooks/model-database/use-model-databases.ts` | 新增发起同步与查询 job hooks |
| `src/app/org/[orgName]/project/[projectSlug]/databases/page.tsx` | 修改，接入任务状态 |
| `src/app/org/[orgName]/project/[projectSlug]/databases/_components/*` | 修改，增加菜单、确认框、结果展示 |

---

## 实施顺序建议

1. 后端任务表与 GraphQL contract
2. 后端任务创建/查询/执行链路
3. 前端发起同步与轮询
4. 前端结果展示与交互打磨

