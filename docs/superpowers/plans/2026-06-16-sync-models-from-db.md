# syncModelsFromDB Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增 `syncModelsFromDB` Mutation（异步 Job，upsert 语义）和 `modelSyncJob` Query，替代 `importModel`。

**Architecture:** 在现有 `modeldatabase` 包下新增 `ModelSyncJob` domain + `ModelSyncAppService`，复用 `ReverseEngineerAppService.ImportModel` 和 `ModelDesignAppService` 的字段 CRUD 能力，通过 `backgroundRunner` 异步执行每张表的 introspect → upsert 流程。字段同步范围以 `StorageHint != nil` 为准，强制覆盖已有字段属性。

**Tech Stack:** Go 1.23, sqlc, gqlgen, GraphQL (Project schema), MySQL 8

---

## File Map

### 新建文件

| 文件 | 职责 |
|------|------|
| `modelcraft-backend/internal/domain/modeldatabase/model_sync_job.go` | `ModelSyncJob` domain entity + status 常量 + repository 接口 |
| `modelcraft-backend/internal/app/modeldatabase/sync_models_app.go` | `SyncModelsAppService`：`StartSync` + `GetJob` + `RunSyncJob` + `processTable` |
| `modelcraft-backend/internal/app/modeldatabase/sync_models_app_test.go` | 应用层单测 |
| `modelcraft-backend/internal/infrastructure/repository/sql_model_sync_job_repository.go` | `SqlModelSyncJobRepository` 实现 |
| `modelcraft-backend/db/schema/mysql/18_model_sync_job.sql` | `model_sync_job` 表 DDL |
| `modelcraft-backend/db/queries/model_sync_job.sql` | sqlc 查询 |

### 修改文件

| 文件 | 变更 |
|------|------|
| `modelcraft-backend/db/schema/mysql/03_model_domain.sql` | ✅ 已添加 `storage_hint` 列（已提交） |
| `modelcraft-backend/db/queries/field.sql` | ✅ 已更新 Create/Update（已提交） |
| `modelcraft-backend/internal/infrastructure/repository/sql_modeldesign_repository.go` | `FieldDefinitionToCreateParams` 和 `FieldDefinitionToUpdateParams` 加 `StorageHint` 字段 |
| `modelcraft-backend/api/graph/project/schema/model.graphql` | 新增 `syncModelsFromDB` mutation + `ModelSyncJob` 类型 + `modelSyncJob` query |
| `modelcraft-backend/internal/interfaces/graphql/project/resolver.go` | 加 `SyncModelsAppService *modeldatabase.SyncModelsAppService` 字段 |
| `modelcraft-backend/internal/interfaces/graphql/project/model.resolvers.go` | 新增 `SyncModelsFromDB` + `ModelSyncJob` resolver |
| `modelcraft-backend/internal/app/modeldatabase/sync_models_app.go` | （新建，见上） |

---

## Task 1：sqlc 生成 storage_hint

`db/schema` 和 `db/queries` 的 `field.sql` 已经更新（上一次提交），需要重新运行 sqlc 让 `dbgen` 感知 `storage_hint`，再更新 repository 的 convert 函数。

**Files:**
- Regenerate: `modelcraft-backend/internal/infrastructure/dbgen/` (auto-generated)
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_modeldesign_repository.go:195-246`

- [ ] **Step 1：运行 sqlc 重新生成**

```bash
cd modelcraft-backend && just generate-sqlc
```

预期：`internal/infrastructure/dbgen/` 下的 `CreateFieldDefinitionParams` 和 `UpdateFieldParams` 新增 `StorageHint sql.NullString` 字段，编译失败（因为 convert 函数还没更新）。

- [ ] **Step 2：更新 `FieldDefinitionToCreateParams`**

在 `sql_modeldesign_repository.go` 的 `FieldDefinitionToCreateParams` 返回值里加一行：

```go
return dbgen.CreateFieldDefinitionParams{
    // ... existing fields ...
    BelongsToFkID: sqlerr.PtrToNullStr(fd.BelongsToFKID),
    StorageHint:   sqlerr.PtrToNullStr(fd.StorageHint), // 新增
}, nil
```

- [ ] **Step 3：更新 `FieldDefinitionToUpdateParams`**

```go
return dbgen.UpdateFieldParams{
    // ... existing fields ...
    Metadata:    ptrJSON(metadataJSON),
    StorageHint: sqlerr.PtrToNullStr(fd.StorageHint), // 新增
    ModelID:     fd.ModelID,
    Name:        fd.Name,
}, nil
```

- [ ] **Step 4：更新 `rowToFieldDefinition`（db → domain 的转换）**

在 `rowToFieldDefinition` 函数的 return 里加：

```go
StorageHint: sqlerr.NullStrToPtr(row.StorageHint),
```

- [ ] **Step 5：编译验证**

```bash
cd modelcraft-backend && go build ./...
```

预期：编译通过，无错误。

- [ ] **Step 6：运行已有 repository convert 测试**

```bash
cd modelcraft-backend && go test ./internal/infrastructure/repository/... -v -run TestFieldDefinition
```

预期：PASS（已有测试加 `StorageHint: nil` 依然 pass，因为是可选字段）。

- [ ] **Step 7：commit**

```bash
cd modelcraft-backend && git add internal/infrastructure/repository/sql_modeldesign_repository.go internal/infrastructure/dbgen/
git commit -m "feat: wire storage_hint through sqlc and repository convert"
```

---

## Task 2：DB 表 + sqlc 查询（model_sync_job）

**Files:**
- Create: `modelcraft-backend/db/schema/mysql/18_model_sync_job.sql`
- Create: `modelcraft-backend/db/queries/model_sync_job.sql`
- Regenerate: `modelcraft-backend/internal/infrastructure/dbgen/`

- [ ] **Step 1：创建表定义文件**

新建 `modelcraft-backend/db/schema/mysql/18_model_sync_job.sql`：

```sql
-- model_sync_job 表 SQL 定义
-- syncModelsFromDB 异步任务表（以 databaseName 为维度，独立于 model_database_sync_job）

CREATE TABLE IF NOT EXISTS `model_sync_job` (
  `id`               VARCHAR(36)  NOT NULL COMMENT '任务唯一标识符',
  `org_name`         VARCHAR(64)  NOT NULL COMMENT '所属组织名称',
  `project_slug`     VARCHAR(64)  NOT NULL COMMENT '所属项目标识符',
  `database_name`    VARCHAR(128) NOT NULL COMMENT '目标数据库名称',
  `table_names`      JSON         NOT NULL COMMENT '指定同步的表名列表，空数组表示全量',
  `status`           ENUM('pending','running','succeeded','partial_success','failed') NOT NULL COMMENT '任务状态',
  `total_tables`     INT          NOT NULL DEFAULT 0 COMMENT '扫描到的总表数',
  `processed_tables` INT          NOT NULL DEFAULT 0 COMMENT '已处理表数',
  `created_models`   INT          NOT NULL DEFAULT 0 COMMENT '新建模型数',
  `synced_models`    INT          NOT NULL DEFAULT 0 COMMENT '已同步模型数',
  `failed_count`     INT          NOT NULL DEFAULT 0 COMMENT '失败表数',
  `failed_tables`    JSON         NOT NULL COMMENT '失败明细，格式：[{tableName, message}]',
  `started_at`       DATETIME(3)  NULL     COMMENT 'worker 开始时间',
  `finished_at`      DATETIME(3)  NULL     COMMENT '任务结束时间',
  `created_at`       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at`       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  PRIMARY KEY (`id`),
  INDEX `idx_model_sync_job_project_db` (`org_name`, `project_slug`, `database_name`, `created_at`),
  INDEX `idx_model_sync_job_status`     (`org_name`, `project_slug`, `status`)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='syncModelsFromDB 异步任务表';
```

- [ ] **Step 2：创建 sqlc 查询文件**

新建 `modelcraft-backend/db/queries/model_sync_job.sql`：

```sql
-- name: CreateModelSyncJob :exec
INSERT INTO model_sync_job (
  id, org_name, project_slug, database_name, table_names,
  status, total_tables, processed_tables, created_models, synced_models,
  failed_count, failed_tables, started_at, finished_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetModelSyncJobByID :one
SELECT * FROM model_sync_job
WHERE id = ? AND org_name = ? AND project_slug = ?
LIMIT 1;

-- name: GetActiveModelSyncJobByDatabase :one
SELECT * FROM model_sync_job
WHERE org_name = ?
  AND project_slug = ?
  AND database_name = ?
  AND status IN ('pending', 'running')
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateModelSyncJob :exec
UPDATE model_sync_job
SET status = ?,
    total_tables = ?,
    processed_tables = ?,
    created_models = ?,
    synced_models = ?,
    failed_count = ?,
    failed_tables = ?,
    started_at = ?,
    finished_at = ?,
    updated_at = NOW(3)
WHERE id = ?;
```

- [ ] **Step 3：重新运行 sqlc**

```bash
cd modelcraft-backend && just generate-sqlc
```

预期：`internal/infrastructure/dbgen/` 下生成 `model_sync_job.sql.go`，包含 `CreateModelSyncJob`、`GetModelSyncJobByID`、`GetActiveModelSyncJobByDatabase`、`UpdateModelSyncJob`。

- [ ] **Step 4：编译验证**

```bash
cd modelcraft-backend && go build ./...
```

预期：编译通过。

- [ ] **Step 5：commit**

```bash
cd modelcraft-backend && git add db/schema/mysql/18_model_sync_job.sql db/queries/model_sync_job.sql internal/infrastructure/dbgen/
git commit -m "feat: add model_sync_job table and sqlc queries"
```

---

## Task 3：Domain Entity + Repository Interface

**Files:**
- Create: `modelcraft-backend/internal/domain/modeldatabase/model_sync_job.go`

- [ ] **Step 1：创建 domain entity 文件**

新建 `modelcraft-backend/internal/domain/modeldatabase/model_sync_job.go`：

```go
package modeldatabase

import "time"

// ModelSyncJobStatus 同步任务状态
type ModelSyncJobStatus string

const (
	ModelSyncJobStatusPending        ModelSyncJobStatus = "pending"
	ModelSyncJobStatusRunning        ModelSyncJobStatus = "running"
	ModelSyncJobStatusSucceeded      ModelSyncJobStatus = "succeeded"
	ModelSyncJobStatusPartialSuccess ModelSyncJobStatus = "partial_success"
	ModelSyncJobStatusFailed         ModelSyncJobStatus = "failed"
)

// ModelSyncFailedTable 单表失败明细
type ModelSyncFailedTable struct {
	TableName string `json:"tableName"`
	Message   string `json:"message"`
}

// ModelSyncJob 同步任务实体
type ModelSyncJob struct {
	ID              string
	OrgName         string
	ProjectSlug     string
	DatabaseName    string
	TableNames      []string // 空 = 全量
	Status          ModelSyncJobStatus
	TotalTables     int
	ProcessedTables int
	CreatedModels   int
	SyncedModels    int
	FailedCount     int
	FailedTables    []ModelSyncFailedTable
	StartedAt       *time.Time
	FinishedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ModelSyncJobRepository 同步任务仓储接口
type ModelSyncJobRepository interface {
	Create(ctx interface{ Deadline() (interface{}, bool) }, job *ModelSyncJob) error
	GetByID(ctx interface{ Deadline() (interface{}, bool) }, orgName, projectSlug, jobID string) (*ModelSyncJob, error)
	// GetActiveByDatabase 返回同一 database 下处于 pending/running 状态的任务（如有）
	GetActiveByDatabase(ctx interface{ Deadline() (interface{}, bool) }, orgName, projectSlug, databaseName string) (*ModelSyncJob, error)
	Update(ctx interface{ Deadline() (interface{}, bool) }, job *ModelSyncJob) error
}
```

**注意**：Go 接口参数用 `context.Context`，上面只是示意。实际写法参照同包的 `repository.go`：

```go
package modeldatabase

import (
	"context"
	"time"
)

type ModelSyncJobStatus string

const (
	ModelSyncJobStatusPending        ModelSyncJobStatus = "pending"
	ModelSyncJobStatusRunning        ModelSyncJobStatus = "running"
	ModelSyncJobStatusSucceeded      ModelSyncJobStatus = "succeeded"
	ModelSyncJobStatusPartialSuccess ModelSyncJobStatus = "partial_success"
	ModelSyncJobStatusFailed         ModelSyncJobStatus = "failed"
)

type ModelSyncFailedTable struct {
	TableName string `json:"tableName"`
	Message   string `json:"message"`
}

type ModelSyncJob struct {
	ID              string
	OrgName         string
	ProjectSlug     string
	DatabaseName    string
	TableNames      []string
	Status          ModelSyncJobStatus
	TotalTables     int
	ProcessedTables int
	CreatedModels   int
	SyncedModels    int
	FailedCount     int
	FailedTables    []ModelSyncFailedTable
	StartedAt       *time.Time
	FinishedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ModelSyncJobRepository interface {
	Create(ctx context.Context, job *ModelSyncJob) error
	GetByID(ctx context.Context, orgName, projectSlug, jobID string) (*ModelSyncJob, error)
	GetActiveByDatabase(ctx context.Context, orgName, projectSlug, databaseName string) (*ModelSyncJob, error)
	Update(ctx context.Context, job *ModelSyncJob) error
}
```

- [ ] **Step 2：编译验证**

```bash
cd modelcraft-backend && go build ./internal/domain/modeldatabase/...
```

预期：编译通过。

- [ ] **Step 3：commit**

```bash
cd modelcraft-backend && git add internal/domain/modeldatabase/model_sync_job.go
git commit -m "feat: add ModelSyncJob domain entity and repository interface"
```

---

## Task 4：Repository 实现

**Files:**
- Create: `modelcraft-backend/internal/infrastructure/repository/sql_model_sync_job_repository.go`

- [ ] **Step 1：创建 repository 实现文件**

新建 `modelcraft-backend/internal/infrastructure/repository/sql_model_sync_job_repository.go`：

```go
package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	domaindb "modelcraft/internal/domain/modeldatabase"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/dbgenwrap"
	"modelcraft/internal/infrastructure/sqlerr"
)

type SqlModelSyncJobRepository struct {
	q dbgen.Querier
}

func NewSqlModelSyncJobRepository(q dbgen.Querier) domaindb.ModelSyncJobRepository {
	return &SqlModelSyncJobRepository{q: dbgenwrap.NewSafeQuerier(q)}
}

func (r *SqlModelSyncJobRepository) Create(ctx context.Context, job *domaindb.ModelSyncJob) error {
	tableNames, err := marshalModelSyncTableNames(job.TableNames)
	if err != nil {
		return err
	}
	failedTables, err := marshalModelSyncFailedTables(job.FailedTables)
	if err != nil {
		return err
	}
	return r.q.CreateModelSyncJob(ctx, dbgen.CreateModelSyncJobParams{
		ID:              job.ID,
		OrgName:         job.OrgName,
		ProjectSlug:     job.ProjectSlug,
		DatabaseName:    job.DatabaseName,
		TableNames:      tableNames,
		Status:          dbgen.ModelSyncJobStatus(job.Status),
		TotalTables:     int32(job.TotalTables),
		ProcessedTables: int32(job.ProcessedTables),
		CreatedModels:   int32(job.CreatedModels),
		SyncedModels:    int32(job.SyncedModels),
		FailedCount:     int32(job.FailedCount),
		FailedTables:    failedTables,
		StartedAt:       timeToNull(job.StartedAt),
		FinishedAt:      timeToNull(job.FinishedAt),
	})
}

func (r *SqlModelSyncJobRepository) GetByID(
	ctx context.Context, orgName, projectSlug, jobID string,
) (*domaindb.ModelSyncJob, error) {
	row, err := r.q.GetModelSyncJobByID(ctx, dbgen.GetModelSyncJobByIDParams{
		ID:          jobID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, nil //nolint:nilnil
		}
		return nil, err
	}
	return modelSyncJobToDomain(row)
}

func (r *SqlModelSyncJobRepository) GetActiveByDatabase(
	ctx context.Context, orgName, projectSlug, databaseName string,
) (*domaindb.ModelSyncJob, error) {
	row, err := r.q.GetActiveModelSyncJobByDatabase(ctx, dbgen.GetActiveModelSyncJobByDatabaseParams{
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		DatabaseName: databaseName,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, nil //nolint:nilnil
		}
		return nil, err
	}
	return modelSyncJobToDomain(row)
}

func (r *SqlModelSyncJobRepository) Update(ctx context.Context, job *domaindb.ModelSyncJob) error {
	failedTables, err := marshalModelSyncFailedTables(job.FailedTables)
	if err != nil {
		return err
	}
	return r.q.UpdateModelSyncJob(ctx, dbgen.UpdateModelSyncJobParams{
		Status:          dbgen.ModelSyncJobStatus(job.Status),
		TotalTables:     int32(job.TotalTables),
		ProcessedTables: int32(job.ProcessedTables),
		CreatedModels:   int32(job.CreatedModels),
		SyncedModels:    int32(job.SyncedModels),
		FailedCount:     int32(job.FailedCount),
		FailedTables:    failedTables,
		StartedAt:       timeToNull(job.StartedAt),
		FinishedAt:      timeToNull(job.FinishedAt),
		ID:              job.ID,
	})
}

func modelSyncJobToDomain(row dbgen.ModelSyncJob) (*domaindb.ModelSyncJob, error) {
	tableNames, err := unmarshalModelSyncTableNames(row.TableNames)
	if err != nil {
		return nil, err
	}
	failedTables, err := unmarshalModelSyncFailedTables(row.FailedTables)
	if err != nil {
		return nil, err
	}
	return &domaindb.ModelSyncJob{
		ID:              row.ID,
		OrgName:         row.OrgName,
		ProjectSlug:     row.ProjectSlug,
		DatabaseName:    row.DatabaseName,
		TableNames:      tableNames,
		Status:          domaindb.ModelSyncJobStatus(row.Status),
		TotalTables:     int(row.TotalTables),
		ProcessedTables: int(row.ProcessedTables),
		CreatedModels:   int(row.CreatedModels),
		SyncedModels:    int(row.SyncedModels),
		FailedCount:     int(row.FailedCount),
		FailedTables:    failedTables,
		StartedAt:       nullToTimePtr(row.StartedAt),
		FinishedAt:      nullToTimePtr(row.FinishedAt),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}, nil
}

func marshalModelSyncTableNames(names []string) (json.RawMessage, error) {
	if names == nil {
		names = []string{}
	}
	data, err := json.Marshal(names)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func unmarshalModelSyncTableNames(data json.RawMessage) ([]string, error) {
	if len(data) == 0 {
		return []string{}, nil
	}
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return nil, err
	}
	return names, nil
}

func marshalModelSyncFailedTables(items []domaindb.ModelSyncFailedTable) (json.RawMessage, error) {
	if items == nil {
		items = []domaindb.ModelSyncFailedTable{}
	}
	data, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

func unmarshalModelSyncFailedTables(data json.RawMessage) ([]domaindb.ModelSyncFailedTable, error) {
	if len(data) == 0 {
		return []domaindb.ModelSyncFailedTable{}, nil
	}
	var items []domaindb.ModelSyncFailedTable
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// 确保接口实现完整
var _ domaindb.ModelSyncJobRepository = (*SqlModelSyncJobRepository)(nil)
```

**注意**：`timeToNull` 和 `nullToTimePtr` 已在 `sql_model_database_sync_job_repository.go` 中定义，本文件直接复用同包函数，无需重复定义。

- [ ] **Step 2：编译验证**

```bash
cd modelcraft-backend && go build ./internal/infrastructure/repository/...
```

预期：编译通过（如果 dbgen 里 `ModelSyncJobStatus` 类型还不存在，可能报错 — 等 Task 2 的 sqlc 生成后才能通过）。

- [ ] **Step 3：commit**

```bash
cd modelcraft-backend && git add internal/infrastructure/repository/sql_model_sync_job_repository.go
git commit -m "feat: add SqlModelSyncJobRepository implementation"
```

---

## Task 5：App Service（SyncModelsAppService）

**Files:**
- Create: `modelcraft-backend/internal/app/modeldatabase/sync_models_app.go`

参考 `sync_job_app.go` 的结构，但差异点：
1. 以 `databaseName` 为维度（不是 `databaseId`）
2. 支持 `tableNames` 过滤
3. `processTable` 中对已有 model 执行字段 full sync（只处理 `StorageHint != nil` 的字段，强制覆盖）

- [ ] **Step 1：创建 App Service 文件**

新建 `modelcraft-backend/internal/app/modeldatabase/sync_models_app.go`：

```go
package modeldatabase

import (
	"context"
	"time"

	"github.com/google/uuid"

	domaindb "modelcraft/internal/domain/modeldatabase"
	domainmodel "modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"

	appmodeldesign "modelcraft/internal/app/modeldesign"
)

// SyncModelsFromDBCommand 触发同步任务的命令
type SyncModelsFromDBCommand struct {
	DatabaseName string
	TableNames   []string // nil 或空 = 全量（配合 SyncAll 使用）
	SyncAll      bool
}

// SyncModelsAppService 处理 syncModelsFromDB 的异步任务
type SyncModelsAppService struct {
	syncJobRepo     domaindb.ModelSyncJobRepository
	reverseEngineer syncModelsReverseEngineer
	modelRepo       domainmodel.ModelRepository
	fieldSyncer     syncModelsFieldSyncer
	runner          modelDatabaseSyncRunner // 复用同包接口
	now             func() time.Time
}

// syncModelsReverseEngineer 反向工程接口（子集）
type syncModelsReverseEngineer interface {
	ListTables(
		ctx context.Context,
		orgName, projectSlug, databaseName string,
		excludeExisting bool,
		limit, offset int,
	) (*appmodeldesign.ListTablesResult, error)
	ImportModel(ctx context.Context, cmd appmodeldesign.ImportModelCommand) (*appmodeldesign.ImportModelResult, error)
	GetTableDefinition(ctx context.Context, orgName, projectSlug, databaseName, tableName string) (*appmodeldesign.TableDefinitionResult, error)
}

// syncModelsFieldSyncer 字段同步接口（按 storageHint 全量同步）
type syncModelsFieldSyncer interface {
	SyncFieldsFromDB(
		ctx context.Context,
		modelID string,
		dbFields []*domainmodel.FieldDefinition,
	) error
}

type SyncModelsAppServiceDeps struct {
	SyncJobRepo     domaindb.ModelSyncJobRepository
	ReverseEngineer syncModelsReverseEngineer
	ModelRepo       domainmodel.ModelRepository
	FieldSyncer     syncModelsFieldSyncer
	Runner          modelDatabaseSyncRunner
	Now             func() time.Time
}

func NewSyncModelsAppService(deps SyncModelsAppServiceDeps) *SyncModelsAppService {
	runner := deps.Runner
	if runner == nil {
		runner = backgroundRunner{}
	}
	nowFn := deps.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	return &SyncModelsAppService{
		syncJobRepo:     deps.SyncJobRepo,
		reverseEngineer: deps.ReverseEngineer,
		modelRepo:       deps.ModelRepo,
		fieldSyncer:     deps.FieldSyncer,
		runner:          runner,
		now:             nowFn,
	}
}

// StartSync 参数校验 → 并发检查 → 创建 job → 异步执行 → 返回 jobID
func (s *SyncModelsAppService) StartSync(
	ctx context.Context,
	cmd SyncModelsFromDBCommand,
) (*domaindb.ModelSyncJob, error) {
	// 参数校验
	if err := validateSyncModelsCommand(cmd); err != nil {
		return nil, err
	}

	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}

	// 并发检查：同一 database 已有 active job → 报错
	active, err := s.syncJobRepo.GetActiveByDatabase(ctx, orgName, projectSlug, cmd.DatabaseName)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, bizerrors.NewError(bizerrors.Conflict, "sync job already running for database "+cmd.DatabaseName)
	}

	now := s.now()
	tableNames := cmd.TableNames
	if tableNames == nil {
		tableNames = []string{}
	}
	job := &domaindb.ModelSyncJob{
		ID:           uuid.NewString(),
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		DatabaseName: cmd.DatabaseName,
		TableNames:   tableNames,
		Status:       domaindb.ModelSyncJobStatusPending,
		FailedTables: []domaindb.ModelSyncFailedTable{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.syncJobRepo.Create(ctx, job); err != nil {
		return nil, err
	}

	s.runner.Go(ctx, func(runCtx context.Context) {
		if err := s.RunSyncJob(runCtx, job.ID); err != nil {
			logfacade.GetLogger(runCtx).Error(runCtx, "model sync job failed",
				logfacade.String("job_id", job.ID),
				logfacade.Err(err),
			)
		}
	})
	return job, nil
}

// GetJob 查询任务状态
func (s *SyncModelsAppService) GetJob(
	ctx context.Context, jobID string,
) (*domaindb.ModelSyncJob, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	return s.syncJobRepo.GetByID(ctx, orgName, projectSlug, jobID)
}

// RunSyncJob 后台执行同步任务（由 runner.Go 调用）
func (s *SyncModelsAppService) RunSyncJob(ctx context.Context, jobID string) error {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	logger := logfacade.GetLogger(ctx)

	job, err := s.syncJobRepo.GetByID(ctx, orgName, projectSlug, jobID)
	if err != nil {
		return err
	}

	// 标记 RUNNING
	now := s.now()
	job.Status = domaindb.ModelSyncJobStatusRunning
	job.StartedAt = &now
	job.UpdatedAt = now
	if err := s.syncJobRepo.Update(ctx, job); err != nil {
		return err
	}

	// 获取表列表
	var tableNames []string
	if len(job.TableNames) > 0 {
		tableNames = job.TableNames
	} else {
		// 全量：从 DB 列举
		tableResult, err := s.reverseEngineer.ListTables(ctx, orgName, projectSlug, job.DatabaseName, false, 0, 0)
		if err != nil {
			logger.Error(ctx, "model sync job: ListTables failed",
				logfacade.String("job_id", jobID), logfacade.Err(err))
			return s.failJob(ctx, job, err)
		}
		tableNames = tableResult.Tables
	}

	job.TotalTables = len(tableNames)
	job.UpdatedAt = s.now()
	if err := s.syncJobRepo.Update(ctx, job); err != nil {
		return err
	}

	// 逐表处理
	for _, tableName := range tableNames {
		if err := s.processTable(ctx, job, tableName); err != nil {
			logger.Error(ctx, "model sync job: processTable fatal",
				logfacade.String("job_id", jobID), logfacade.String("table", tableName), logfacade.Err(err))
			return err
		}
	}

	// 收敛状态
	finishedAt := s.now()
	job.FinishedAt = &finishedAt
	job.UpdatedAt = finishedAt
	switch {
	case job.FailedCount == 0:
		job.Status = domaindb.ModelSyncJobStatusSucceeded
	case job.CreatedModels > 0 || job.SyncedModels > 0:
		job.Status = domaindb.ModelSyncJobStatusPartialSuccess
	default:
		job.Status = domaindb.ModelSyncJobStatusFailed
	}
	return s.syncJobRepo.Update(ctx, job)
}

func (s *SyncModelsAppService) processTable(
	ctx context.Context,
	job *domaindb.ModelSyncJob,
	tableName string,
) error {
	orgName := job.OrgName
	projectSlug := job.ProjectSlug
	databaseName := job.DatabaseName

	// 1. 获取 DB 表定义（需 ReverseEngineerAppService 暴露 GetTableDefinition）
	tableDef, err := s.reverseEngineer.GetTableDefinition(ctx, orgName, projectSlug, databaseName, tableName)
	if err != nil {
		return s.recordTableFailure(ctx, job, tableName, err)
	}

	modelName := appmodeldesign.NormalizeModelName(tableName)

	// 2. 查询 model 是否已存在
	existingModel, err := s.modelRepo.GetByName(ctx, orgName, databaseName, modelName, projectSlug)
	if err != nil && !shared.IsNotFoundError(err) {
		return s.recordTableFailure(ctx, job, tableName, err)
	}

	if existingModel != nil {
		// 3a. model 已存在：full sync storageHint 非空字段
		if err := s.fieldSyncer.SyncFieldsFromDB(ctx, existingModel.ID, tableDef.Fields); err != nil {
			return s.recordTableFailure(ctx, job, tableName, err)
		}
		job.SyncedModels++
	} else {
		// 3b. model 不存在：创建新模型
		_, importErr := s.reverseEngineer.ImportModel(ctx, appmodeldesign.ImportModelCommand{
			OrgName:      orgName,
			ProjectSlug:  projectSlug,
			DatabaseName: databaseName,
			TableName:    tableName,
		})
		if importErr != nil {
			return s.recordTableFailure(ctx, job, tableName, importErr)
		}
		job.CreatedModels++
	}

	job.ProcessedTables++
	job.UpdatedAt = s.now()
	return s.syncJobRepo.Update(ctx, job)
}

func (s *SyncModelsAppService) recordTableFailure(
	ctx context.Context,
	job *domaindb.ModelSyncJob,
	tableName string,
	err error,
) error {
	job.ProcessedTables++
	job.FailedCount++
	job.FailedTables = append(job.FailedTables, domaindb.ModelSyncFailedTable{
		TableName: tableName,
		Message:   err.Error(),
	})
	job.UpdatedAt = s.now()
	return s.syncJobRepo.Update(ctx, job)
}

func (s *SyncModelsAppService) failJob(
	ctx context.Context,
	job *domaindb.ModelSyncJob,
	err error,
) error {
	logfacade.GetLogger(ctx).Error(ctx, "model sync job failed",
		logfacade.String("job_id", job.ID),
		logfacade.String("database_name", job.DatabaseName),
		logfacade.Err(err),
	)
	now := s.now()
	job.Status = domaindb.ModelSyncJobStatusFailed
	job.FinishedAt = &now
	job.UpdatedAt = now
	if updateErr := s.syncJobRepo.Update(ctx, job); updateErr != nil {
		return updateErr
	}
	return err
}

// validateSyncModelsCommand 参数互斥校验
func validateSyncModelsCommand(cmd SyncModelsFromDBCommand) error {
	hasTableNames := len(cmd.TableNames) > 0
	hasSyncAll := cmd.SyncAll

	if hasTableNames && hasSyncAll {
		return bizerrors.NewError(bizerrors.ParamInvalid, "cannot specify both tableNames and syncAll")
	}
	if !hasTableNames && !hasSyncAll {
		return bizerrors.NewError(bizerrors.ParamInvalid, "must specify either tableNames or syncAll=true")
	}
	return nil
}

// 确保 bizutils 被引用（backgroundRunner 在同包已定义）
var _ = bizutils.GoWithCtx
```

**注意：**
- `GetTableDefinition`、`NormalizeModelName`、`TableDefinitionResult` 是新增的，需要在 Task 6 暴露到 `ReverseEngineerAppService`。
- `SyncFieldsFromDB` 是新增的字段同步方法，需要在 Task 6 实现到 `ModelDesignAppService`。

- [ ] **Step 2：编译（先不求 pass，确认缺少的符号）**

```bash
cd modelcraft-backend && go build ./internal/app/modeldatabase/...
```

预期：报错提示 `GetTableDefinition` 和 `SyncFieldsFromDB` 未定义 — 这是正常的，留给 Task 6 解决。

- [ ] **Step 3：commit（记录 WIP）**

```bash
cd modelcraft-backend && git add internal/app/modeldatabase/sync_models_app.go
git commit -m "feat: add SyncModelsAppService (wip, pending GetTableDefinition+SyncFieldsFromDB)"
```

---

## Task 6：暴露 GetTableDefinition + 实现 SyncFieldsFromDB

**Files:**
- Modify: `modelcraft-backend/internal/app/modeldesign/reverse_engineer_app.go`
- Modify: `modelcraft-backend/internal/app/modeldesign/model_app.go`

### 6A：在 ReverseEngineerAppService 暴露 GetTableDefinition

当前 `getTableDefinition` 是私有方法。需要：
1. 新增公开类型 `TableDefinitionResult`（包含按 DB 列构建的 `[]*domainmodel.FieldDefinition`）
2. 新增公开方法 `GetTableDefinition`
3. 新增公开函数 `NormalizeModelName`（当前已私有，需提升）

- [ ] **Step 1：在 `reverse_engineer_app.go` 末尾添加**

```go
// TableDefinitionResult GetTableDefinition 的返回类型
type TableDefinitionResult struct {
	TableName string
	// Fields 按 DB 列构建的 FieldDefinition 列表（StorageHint 已填充为 DB 列名）
	// 仅包含 ReverseTypeMapper 支持的列，不可映射的列被跳过。
	Fields []*domainmodel.FieldDefinition
}

// GetTableDefinition 从数据库内省获取表定义，并转换为 FieldDefinition 列表。
// 供 SyncModelsAppService 使用。
func (s *ReverseEngineerAppService) GetTableDefinition(
	ctx context.Context,
	orgName, projectSlug, databaseName, tableName string,
) (*TableDefinitionResult, error) {
	cmd := ImportModelCommand{
		OrgName:      orgName,
		ProjectSlug:  projectSlug,
		DatabaseName: databaseName,
		TableName:    tableName,
	}
	if err := s.validateCommand(cmd); err != nil {
		return nil, err
	}
	tableDef, err := s.getTableDefinition(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// 确定 isReadOnly
	isReadOnly := false
	if s.modelDatabaseRepo != nil {
		db, err := s.modelDatabaseRepo.GetByName(ctx, orgName, projectSlug, databaseName)
		if err == nil && db != nil && db.Mode == modeldatabase.DatabaseModeManaged {
			isReadOnly = true
		}
	}

	buildResult, err := s.buildModelFromTable(tableDef, cmd, isReadOnly)
	if err != nil {
		return nil, err
	}
	return &TableDefinitionResult{
		TableName: tableName,
		Fields:    buildResult.Model.Fields,
	}, nil
}
```

- [ ] **Step 2：在 `model_builder.go` 将 `normalizeModelName` 提升为公开函数**

```go
// NormalizeModelName 规范化模型名称（转为小写）— 公开版，供跨包调用
func NormalizeModelName(tableName string) string {
	return normalizeModelName(tableName)
}
```

添加到 `model_builder.go` 末尾即可。

### 6B：在 ModelDesignAppService 实现 SyncFieldsFromDB

字段同步规则：
- 只处理 `StorageHint != nil` 的 `dbFields`
- DB 有对应 storageHint，model 已有该字段 → 强制覆盖字段属性
- DB 有但 model 无 → 新增字段
- model 有（StorageHint 非空）但 DB 无 → 删除字段
- model 中 `StorageHint == nil` 的字段不碰

- [ ] **Step 3：在 `model_app.go` 末尾添加 `SyncFieldsFromDB`**

```go
// SyncFieldsFromDB 按 DB 列定义对模型字段进行 full sync。
// 只处理 StorageHint 非空的字段：
//   - DB 有对应 storageHint，model 已有 → 强制覆盖属性
//   - DB 有，model 无 → 新增
//   - model 有（StorageHint 非空），DB 无 → 删除
//   - model 中 StorageHint 为 nil 的字段不受影响
func (s *ModelDesignAppService) SyncFieldsFromDB(
	ctx context.Context,
	modelID string,
	dbFields []*modeldesign.FieldDefinition,
) error {
	opts := modeldesign.NewModelQueryOptions().WithFields()
	existingModel, err := s.modelRepo.GetByID(ctx, modelID, opts)
	if err != nil {
		return err
	}
	if existingModel == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ModelNotFound, modelID)
	}

	// 构建 DB 字段的 storageHint → field 映射
	dbFieldByHint := make(map[string]*modeldesign.FieldDefinition, len(dbFields))
	for _, f := range dbFields {
		if f.StorageHint != nil && *f.StorageHint != "" {
			dbFieldByHint[*f.StorageHint] = f
		}
	}

	// 分类现有字段
	var toDelete []string
	var toAdd []*modeldesign.FieldDefinition
	var toUpdate []*modeldesign.FieldDefinition

	existingByHint := make(map[string]*modeldesign.FieldDefinition)
	for _, ef := range existingModel.Fields {
		if ef.StorageHint == nil || *ef.StorageHint == "" {
			continue // 逻辑字段，不碰
		}
		hint := *ef.StorageHint
		existingByHint[hint] = ef
		if _, inDB := dbFieldByHint[hint]; !inDB {
			toDelete = append(toDelete, ef.Name)
		}
	}

	for hint, dbField := range dbFieldByHint {
		if ef, exists := existingByHint[hint]; exists {
			// 强制覆盖：用 DB 字段属性更新
			dbField.ModelID = modelID
			dbField.ModelLocator = existingModel.GetModelLocator()
			dbField.Name = ef.Name // 保留字段名不变
			dbField.DisplayOrder = ef.DisplayOrder
			toUpdate = append(toUpdate, dbField)
		} else {
			dbField.ModelID = modelID
			dbField.ModelLocator = existingModel.GetModelLocator()
			toAdd = append(toAdd, dbField)
		}
	}

	// 执行删除
	if len(toDelete) > 0 {
		if err := s.modelRepo.DeleteFields(ctx, modelID, toDelete); err != nil {
			return bizerrors.Wrapf(err, "SyncFieldsFromDB: delete fields")
		}
	}

	// 执行覆盖更新
	if len(toUpdate) > 0 {
		if err := s.modelRepo.BulkUpdateFields(ctx, toUpdate); err != nil {
			return bizerrors.Wrapf(err, "SyncFieldsFromDB: update fields")
		}
	}

	// 执行新增
	if len(toAdd) > 0 {
		addCmd := AddFieldCommand{ModelID: modelID, Fields: toAdd}
		if err := s.AddFieldSync(ctx, addCmd); err != nil {
			return bizerrors.Wrapf(err, "SyncFieldsFromDB: add fields")
		}
	}

	return nil
}
```

- [ ] **Step 4：编译验证**

```bash
cd modelcraft-backend && go build ./...
```

预期：编译通过。

- [ ] **Step 5：commit**

```bash
cd modelcraft-backend && git add \
  internal/app/modeldesign/reverse_engineer_app.go \
  internal/app/modeldesign/model_builder.go \
  internal/app/modeldesign/model_app.go
git commit -m "feat: expose GetTableDefinition, NormalizeModelName; add SyncFieldsFromDB"
```

---

## Task 7：App Service 单测

**Files:**
- Create: `modelcraft-backend/internal/app/modeldatabase/sync_models_app_test.go`

参照 `sync_job_app_test.go` 的结构编写 fake 依赖。

- [ ] **Step 1：创建测试文件**

新建 `modelcraft-backend/internal/app/modeldatabase/sync_models_app_test.go`：

```go
package modeldatabase

import (
	"context"
	"errors"
	"testing"
	"time"

	appmodeldesign "modelcraft/internal/app/modeldesign"
	domaindb "modelcraft/internal/domain/modeldatabase"
	domainmodel "modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/ctxutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Fakes ────────────────────────────────────────────────────────────────

type fakeSyncModelsJobRepo struct {
	jobs          map[string]*domaindb.ModelSyncJob
	activeByDB    map[string]*domaindb.ModelSyncJob
	created       []*domaindb.ModelSyncJob
	updateHistory []*domaindb.ModelSyncJob
}

func newFakeSyncModelsJobRepo() *fakeSyncModelsJobRepo {
	return &fakeSyncModelsJobRepo{
		jobs:       map[string]*domaindb.ModelSyncJob{},
		activeByDB: map[string]*domaindb.ModelSyncJob{},
	}
}

func (f *fakeSyncModelsJobRepo) Create(_ context.Context, job *domaindb.ModelSyncJob) error {
	cloned := *job
	f.jobs[job.ID] = &cloned
	f.created = append(f.created, &cloned)
	if job.Status == domaindb.ModelSyncJobStatusPending || job.Status == domaindb.ModelSyncJobStatusRunning {
		f.activeByDB[job.DatabaseName] = &cloned
	}
	return nil
}

func (f *fakeSyncModelsJobRepo) GetByID(_ context.Context, orgName, projectSlug, jobID string) (*domaindb.ModelSyncJob, error) {
	job := f.jobs[jobID]
	if job == nil || job.OrgName != orgName || job.ProjectSlug != projectSlug {
		return nil, shared.NewNotFoundError("sync job not found")
	}
	cloned := *job
	return &cloned, nil
}

func (f *fakeSyncModelsJobRepo) GetActiveByDatabase(_ context.Context, orgName, projectSlug, databaseName string) (*domaindb.ModelSyncJob, error) {
	job := f.activeByDB[databaseName]
	if job == nil || job.OrgName != orgName || job.ProjectSlug != projectSlug {
		return nil, nil
	}
	cloned := *job
	return &cloned, nil
}

func (f *fakeSyncModelsJobRepo) Update(_ context.Context, job *domaindb.ModelSyncJob) error {
	cloned := *job
	f.jobs[job.ID] = &cloned
	f.updateHistory = append(f.updateHistory, &cloned)
	if job.Status == domaindb.ModelSyncJobStatusPending || job.Status == domaindb.ModelSyncJobStatusRunning {
		f.activeByDB[job.DatabaseName] = &cloned
	} else {
		delete(f.activeByDB, job.DatabaseName)
	}
	return nil
}

type fakeSyncModelsReverseEngineer struct {
	tables     []string
	listErr    error
	tableDefs  map[string]*appmodeldesign.TableDefinitionResult
	tableErrs  map[string]error
	importErrs map[string]error
}

func (f *fakeSyncModelsReverseEngineer) ListTables(
	_ context.Context, _, _, _ string, _ bool, _, _ int,
) (*appmodeldesign.ListTablesResult, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return &appmodeldesign.ListTablesResult{Tables: f.tables, TotalCount: len(f.tables)}, nil
}

func (f *fakeSyncModelsReverseEngineer) ImportModel(
	_ context.Context, cmd appmodeldesign.ImportModelCommand,
) (*appmodeldesign.ImportModelResult, error) {
	if err := f.importErrs[cmd.TableName]; err != nil {
		return nil, err
	}
	return &appmodeldesign.ImportModelResult{ModelID: "model-" + cmd.TableName, ModelName: cmd.TableName}, nil
}

func (f *fakeSyncModelsReverseEngineer) GetTableDefinition(
	_ context.Context, _, _, _, tableName string,
) (*appmodeldesign.TableDefinitionResult, error) {
	if err := f.tableErrs[tableName]; err != nil {
		return nil, err
	}
	return f.tableDefs[tableName], nil
}

type fakeSyncModelsModelRepo struct {
	modelsByName map[string]*domainmodel.DataModel
}

func (f *fakeSyncModelsModelRepo) GetByName(
	_ context.Context, orgName, databaseName, name, projectSlug string,
	_ ...*domainmodel.ModelQueryOptions,
) (*domainmodel.DataModel, error) {
	key := databaseName + ":" + name
	m := f.modelsByName[key]
	if m == nil {
		return nil, shared.NewNotFoundError("not found")
	}
	return m, nil
}

type fakeFieldSyncer struct {
	calls  []string
	errors map[string]error
}

func (f *fakeFieldSyncer) SyncFieldsFromDB(_ context.Context, modelID string, _ []*domainmodel.FieldDefinition) error {
	f.calls = append(f.calls, modelID)
	if err := f.errors[modelID]; err != nil {
		return err
	}
	return nil
}

func syncModelsProjectCtx() context.Context {
	return ctxutils.SetProjectSlug(ctxutils.SetOrgName(context.Background(), "org-a"), "proj-a")
}

// ── Tests ────────────────────────────────────────────────────────────────

func TestSyncModels_RejectsActiveJob(t *testing.T) {
	ctx := syncModelsProjectCtx()
	repo := newFakeSyncModelsJobRepo()
	// seed active job
	repo.activeByDB["orders"] = &domaindb.ModelSyncJob{
		ID:           "j1",
		OrgName:      "org-a",
		ProjectSlug:  "proj-a",
		DatabaseName: "orders",
		Status:       domaindb.ModelSyncJobStatusRunning,
	}

	svc := NewSyncModelsAppService(SyncModelsAppServiceDeps{
		SyncJobRepo: repo,
		Runner:      &fakeBackgroundRunner{},
	})

	_, err := svc.StartSync(ctx, SyncModelsFromDBCommand{DatabaseName: "orders", TableNames: []string{"t1"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestSyncModels_RejectsNoTablesAndNoSyncAll(t *testing.T) {
	ctx := syncModelsProjectCtx()
	svc := NewSyncModelsAppService(SyncModelsAppServiceDeps{
		SyncJobRepo: newFakeSyncModelsJobRepo(),
		Runner:      &fakeBackgroundRunner{},
	})
	_, err := svc.StartSync(ctx, SyncModelsFromDBCommand{DatabaseName: "orders"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must specify")
}

func TestSyncModels_RejectsBothTablesAndSyncAll(t *testing.T) {
	ctx := syncModelsProjectCtx()
	svc := NewSyncModelsAppService(SyncModelsAppServiceDeps{
		SyncJobRepo: newFakeSyncModelsJobRepo(),
		Runner:      &fakeBackgroundRunner{},
	})
	_, err := svc.StartSync(ctx, SyncModelsFromDBCommand{
		DatabaseName: "orders",
		TableNames:   []string{"t1"},
		SyncAll:      true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot specify both")
}

func TestSyncModels_CreatesNewModels(t *testing.T) {
	ctx := syncModelsProjectCtx()
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	repo := newFakeSyncModelsJobRepo()

	storageHint := "id"
	reverseEng := &fakeSyncModelsReverseEngineer{
		tableDefs: map[string]*appmodeldesign.TableDefinitionResult{
			"new_table": {
				TableName: "new_table",
				Fields: []*domainmodel.FieldDefinition{
					{Name: "id", StorageHint: &storageHint},
				},
			},
		},
		tableErrs:  map[string]error{},
		importErrs: map[string]error{},
	}
	modelRepo := &fakeSyncModelsModelRepo{modelsByName: map[string]*domainmodel.DataModel{}}
	fieldSyncer := &fakeFieldSyncer{errors: map[string]error{}}

	svc := NewSyncModelsAppService(SyncModelsAppServiceDeps{
		SyncJobRepo:     repo,
		ReverseEngineer: reverseEng,
		ModelRepo:       modelRepo,
		FieldSyncer:     fieldSyncer,
		Runner: &fakeBackgroundRunner{
			run: func(ctx context.Context, fn func(context.Context)) { fn(ctx) },
		},
		Now: func() time.Time { return now },
	})

	job, err := svc.StartSync(ctx, SyncModelsFromDBCommand{
		DatabaseName: "orders",
		TableNames:   []string{"new_table"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, job.ID)

	saved, err := repo.GetByID(ctx, "org-a", "proj-a", job.ID)
	require.NoError(t, err)
	assert.Equal(t, domaindb.ModelSyncJobStatusSucceeded, saved.Status)
	assert.Equal(t, 1, saved.TotalTables)
	assert.Equal(t, 1, saved.CreatedModels)
	assert.Equal(t, 0, saved.SyncedModels)
	assert.Equal(t, 0, saved.FailedCount)
}

func TestSyncModels_SyncsExistingModel(t *testing.T) {
	ctx := syncModelsProjectCtx()
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	repo := newFakeSyncModelsJobRepo()

	storageHint := "id"
	reverseEng := &fakeSyncModelsReverseEngineer{
		tableDefs: map[string]*appmodeldesign.TableDefinitionResult{
			"orders": {
				TableName: "orders",
				Fields:    []*domainmodel.FieldDefinition{{Name: "id", StorageHint: &storageHint}},
			},
		},
		tableErrs:  map[string]error{},
		importErrs: map[string]error{},
	}
	modelRepo := &fakeSyncModelsModelRepo{
		modelsByName: map[string]*domainmodel.DataModel{
			"orders:orders": {ModelMeta: domainmodel.ModelMeta{ID: "model-orders"}},
		},
	}
	fieldSyncer := &fakeFieldSyncer{errors: map[string]error{}}

	svc := NewSyncModelsAppService(SyncModelsAppServiceDeps{
		SyncJobRepo:     repo,
		ReverseEngineer: reverseEng,
		ModelRepo:       modelRepo,
		FieldSyncer:     fieldSyncer,
		Runner: &fakeBackgroundRunner{
			run: func(ctx context.Context, fn func(context.Context)) { fn(ctx) },
		},
		Now: func() time.Time { return now },
	})

	job, err := svc.StartSync(ctx, SyncModelsFromDBCommand{
		DatabaseName: "orders",
		TableNames:   []string{"orders"},
	})
	require.NoError(t, err)

	saved, err := repo.GetByID(ctx, "org-a", "proj-a", job.ID)
	require.NoError(t, err)
	assert.Equal(t, domaindb.ModelSyncJobStatusSucceeded, saved.Status)
	assert.Equal(t, 1, saved.SyncedModels)
	assert.Equal(t, 0, saved.CreatedModels)
	assert.Equal(t, []string{"model-orders"}, fieldSyncer.calls)
}

func TestSyncModels_SingleTableFailureContinues(t *testing.T) {
	ctx := syncModelsProjectCtx()
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	repo := newFakeSyncModelsJobRepo()

	storageHint := "id"
	reverseEng := &fakeSyncModelsReverseEngineer{
		tableDefs: map[string]*appmodeldesign.TableDefinitionResult{
			"good_table":   {TableName: "good_table", Fields: []*domainmodel.FieldDefinition{{Name: "id", StorageHint: &storageHint}}},
			"broken_table": {TableName: "broken_table", Fields: []*domainmodel.FieldDefinition{{Name: "id", StorageHint: &storageHint}}},
		},
		tableErrs:  map[string]error{"broken_table": errors.New("introspect failed")},
		importErrs: map[string]error{},
	}
	modelRepo := &fakeSyncModelsModelRepo{modelsByName: map[string]*domainmodel.DataModel{}}
	fieldSyncer := &fakeFieldSyncer{errors: map[string]error{}}

	svc := NewSyncModelsAppService(SyncModelsAppServiceDeps{
		SyncJobRepo:     repo,
		ReverseEngineer: reverseEng,
		ModelRepo:       modelRepo,
		FieldSyncer:     fieldSyncer,
		Runner: &fakeBackgroundRunner{
			run: func(ctx context.Context, fn func(context.Context)) { fn(ctx) },
		},
		Now: func() time.Time { return now },
	})

	job, err := svc.StartSync(ctx, SyncModelsFromDBCommand{
		DatabaseName: "orders",
		TableNames:   []string{"good_table", "broken_table"},
	})
	require.NoError(t, err)

	saved, err := repo.GetByID(ctx, "org-a", "proj-a", job.ID)
	require.NoError(t, err)
	assert.Equal(t, domaindb.ModelSyncJobStatusPartialSuccess, saved.Status)
	assert.Equal(t, 1, saved.CreatedModels)
	assert.Equal(t, 1, saved.FailedCount)
	require.Len(t, saved.FailedTables, 1)
	assert.Equal(t, "broken_table", saved.FailedTables[0].TableName)
}
```

- [ ] **Step 2：运行测试（先确认失败，再迭代）**

```bash
cd modelcraft-backend && go test ./internal/app/modeldatabase/... -v -run TestSyncModels
```

预期：如果 Task 5 和 Task 6 都完成，这里全部 PASS。

- [ ] **Step 3：commit**

```bash
cd modelcraft-backend && git add internal/app/modeldatabase/sync_models_app_test.go
git commit -m "test: add SyncModelsAppService unit tests"
```

---

## Task 8：GraphQL Schema + codegen

**Files:**
- Modify: `modelcraft-backend/api/graph/project/schema/model.graphql`
- Regenerate: `modelcraft-backend/internal/interfaces/graphql/project/generated/`

- [ ] **Step 1：在 `model.graphql` 末尾（mutations 块之前）添加新类型**

在文件的 `# Model queries` 注释之前插入以下内容：

```graphql
# ============================================
# syncModelsFromDB Job Types
# ============================================

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
  tableNames:       [String!]!
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

type SyncModelsFromDBPayload {
  jobId: ID!
}

input SyncModelsFromDBInput {
  databaseName: String!
  tableNames:   [String!]
  syncAll:      Boolean
}
```

- [ ] **Step 2：在 `extend type Query` 块添加**

```graphql
  modelSyncJob(jobId: ID!): ModelSyncJob @hasPermission(action: "model:read")
```

- [ ] **Step 3：在 `extend type Mutation` 块添加**

```graphql
  syncModelsFromDB(input: SyncModelsFromDBInput!): SyncModelsFromDBPayload! @hasPermission(action: "model:create")
```

- [ ] **Step 4：运行 gqlgen codegen**

```bash
cd modelcraft-backend && just generate-gql
```

预期：生成新的 resolver stub，编译报错（`SyncModelsFromDB` 和 `ModelSyncJob` resolver 未实现）。

- [ ] **Step 5：commit（schema + 生成代码）**

```bash
cd modelcraft-backend && git add api/graph/project/schema/model.graphql internal/interfaces/graphql/project/generated/
git commit -m "feat: add syncModelsFromDB and modelSyncJob to GraphQL schema"
```

---

## Task 9：Resolver 实现

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/resolver.go`
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/model.resolvers.go`

- [ ] **Step 1：在 `resolver.go` 的 `Resolver` struct 加字段**

```go
// Model sync (syncModelsFromDB)
SyncModelsAppService *appmodeldatabase.SyncModelsAppService
```

- [ ] **Step 2：在 `model.resolvers.go` 实现 `SyncModelsFromDB` resolver**

```go
// SyncModelsFromDB is the resolver for the syncModelsFromDB field.
func (r *mutationResolver) SyncModelsFromDB(ctx context.Context, input generated.SyncModelsFromDBInput) (*generated.SyncModelsFromDBPayload, error) {
	tableNames := make([]string, 0)
	if input.TableNames != nil {
		tableNames = input.TableNames
	}
	syncAll := false
	if input.SyncAll != nil {
		syncAll = *input.SyncAll
	}

	job, err := r.SyncModelsAppService.StartSync(ctx, appmodeldatabase.SyncModelsFromDBCommand{
		DatabaseName: input.DatabaseName,
		TableNames:   tableNames,
		SyncAll:      syncAll,
	})
	if err != nil {
		return nil, err
	}
	return &generated.SyncModelsFromDBPayload{JobID: job.ID}, nil
}
```

- [ ] **Step 3：在 `model.resolvers.go` 实现 `ModelSyncJob` query resolver**

```go
// ModelSyncJob is the resolver for the modelSyncJob field.
func (r *queryResolver) ModelSyncJob(ctx context.Context, jobID string) (*generated.ModelSyncJob, error) {
	job, err := r.SyncModelsAppService.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, nil
	}
	return modelSyncJobToGQL(job), nil
}

func modelSyncJobToGQL(job *domaindb.ModelSyncJob) *generated.ModelSyncJob {
	failedTables := make([]generated.ModelSyncFailedTable, len(job.FailedTables))
	for i, ft := range job.FailedTables {
		failedTables[i] = generated.ModelSyncFailedTable{
			TableName: ft.TableName,
			Message:   ft.Message,
		}
	}
	tableNames := job.TableNames
	if tableNames == nil {
		tableNames = []string{}
	}
	return &generated.ModelSyncJob{
		ID:              job.ID,
		DatabaseName:    job.DatabaseName,
		TableNames:      tableNames,
		Status:          generated.ModelSyncJobStatus(job.Status),
		TotalTables:     job.TotalTables,
		ProcessedTables: job.ProcessedTables,
		CreatedModels:   job.CreatedModels,
		SyncedModels:    job.SyncedModels,
		FailedCount:     job.FailedCount,
		FailedTables:    failedTables,
		StartedAt:       job.StartedAt,
		FinishedAt:      job.FinishedAt,
		CreatedAt:       job.CreatedAt,
		UpdatedAt:       job.UpdatedAt,
	}
}
```

**注意：** 需在文件顶部 import `appmodeldatabase "modelcraft/internal/app/modeldatabase"` 和 `domaindb "modelcraft/internal/domain/modeldatabase"`。

- [ ] **Step 4：编译验证**

```bash
cd modelcraft-backend && go build ./...
```

预期：编译通过。

- [ ] **Step 5：运行所有测试**

```bash
cd modelcraft-backend && go test ./... -count=1
```

预期：全部 PASS（无回归）。

- [ ] **Step 6：commit**

```bash
cd modelcraft-backend && git add \
  internal/interfaces/graphql/project/resolver.go \
  internal/interfaces/graphql/project/model.resolvers.go
git commit -m "feat: implement syncModelsFromDB and modelSyncJob resolvers"
```

---

## Task 10：前端 contract 同步 + hooks

**Files:**
- Run: `front-contract-pull` skill（同步 contract）
- Modify: `modelcraft-front/src/api-client/project/model-graphql-docs.ts`
- Create: `modelcraft-front/src/web/hooks/model/use-sync-models-from-db.ts`

- [ ] **Step 1：同步前端 contract**

在 Claude Code 中运行：
```
/front-contract-pull
```

预期：`modelcraft-front/src/generated/graphql.ts` 和 `contract/` 下新增 `SyncModelsFromDB`、`ModelSyncJob` 等类型。

- [ ] **Step 2：在 `model-graphql-docs.ts` 添加 GraphQL 文档**

在文件末尾添加：

```typescript
export const SYNC_MODELS_FROM_DB_MUTATION = gql`
  mutation SyncModelsFromDB($input: SyncModelsFromDBInput!) {
    syncModelsFromDB(input: $input) {
      jobId
    }
  }
`

export const MODEL_SYNC_JOB_QUERY = gql`
  query ModelSyncJob($jobId: ID!) {
    modelSyncJob(jobId: $jobId) {
      id
      databaseName
      tableNames
      status
      totalTables
      processedTables
      createdModels
      syncedModels
      failedCount
      failedTables {
        tableName
        message
      }
      startedAt
      finishedAt
      createdAt
      updatedAt
    }
  }
`
```

- [ ] **Step 3：创建 hooks**

新建 `modelcraft-front/src/web/hooks/model/use-sync-models-from-db.ts`：

```typescript
import { useMutation, useQuery } from '@apollo/client'
import {
  SYNC_MODELS_FROM_DB_MUTATION,
  MODEL_SYNC_JOB_QUERY,
} from '@/api-client/project/model-graphql-docs'
import type {
  SyncModelsFromDBInput,
  ModelSyncJob,
} from '@/generated/graphql'

interface SyncModelsFromDBData {
  syncModelsFromDB: { jobId: string }
}

interface ModelSyncJobData {
  modelSyncJob: ModelSyncJob | null
}

export function useSyncModelsFromDB() {
  return useMutation<SyncModelsFromDBData, { input: SyncModelsFromDBInput }>(
    SYNC_MODELS_FROM_DB_MUTATION
  )
}

export function useModelSyncJob(jobId: string | null) {
  return useQuery<ModelSyncJobData, { jobId: string }>(MODEL_SYNC_JOB_QUERY, {
    variables: { jobId: jobId! },
    skip: !jobId,
    pollInterval: 2000, // 每 2 秒轮询一次
  })
}
```

- [ ] **Step 4：前端编译验证**

```bash
cd modelcraft-front && npx tsc --noEmit
```

预期：无类型错误。

- [ ] **Step 5：commit**

```bash
cd modelcraft-front && git add \
  src/api-client/project/model-graphql-docs.ts \
  src/web/hooks/model/use-sync-models-from-db.ts \
  src/generated/ \
  contract/
git commit -m "feat: add syncModelsFromDB and modelSyncJob frontend hooks"
```

---

## Task 11：迁移 ImportModelDialog

**Files:**
- Modify: `modelcraft-front/src/web/components/features/model-editor/ImportModelDialog.tsx`

- [ ] **Step 1：查看当前 Dialog 实现**

```bash
cat modelcraft-front/src/web/components/features/model-editor/ImportModelDialog.tsx
```

- [ ] **Step 2：替换 mutation 调用**

将 Dialog 中调用 `importModel` mutation 的部分替换为调用 `syncModelsFromDB`（传 `tableNames: [tableName]`），并在成功后轮询 `modelSyncJob` 直到状态不是 `PENDING/RUNNING`，再调用 `onSuccess`。

核心改动示意：

```typescript
// 替换前
const [importModel] = useMutation(IMPORT_MODEL_MUTATION)
// 在 handleSubmit 中：
await importModel({ variables: { input: { databaseName, tableName } } })
onSuccess()

// 替换后
const [syncModels] = useSyncModelsFromDB()
const [jobId, setJobId] = useState<string | null>(null)
const { data: jobData } = useModelSyncJob(jobId)

// 在 handleSubmit 中：
const { data } = await syncModels({
  variables: { input: { databaseName, tableNames: [tableName] } }
})
setJobId(data?.syncModelsFromDB.jobId ?? null)

// 在 useEffect 中监听 job 完成：
useEffect(() => {
  if (!jobData?.modelSyncJob) return
  const status = jobData.modelSyncJob.status
  if (status === 'SUCCEEDED' || status === 'PARTIAL_SUCCESS') {
    onSuccess()
    setJobId(null)
  } else if (status === 'FAILED') {
    setError('同步失败，请重试')
    setJobId(null)
  }
}, [jobData])
```

- [ ] **Step 3：前端 lint**

```bash
cd modelcraft-front && npm run lint
```

预期：无 lint 错误。

- [ ] **Step 4：前端编译**

```bash
cd modelcraft-front && npx tsc --noEmit
```

预期：无类型错误。

- [ ] **Step 5：commit**

```bash
cd modelcraft-front && git add src/web/components/features/model-editor/ImportModelDialog.tsx
git commit -m "feat: migrate ImportModelDialog to use syncModelsFromDB"
```

---

## Self-Review

**Spec coverage 检查：**

| Spec 要求 | 计划中的 Task |
|-----------|--------------|
| `storage_hint` 列加入 DB schema | Task 1（已部分完成）|
| `model_sync_job` 表 | Task 2 |
| GraphQL `syncModelsFromDB` + `modelSyncJob` | Task 8 + 9 |
| 参数校验（互斥、空数组、activeJob 报错） | Task 5 `validateSyncModelsCommand` + `StartSync` |
| upsert 语义（model 不存在建新，已存在同步字段） | Task 5 `processTable` |
| 字段 full sync 以 `storageHint` 为准，强制覆盖 | Task 6 `SyncFieldsFromDB` |
| Job 状态机（PENDING→RUNNING→SUCCEEDED/PARTIAL_SUCCESS/FAILED） | Task 5 `RunSyncJob` |
| repository 接口 + 实现 | Task 3 + 4 |
| 应用层单测 | Task 7 |
| 前端 hooks | Task 10 |
| ImportModelDialog 迁移 | Task 11 |

所有 spec 要求均已覆盖。`importModel` 移除留作后续独立 PR（非本计划范围，spec Section 6 已明确废弃顺序）。
