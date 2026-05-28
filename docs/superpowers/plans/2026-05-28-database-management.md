# Database Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增"数据库管理"功能，让开发者显式接管（注册）MySQL database，设置自建或托管模式，并限制 ModelSidebar 只展示已接管的 database。

**Architecture:** 后端新增 `model_database` 表和 domain 层，通过 Project GraphQL Schema 暴露 CRUD API；前端新增独立页面和注册/编辑 Dialog，同时修改 ModelSidebar 改用已接管 database 列表。

**Tech Stack:** Go (DDD 分层架构 + sqlc + Atlas 迁移), GraphQL (gqlgen), Next.js (App Router), Apollo Client, shadcn/ui

**Spec:** `docs/superpowers/specs/2026-05-28-database-management-design.md`

---

## 总体顺序

1. 后端数据库迁移
2. sqlc 查询
3. Domain 层
4. Repository 层
5. AppService 层
6. GraphQL Schema + codegen
7. GraphQL Resolver
8. DI 注册
9. 前端 GraphQL 文档
10. 前端 Hook
11. 前端页面和组件
12. ModelSidebar 改动

---

## Task 1: Atlas 数据库迁移

**Files:**
- Create: `modelcraft-backend/db/schema/mysql/16_model_database.sql`

- [ ] **Step 1: 创建迁移文件**

```sql
-- =============================================================================
-- Model Database Registry (2026)
-- 说明：
-- - 新增 model_database 表，记录 Project 已接管的 MySQL database
-- - mode: self_hosted（可读写，支持新建/导入模型）/ managed（只读，仅同步模型）
-- - 通过 (project_slug, name, delete_token) 联合唯一索引支持软删除后重复接管
-- =============================================================================

CREATE TABLE `model_database` (
  `id`           VARCHAR(36)   NOT NULL,
  `org_name`     VARCHAR(64)   NOT NULL,
  `project_slug` VARCHAR(64)   NOT NULL,
  `cluster_id`   VARCHAR(36)   NOT NULL,
  `name`         VARCHAR(64)   NOT NULL         COMMENT 'MySQL database 原始名，来自 SHOW DATABASES，注册后不可修改',
  `title`        VARCHAR(128)  NOT NULL         COMMENT '用户设置的友好名称，默认等于 name',
  `description`  TEXT          NOT NULL         COMMENT '可选描述',
  `mode`         ENUM('self_hosted','managed') NOT NULL COMMENT 'self_hosted=可读写; managed=只读',
  `delete_token` VARCHAR(36)   NOT NULL DEFAULT '' COMMENT '软删除 token，配合联合唯一索引允许重复接管',
  `deleted_at`   DATETIME,
  `created_at`   DATETIME      NOT NULL,
  `updated_at`   DATETIME      NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_project_database` (`project_slug`, `name`, `delete_token`),
  INDEX `idx_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

- [ ] **Step 2: 应用迁移**

在 `modelcraft-backend/` 目录下运行：
```bash
just migrate
```

Expected: 迁移成功，`model_database` 表已创建。

- [ ] **Step 3: Commit**

```bash
git add modelcraft-backend/db/schema/mysql/16_model_database.sql
git commit -m "feat(db): add model_database table for database registry"
```

---

## Task 2: sqlc 查询文件

**Files:**
- Create: `modelcraft-backend/db/queries/model_database.sql`

- [ ] **Step 1: 创建 sqlc 查询文件**

```sql
-- name: CreateModelDatabase :exec
INSERT INTO model_database (id, org_name, project_slug, cluster_id, name, title, description, mode, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(3), NOW(3));

-- name: GetModelDatabaseByID :one
SELECT * FROM model_database
WHERE id = ? AND org_name = ? AND project_slug = ? AND `model_database`.`deleted_at` IS NULL
LIMIT 1;

-- name: GetModelDatabaseByName :one
SELECT * FROM model_database
WHERE org_name = ? AND project_slug = ? AND name = ? AND `model_database`.`deleted_at` IS NULL
LIMIT 1;

-- name: ListModelDatabases :many
SELECT * FROM model_database
WHERE org_name = ? AND project_slug = ? AND `model_database`.`deleted_at` IS NULL
ORDER BY created_at ASC;

-- name: UpdateModelDatabase :exec
UPDATE model_database
SET title = ?, description = ?, mode = ?, updated_at = NOW(3)
WHERE id = ? AND org_name = ? AND project_slug = ? AND `model_database`.`deleted_at` IS NULL;

-- name: DeleteModelDatabase :exec
UPDATE model_database
SET `deleted_at` = NOW(3),
    `delete_token` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS CHAR)
WHERE id = ? AND org_name = ? AND project_slug = ? AND `model_database`.`deleted_at` IS NULL;
```

- [ ] **Step 2: 运行 sqlc 代码生成**

```bash
cd modelcraft-backend && just generate-sqlc
```

Expected: `internal/infrastructure/dbgen/` 目录下生成 `model_database.sql.go`，包含 `CreateModelDatabase`、`GetModelDatabaseByID` 等方法。

- [ ] **Step 3: Commit**

```bash
git add modelcraft-backend/db/queries/model_database.sql modelcraft-backend/internal/infrastructure/dbgen/
git commit -m "feat(db): add sqlc queries for model_database"
```

---

## Task 3: Domain 层

**Files:**
- Create: `modelcraft-backend/internal/domain/modeldatabase/model_database.go`
- Create: `modelcraft-backend/internal/domain/modeldatabase/repository.go`

- [ ] **Step 1: 创建领域实体文件**

`modelcraft-backend/internal/domain/modeldatabase/model_database.go`:

```go
package modeldatabase

import "time"

// DatabaseMode represents the access mode of a registered database.
type DatabaseMode string

const (
	// DatabaseModeSelfHosted allows full read/write access and model creation.
	DatabaseModeSelfHosted DatabaseMode = "self_hosted"
	// DatabaseModeManaged allows read-only access; only model sync is permitted.
	DatabaseModeManaged DatabaseMode = "managed"
)

// ModelDatabase represents a MySQL database that has been registered (taken over)
// by a project. Only registered databases appear in the ModelSidebar.
type ModelDatabase struct {
	ID          string
	OrgName     string
	ProjectSlug string
	ClusterID   string
	Name        string       // MySQL原始库名，来自 SHOW DATABASES，注册后不可修改
	Title       string       // 用户设置的友好名称
	Description string
	Mode        DatabaseMode
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
```

- [ ] **Step 2: 创建 Repository 接口文件**

`modelcraft-backend/internal/domain/modeldatabase/repository.go`:

```go
package modeldatabase

import "context"

// ModelDatabaseRepository defines the persistence interface for ModelDatabase.
type ModelDatabaseRepository interface {
	Create(ctx context.Context, db *ModelDatabase) error
	GetByID(ctx context.Context, orgName, projectSlug, id string) (*ModelDatabase, error)
	GetByName(ctx context.Context, orgName, projectSlug, name string) (*ModelDatabase, error)
	List(ctx context.Context, orgName, projectSlug string) ([]*ModelDatabase, error)
	Update(ctx context.Context, orgName, projectSlug string, db *ModelDatabase) error
	Delete(ctx context.Context, orgName, projectSlug, id string) error
}
```

- [ ] **Step 3: Commit**

```bash
git add modelcraft-backend/internal/domain/modeldatabase/
git commit -m "feat(domain): add ModelDatabase domain entity and repository interface"
```

---

## Task 4: Repository（Infrastructure）层

**Files:**
- Create: `modelcraft-backend/internal/infrastructure/repository/sql_model_database_repository.go`

- [ ] **Step 1: 创建 Repository 实现**

`modelcraft-backend/internal/infrastructure/repository/sql_model_database_repository.go`:

```go
package repository

import (
	"context"
	"modelcraft/internal/domain/modeldatabase"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/sqlerr"
	"time"
)

// SqlModelDatabaseRepository implements modeldatabase.ModelDatabaseRepository using sqlc.
type SqlModelDatabaseRepository struct {
	q dbgen.Querier
}

// NewSqlModelDatabaseRepository creates a new SqlModelDatabaseRepository.
func NewSqlModelDatabaseRepository(q dbgen.Querier) *SqlModelDatabaseRepository {
	return &SqlModelDatabaseRepository{q: q}
}

func (r *SqlModelDatabaseRepository) Create(ctx context.Context, db *modeldatabase.ModelDatabase) error {
	return r.q.CreateModelDatabase(ctx, dbgen.CreateModelDatabaseParams{
		ID:          db.ID,
		OrgName:     db.OrgName,
		ProjectSlug: db.ProjectSlug,
		ClusterID:   db.ClusterID,
		Name:        db.Name,
		Title:       db.Title,
		Description: db.Description,
		Mode:        string(db.Mode),
	})
}

func (r *SqlModelDatabaseRepository) GetByID(ctx context.Context, orgName, projectSlug, id string) (*modeldatabase.ModelDatabase, error) {
	row, err := r.q.GetModelDatabaseByID(ctx, dbgen.GetModelDatabaseByIDParams{
		ID:          id,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("model database not found: " + id)
		}
		return nil, err
	}
	return modelDatabaseToDomain(row), nil
}

func (r *SqlModelDatabaseRepository) GetByName(ctx context.Context, orgName, projectSlug, name string) (*modeldatabase.ModelDatabase, error) {
	row, err := r.q.GetModelDatabaseByName(ctx, dbgen.GetModelDatabaseByNameParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		Name:        name,
	})
	if err != nil {
		if sqlerr.IsNotFoundError(err) {
			return nil, shared.NewNotFoundError("model database not found: " + name)
		}
		return nil, err
	}
	return modelDatabaseToDomain(row), nil
}

func (r *SqlModelDatabaseRepository) List(ctx context.Context, orgName, projectSlug string) ([]*modeldatabase.ModelDatabase, error) {
	rows, err := r.q.ListModelDatabases(ctx, dbgen.ListModelDatabasesParams{
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}
	result := make([]*modeldatabase.ModelDatabase, 0, len(rows))
	for _, row := range rows {
		result = append(result, modelDatabaseToDomain(row))
	}
	return result, nil
}

func (r *SqlModelDatabaseRepository) Update(ctx context.Context, orgName, projectSlug string, db *modeldatabase.ModelDatabase) error {
	return r.q.UpdateModelDatabase(ctx, dbgen.UpdateModelDatabaseParams{
		Title:       db.Title,
		Description: db.Description,
		Mode:        string(db.Mode),
		ID:          db.ID,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
}

func (r *SqlModelDatabaseRepository) Delete(ctx context.Context, orgName, projectSlug, id string) error {
	return r.q.DeleteModelDatabase(ctx, dbgen.DeleteModelDatabaseParams{
		ID:          id,
		OrgName:     orgName,
		ProjectSlug: projectSlug,
	})
}

// modelDatabaseToDomain converts a dbgen row to domain entity.
func modelDatabaseToDomain(row dbgen.ModelDatabase) *modeldatabase.ModelDatabase {
	var createdAt, updatedAt time.Time
	if row.CreatedAt.Valid {
		createdAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		updatedAt = row.UpdatedAt.Time
	}
	return &modeldatabase.ModelDatabase{
		ID:          row.ID,
		OrgName:     row.OrgName,
		ProjectSlug: row.ProjectSlug,
		ClusterID:   row.ClusterID,
		Name:        row.Name,
		Title:       row.Title,
		Description: row.Description,
		Mode:        modeldatabase.DatabaseMode(row.Mode),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
```

- [ ] **Step 2: 编译验证**

```bash
cd modelcraft-backend && go build ./...
```

Expected: 编译成功，无报错。

- [ ] **Step 3: Commit**

```bash
git add modelcraft-backend/internal/infrastructure/repository/sql_model_database_repository.go
git commit -m "feat(infra): add SqlModelDatabaseRepository"
```

---

## Task 5: AppService 层

**Files:**
- Create: `modelcraft-backend/internal/app/modeldatabase/model_database_app.go`

- [ ] **Step 1: 创建 AppService**

`modelcraft-backend/internal/app/modeldatabase/model_database_app.go`:

```go
package modeldatabase

import (
	"context"
	"modelcraft/internal/domain/modeldatabase"
	"modelcraft/internal/domain/shared"
	domaincluster "modelcraft/internal/domain/cluster"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/ctxutils"
)

// ModelDatabaseAppService handles use cases for database registration.
type ModelDatabaseAppService struct {
	dbRepo      modeldatabase.ModelDatabaseRepository
	clusterRepo domaincluster.DatabaseClusterRepository
	connManager *repository.ClusterConnectionManager
}

// NewModelDatabaseAppService creates a new ModelDatabaseAppService.
func NewModelDatabaseAppService(
	dbRepo modeldatabase.ModelDatabaseRepository,
	clusterRepo domaincluster.DatabaseClusterRepository,
	connManager *repository.ClusterConnectionManager,
) *ModelDatabaseAppService {
	return &ModelDatabaseAppService{
		dbRepo:      dbRepo,
		clusterRepo: clusterRepo,
		connManager: connManager,
	}
}

// ListRegistered returns all registered (taken-over) databases for a project.
func (s *ModelDatabaseAppService) ListRegistered(ctx context.Context) ([]*modeldatabase.ModelDatabase, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	return s.dbRepo.List(ctx, orgName, projectSlug)
}

// ListRaw returns all databases from the cluster, annotated with registration status.
func (s *ModelDatabaseAppService) ListRaw(ctx context.Context) ([]*RawDatabase, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}

	cluster, err := s.clusterRepo.GetByProjectKey(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}

	conn, err := s.connManager.GetConnection(cluster.ID)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ResourceNotFound, "cluster connection not available")
	}

	rawNames, err := listMySQLDatabases(ctx, conn)
	if err != nil {
		return nil, err
	}

	registered, err := s.dbRepo.List(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}

	registeredSet := make(map[string]bool, len(registered))
	for _, r := range registered {
		registeredSet[r.Name] = true
	}

	result := make([]*RawDatabase, 0, len(rawNames))
	for _, name := range rawNames {
		result = append(result, &RawDatabase{
			Name:         name,
			IsRegistered: registeredSet[name],
		})
	}
	return result, nil
}

// Register takes over a database by creating a ModelDatabase record.
func (s *ModelDatabaseAppService) Register(ctx context.Context, cmd RegisterCommand) (*modeldatabase.ModelDatabase, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}

	cluster, err := s.clusterRepo.GetByProjectKey(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}

	// Check duplicate
	_, dupErr := s.dbRepo.GetByName(ctx, orgName, projectSlug, cmd.Name)
	if dupErr == nil {
		return nil, bizerrors.NewError(bizerrors.AlreadyExists, "database already registered: "+cmd.Name)
	}
	if !shared.IsNotFoundError(dupErr) {
		return nil, dupErr
	}

	db := &modeldatabase.ModelDatabase{
		ID:          bizutils.NewULID(),
		OrgName:     orgName,
		ProjectSlug: projectSlug,
		ClusterID:   cluster.ID,
		Name:        cmd.Name,
		Title:       cmd.Title,
		Description: cmd.Description,
		Mode:        cmd.Mode,
	}
	if err := s.dbRepo.Create(ctx, db); err != nil {
		return nil, err
	}
	return db, nil
}

// Update edits the title, description, or mode of a registered database.
func (s *ModelDatabaseAppService) Update(ctx context.Context, cmd UpdateCommand) (*modeldatabase.ModelDatabase, error) {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}

	db, err := s.dbRepo.GetByID(ctx, orgName, projectSlug, cmd.ID)
	if err != nil {
		return nil, err
	}

	if cmd.Title != nil {
		db.Title = *cmd.Title
	}
	if cmd.Description != nil {
		db.Description = *cmd.Description
	}
	if cmd.Mode != nil {
		db.Mode = *cmd.Mode
	}

	if err := s.dbRepo.Update(ctx, orgName, projectSlug, db); err != nil {
		return nil, err
	}
	return db, nil
}

// Unregister soft-deletes a registered database.
func (s *ModelDatabaseAppService) Unregister(ctx context.Context, id string) error {
	orgName, err := ctxutils.GetOrgNameFromContext(ctx)
	if err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, "orgName required")
	}
	projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
	if err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, "projectSlug required")
	}
	return s.dbRepo.Delete(ctx, orgName, projectSlug, id)
}

// RawDatabase is a MySQL database from SHOW DATABASES, annotated with registration status.
type RawDatabase struct {
	Name         string
	IsRegistered bool
}

// RegisterCommand holds input for registering a database.
type RegisterCommand struct {
	Name        string
	Title       string
	Description string
	Mode        modeldatabase.DatabaseMode
}

// UpdateCommand holds input for updating a registered database.
type UpdateCommand struct {
	ID          string
	Title       *string
	Description *string
	Mode        *modeldatabase.DatabaseMode
}
```

- [ ] **Step 2: 创建 MySQL 工具函数**

在同一目录下创建 `modelcraft-backend/internal/app/modeldatabase/mysql_utils.go`:

```go
package modeldatabase

import (
	"context"
	"database/sql"
)

// systemDatabases are MySQL built-in databases that should never be shown to users.
var systemDatabases = map[string]bool{
	"information_schema": true,
	"mysql":              true,
	"performance_schema": true,
	"sys":                true,
}

// listMySQLDatabases runs SHOW DATABASES and returns user-visible database names.
func listMySQLDatabases(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if !systemDatabases[name] {
			names = append(names, name)
		}
	}
	return names, rows.Err()
}
```

- [ ] **Step 3: 编译验证**

```bash
cd modelcraft-backend && go build ./...
```

Expected: 编译成功。

- [ ] **Step 4: Commit**

```bash
git add modelcraft-backend/internal/app/modeldatabase/
git commit -m "feat(app): add ModelDatabaseAppService with register/update/unregister/list"
```

---

## Task 6: GraphQL Schema + 代码生成

**Files:**
- Create: `modelcraft-backend/api/graph/project/schema/database.graphql`

- [ ] **Step 1: 创建 GraphQL Schema 文件**

`modelcraft-backend/api/graph/project/schema/database.graphql`:

```graphql
enum DatabaseMode {
  SELF_HOSTED
  MANAGED
}

type ModelDatabase {
  id: ID!
  name: String!
  title: String!
  description: String!
  mode: DatabaseMode!
  createdAt: Time!
  updatedAt: Time!
}

type RawDatabase {
  name: String!
  isRegistered: Boolean!
}

input RegisterModelDatabaseInput {
  name: String!
  title: String!
  description: String
  mode: DatabaseMode!
}

input UpdateModelDatabaseInput {
  title: String
  description: String
  mode: DatabaseMode
}

union RegisterModelDatabaseResult = ModelDatabase | InvalidInput | ResourceNotFound

extend type Query {
  modelDatabases: [ModelDatabase!]!
  clusterRawDatabases: [RawDatabase!]!
}

extend type Mutation {
  registerModelDatabase(input: RegisterModelDatabaseInput!): RegisterModelDatabaseResult!
  updateModelDatabase(id: ID!, input: UpdateModelDatabaseInput!): ModelDatabase!
  unregisterModelDatabase(id: ID!): Boolean!
}
```

- [ ] **Step 2: 运行 gqlgen 代码生成**

```bash
cd modelcraft-backend && just generate-gql
```

Expected: 生成成功，`internal/interfaces/graphql/project/generated/` 目录更新，并在 `internal/interfaces/graphql/project/` 目录自动创建 `database.resolvers.go` 框架文件（包含空实现）。

- [ ] **Step 3: Commit**

```bash
git add modelcraft-backend/api/graph/project/schema/database.graphql modelcraft-backend/internal/interfaces/graphql/project/generated/ modelcraft-backend/internal/interfaces/graphql/project/database.resolvers.go
git commit -m "feat(graphql): add ModelDatabase schema and regenerate project GraphQL code"
```

---

## Task 7: GraphQL Resolver 实现

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/database.resolvers.go`
- Modify: `modelcraft-backend/internal/interfaces/graphql/project/resolver.go`

- [ ] **Step 1: 在 Resolver struct 中注册 AppService**

修改 `internal/interfaces/graphql/project/resolver.go`，在 `Resolver` struct 中添加字段：

```go
// Database management
ModelDatabaseAppService *modeldatabase.ModelDatabaseAppService
```

同时在文件顶部 import 中添加：
```go
appmodeldatabase "modelcraft/internal/app/modeldatabase"
```

将 `ModelDatabaseAppService` 字段类型改为：
```go
ModelDatabaseAppService *appmodeldatabase.ModelDatabaseAppService
```

- [ ] **Step 2: 实现 Resolver 方法**

将 `internal/interfaces/graphql/project/database.resolvers.go` 中自动生成的空方法替换为以下实现：

```go
package projectgraphql

import (
	"context"
	appmodeldatabase "modelcraft/internal/app/modeldatabase"
	domainmodeldatabase "modelcraft/internal/domain/modeldatabase"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
)

func (r *queryResolver) ModelDatabases(ctx context.Context) ([]*generated.ModelDatabase, error) {
	items, err := r.ModelDatabaseAppService.ListRegistered(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*generated.ModelDatabase, 0, len(items))
	for _, item := range items {
		result = append(result, modelDatabaseToGQL(item))
	}
	return result, nil
}

func (r *queryResolver) ClusterRawDatabases(ctx context.Context) ([]*generated.RawDatabase, error) {
	items, err := r.ModelDatabaseAppService.ListRaw(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*generated.RawDatabase, 0, len(items))
	for _, item := range items {
		result = append(result, &generated.RawDatabase{
			Name:         item.Name,
			IsRegistered: item.IsRegistered,
		})
	}
	return result, nil
}

func (r *mutationResolver) RegisterModelDatabase(ctx context.Context, input generated.RegisterModelDatabaseInput) (generated.RegisterModelDatabaseResult, error) {
	logger := logfacade.GetLogger(ctx)
	cmd := appmodeldatabase.RegisterCommand{
		Name:        input.Name,
		Title:       input.Title,
		Description: stringOrEmpty(input.Description),
		Mode:        gqlModeToDomain(input.Mode),
	}
	db, err := r.ModelDatabaseAppService.Register(ctx, cmd)
	if err != nil {
		if bizErr, ok := err.(*bizerrors.BusinessError); ok {
			logger.Error(ctx, "RegisterModelDatabase failed", logfacade.Err(bizErr))
			switch bizErr.Code {
			case bizerrors.AlreadyExists, bizerrors.ParamInvalid:
				return &generated.InvalidInput{Message: bizErr.Message}, nil
			case bizerrors.ResourceNotFound:
				return &generated.ResourceNotFound{Message: bizErr.Message, ResourceType: generated.ResourceTypeModelDatabase}, nil
			}
		}
		return nil, err
	}
	return modelDatabaseToGQL(db), nil
}

func (r *mutationResolver) UpdateModelDatabase(ctx context.Context, id string, input generated.UpdateModelDatabaseInput) (*generated.ModelDatabase, error) {
	mode := domainModeFromGQLPtr(input.Mode)
	cmd := appmodeldatabase.UpdateCommand{
		ID:          id,
		Title:       input.Title,
		Description: input.Description,
		Mode:        mode,
	}
	db, err := r.ModelDatabaseAppService.Update(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return modelDatabaseToGQL(db), nil
}

func (r *mutationResolver) UnregisterModelDatabase(ctx context.Context, id string) (bool, error) {
	err := r.ModelDatabaseAppService.Unregister(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

// --- helpers ---

func modelDatabaseToGQL(db *domainmodeldatabase.ModelDatabase) *generated.ModelDatabase {
	return &generated.ModelDatabase{
		ID:          db.ID,
		Name:        db.Name,
		Title:       db.Title,
		Description: db.Description,
		Mode:        domainModeToGQL(db.Mode),
		CreatedAt:   db.CreatedAt,
		UpdatedAt:   db.UpdatedAt,
	}
}

func domainModeToGQL(m domainmodeldatabase.DatabaseMode) generated.DatabaseMode {
	if m == domainmodeldatabase.DatabaseModeManaged {
		return generated.DatabaseModeManaged
	}
	return generated.DatabaseModeSelfHosted
}

func gqlModeToDomain(m generated.DatabaseMode) domainmodeldatabase.DatabaseMode {
	if m == generated.DatabaseModeManaged {
		return domainmodeldatabase.DatabaseModeManaged
	}
	return domainmodeldatabase.DatabaseModeSelfHosted
}

func domainModeFromGQLPtr(m *generated.DatabaseMode) *domainmodeldatabase.DatabaseMode {
	if m == nil {
		return nil
	}
	mode := gqlModeToDomain(*m)
	return &mode
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
```

> 注意：`gqlModeToomain` 函数名需要修正为 `gqlModeToDomain`（两处），确保名字一致。

- [ ] **Step 3: 编译验证**

```bash
cd modelcraft-backend && go build ./...
```

Expected: 编译成功。

- [ ] **Step 4: Commit**

```bash
git add modelcraft-backend/internal/interfaces/graphql/project/database.resolvers.go modelcraft-backend/internal/interfaces/graphql/project/resolver.go
git commit -m "feat(graphql): implement ModelDatabase resolvers"
```

---

## Task 8: 依赖注入注册

**Files:**
- Modify: `modelcraft-backend/internal/interfaces/http/routes.go`

- [ ] **Step 1: 在 DesignHandlers 中添加字段**

在 `routes.go` 的 `DesignHandlers` struct 中添加：

```go
// Database management
ModelDatabaseAppService *appmodeldatabase.ModelDatabaseAppService
```

同时在文件顶部 import 添加：
```go
appmodeldatabase "modelcraft/internal/app/modeldatabase"
```

- [ ] **Step 2: 在 CreateDesignHandlers 函数中初始化**

找到 `CreateDesignHandlers()` 函数，在初始化 `clusterRepository` 后添加：

```go
modelDatabaseRepo := repository.NewSqlModelDatabaseRepository(dbgen.New(loggingDB))
modelDatabaseAppService := appmodeldatabase.NewModelDatabaseAppService(
    modelDatabaseRepo,
    clusterRepository,
    clusterManager,
)
```

并在返回的 `DesignHandlers` struct 中添加：
```go
ModelDatabaseAppService: modelDatabaseAppService,
```

- [ ] **Step 3: 在 SetupProjectGraphQLRoutesOnChi 中注入 Resolver**

找到 `projectResolver := &projectgraphql.Resolver{...}` 的初始化，添加：
```go
ModelDatabaseAppService: handlers.ModelDatabaseAppService,
```

- [ ] **Step 4: 编译验证**

```bash
cd modelcraft-backend && go build ./...
```

Expected: 编译成功。

- [ ] **Step 5: Commit**

```bash
git add modelcraft-backend/internal/interfaces/http/routes.go
git commit -m "feat(di): wire ModelDatabaseAppService into project GraphQL resolver"
```

---

## Task 9: 前端 GraphQL 文档

**Files:**
- Create: `modelcraft-front/src/api-client/project/model-database-graphql-docs.ts`
- Modify: `modelcraft-front/src/api-client/project/index.ts`

- [ ] **Step 1: 创建 GraphQL 文档文件**

`modelcraft-front/src/api-client/project/model-database-graphql-docs.ts`:

```typescript
import { gql } from '@apollo/client'

export const MODEL_DATABASE_FRAGMENT = gql`
  fragment ModelDatabaseFields on ModelDatabase {
    id
    name
    title
    description
    mode
    createdAt
    updatedAt
  }
`

export const LIST_MODEL_DATABASES = gql`
  ${MODEL_DATABASE_FRAGMENT}
  query ListModelDatabases {
    modelDatabases {
      ...ModelDatabaseFields
    }
  }
`

export const LIST_CLUSTER_RAW_DATABASES = gql`
  query ListClusterRawDatabases {
    clusterRawDatabases {
      name
      isRegistered
    }
  }
`

export const REGISTER_MODEL_DATABASE = gql`
  ${MODEL_DATABASE_FRAGMENT}
  mutation RegisterModelDatabase($input: RegisterModelDatabaseInput!) {
    registerModelDatabase(input: $input) {
      ... on ModelDatabase {
        ...ModelDatabaseFields
      }
      ... on InvalidInput {
        message
      }
      ... on ResourceNotFound {
        message
        resourceType
      }
    }
  }
`

export const UPDATE_MODEL_DATABASE = gql`
  ${MODEL_DATABASE_FRAGMENT}
  mutation UpdateModelDatabase($id: ID!, $input: UpdateModelDatabaseInput!) {
    updateModelDatabase(id: $id, input: $input) {
      ...ModelDatabaseFields
    }
  }
`

export const UNREGISTER_MODEL_DATABASE = gql`
  mutation UnregisterModelDatabase($id: ID!) {
    unregisterModelDatabase(id: $id)
  }
`
```

- [ ] **Step 2: 在 index.ts 中导出**

在 `modelcraft-front/src/api-client/project/index.ts` 末尾添加：
```typescript
export * from './model-database-graphql-docs'
```

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/api-client/project/model-database-graphql-docs.ts modelcraft-front/src/api-client/project/index.ts
git commit -m "feat(api-client): add ModelDatabase GraphQL documents"
```

---

## Task 10: 前端 Hook

**Files:**
- Create: `modelcraft-front/src/web/hooks/model-database/use-model-databases.ts`

- [ ] **Step 1: 创建 Hook**

`modelcraft-front/src/web/hooks/model-database/use-model-databases.ts`:

```typescript
import { useMutation, useQuery } from '@apollo/client'
import {
  LIST_MODEL_DATABASES,
  LIST_CLUSTER_RAW_DATABASES,
  REGISTER_MODEL_DATABASE,
  UPDATE_MODEL_DATABASE,
  UNREGISTER_MODEL_DATABASE,
} from '@/api-client/project'
import { useProjectScopedClient } from '@api-client/apollo/public'

export type DatabaseMode = 'SELF_HOSTED' | 'MANAGED'

export interface ModelDatabase {
  id: string
  name: string
  title: string
  description: string
  mode: DatabaseMode
  createdAt: string
  updatedAt: string
}

export interface RawDatabase {
  name: string
  isRegistered: boolean
}

export interface RegisterModelDatabaseInput {
  name: string
  title: string
  description?: string
  mode: DatabaseMode
}

export interface UpdateModelDatabaseInput {
  title?: string
  description?: string
  mode?: DatabaseMode
}

// ── useModelDatabases: list registered databases ─────────────────────────────

export function useModelDatabases(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const { data, loading, error, refetch } = useQuery<{ modelDatabases: ModelDatabase[] }>(
    LIST_MODEL_DATABASES,
    { client, skip: !projectSlug }
  )
  return {
    databases: data?.modelDatabases ?? [],
    loading,
    error,
    refetch,
  }
}

// ── useClusterRawDatabases: list all databases from cluster ──────────────────

export function useClusterRawDatabases(projectSlug: string | null | undefined, skip?: boolean) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const { data, loading, error, refetch } = useQuery<{ clusterRawDatabases: RawDatabase[] }>(
    LIST_CLUSTER_RAW_DATABASES,
    { client, skip: !projectSlug || skip }
  )
  return {
    rawDatabases: data?.clusterRawDatabases ?? [],
    loading,
    error,
    refetch,
  }
}

// ── useRegisterModelDatabase ─────────────────────────────────────────────────

export function useRegisterModelDatabase(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const [mutate, { loading, error }] = useMutation(REGISTER_MODEL_DATABASE, {
    client,
    refetchQueries: [LIST_MODEL_DATABASES],
  })

  const register = async (input: RegisterModelDatabaseInput) => {
    const result = await mutate({ variables: { input } })
    const data = result.data?.registerModelDatabase
    if (data?.__typename === 'InvalidInput' || data?.__typename === 'ResourceNotFound') {
      throw new Error(data.message)
    }
    return data as ModelDatabase
  }

  return { register, loading, error }
}

// ── useUpdateModelDatabase ───────────────────────────────────────────────────

export function useUpdateModelDatabase(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const [mutate, { loading, error }] = useMutation(UPDATE_MODEL_DATABASE, {
    client,
    refetchQueries: [LIST_MODEL_DATABASES],
  })

  const update = async (id: string, input: UpdateModelDatabaseInput) => {
    const result = await mutate({ variables: { id, input } })
    return result.data?.updateModelDatabase as ModelDatabase
  }

  return { update, loading, error }
}

// ── useUnregisterModelDatabase ───────────────────────────────────────────────

export function useUnregisterModelDatabase(projectSlug: string | null | undefined) {
  const client = useProjectScopedClient(projectSlug ?? undefined)
  const [mutate, { loading }] = useMutation(UNREGISTER_MODEL_DATABASE, {
    client,
    refetchQueries: [LIST_MODEL_DATABASES],
  })

  const unregister = async (id: string) => {
    await mutate({ variables: { id } })
  }

  return { unregister, loading }
}
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-front/src/web/hooks/model-database/
git commit -m "feat(hooks): add useModelDatabases hooks"
```

---

## Task 11: 前端页面和组件

**Files:**
- Create: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/databases/page.tsx`
- Create: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/databases/_components/RegisterDatabaseDialog.tsx`
- Create: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/databases/_components/EditDatabaseSheet.tsx`

- [ ] **Step 1: 创建注册 Dialog**

`modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/databases/_components/RegisterDatabaseDialog.tsx`:

```tsx
'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@web/components/ui/dialog'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Textarea } from '@web/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@web/components/ui/select'
import { RadioGroup, RadioGroupItem } from '@web/components/ui/radio-group'
import { Loader2 } from 'lucide-react'
import {
  useClusterRawDatabases,
  useRegisterModelDatabase,
  type DatabaseMode,
} from '@web/hooks/model-database/use-model-databases'

interface RegisterDatabaseDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function RegisterDatabaseDialog({ open, onOpenChange }: RegisterDatabaseDialogProps) {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const { rawDatabases, loading: rawLoading } = useClusterRawDatabases(params.projectSlug, !open)
  const { register, loading: registering } = useRegisterModelDatabase(params.projectSlug)

  const unregistered = rawDatabases.filter((db) => !db.isRegistered)

  const [selectedName, setSelectedName] = useState('')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [mode, setMode] = useState<DatabaseMode>('SELF_HOSTED')
  const [error, setError] = useState('')

  const handleNameChange = (name: string) => {
    setSelectedName(name)
    if (!title || title === selectedName) {
      setTitle(name)
    }
  }

  const handleSubmit = async () => {
    if (!selectedName || !title) return
    setError('')
    try {
      await register({ name: selectedName, title, description: description || undefined, mode })
      onOpenChange(false)
      setSelectedName('')
      setTitle('')
      setDescription('')
      setMode('SELF_HOSTED')
    } catch (e) {
      setError(e instanceof Error ? e.message : '接管失败')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>接管数据库</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-4 py-2">
          <div className="flex flex-col gap-2">
            <Label>选择数据库</Label>
            {rawLoading ? (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="size-4 animate-spin" /> 加载中...
              </div>
            ) : (
              <Select value={selectedName} onValueChange={handleNameChange}>
                <SelectTrigger>
                  <SelectValue placeholder="选择要接管的数据库" />
                </SelectTrigger>
                <SelectContent>
                  {unregistered.length === 0 ? (
                    <div className="px-2 py-3 text-center text-sm text-muted-foreground">
                      所有数据库已接管
                    </div>
                  ) : (
                    unregistered.map((db) => (
                      <SelectItem key={db.name} value={db.name}>
                        {db.name}
                      </SelectItem>
                    ))
                  )}
                </SelectContent>
              </Select>
            )}
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="db-title">友好名称</Label>
            <Input
              id="db-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="数据库显示名称"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="db-description">描述（可选）</Label>
            <Textarea
              id="db-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="简要描述此数据库的用途"
              rows={2}
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label>访问模式</Label>
            <RadioGroup value={mode} onValueChange={(v) => setMode(v as DatabaseMode)}>
              <div className="flex items-start gap-3 rounded-md border border-border p-3">
                <RadioGroupItem value="SELF_HOSTED" id="mode-self" className="mt-0.5" />
                <div>
                  <label htmlFor="mode-self" className="cursor-pointer text-sm font-medium">
                    自建
                  </label>
                  <p className="text-xs text-muted-foreground">可读写，支持新建和导入模型</p>
                </div>
              </div>
              <div className="flex items-start gap-3 rounded-md border border-border p-3">
                <RadioGroupItem value="MANAGED" id="mode-managed" className="mt-0.5" />
                <div>
                  <label htmlFor="mode-managed" className="cursor-pointer text-sm font-medium">
                    托管
                  </label>
                  <p className="text-xs text-muted-foreground">只读，仅支持同步模型</p>
                </div>
              </div>
            </RadioGroup>
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={registering}>
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!selectedName || !title || registering}
          >
            {registering && <Loader2 className="mr-2 size-4 animate-spin" />}
            确认接管
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 2: 创建编辑 Sheet**

`modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/databases/_components/EditDatabaseSheet.tsx`:

```tsx
'use client'

import { useState, useEffect } from 'react'
import { useParams } from 'next/navigation'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetFooter,
} from '@web/components/ui/sheet'
import { Button } from '@web/components/ui/button'
import { Input } from '@web/components/ui/input'
import { Label } from '@web/components/ui/label'
import { Textarea } from '@web/components/ui/textarea'
import { RadioGroup, RadioGroupItem } from '@web/components/ui/radio-group'
import { Loader2 } from 'lucide-react'
import {
  useUpdateModelDatabase,
  type ModelDatabase,
  type DatabaseMode,
} from '@web/hooks/model-database/use-model-databases'

interface EditDatabaseSheetProps {
  database: ModelDatabase | null
  onClose: () => void
}

export function EditDatabaseSheet({ database, onClose }: EditDatabaseSheetProps) {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const { update, loading } = useUpdateModelDatabase(params.projectSlug)

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [mode, setMode] = useState<DatabaseMode>('SELF_HOSTED')

  useEffect(() => {
    if (database) {
      setTitle(database.title)
      setDescription(database.description)
      setMode(database.mode)
    }
  }, [database])

  const handleSave = async () => {
    if (!database) return
    await update(database.id, { title, description, mode })
    onClose()
  }

  return (
    <Sheet open={!!database} onOpenChange={(open) => !open && onClose()}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>编辑数据库</SheetTitle>
        </SheetHeader>

        <div className="flex flex-col gap-4 py-4">
          <div className="flex flex-col gap-2">
            <Label className="text-muted-foreground">数据库名</Label>
            <p className="text-sm font-medium">{database?.name}</p>
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="edit-title">友好名称</Label>
            <Input id="edit-title" value={title} onChange={(e) => setTitle(e.target.value)} />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="edit-description">描述</Label>
            <Textarea
              id="edit-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label>访问模式</Label>
            <RadioGroup value={mode} onValueChange={(v) => setMode(v as DatabaseMode)}>
              <div className="flex items-start gap-3 rounded-md border border-border p-3">
                <RadioGroupItem value="SELF_HOSTED" id="edit-mode-self" className="mt-0.5" />
                <div>
                  <label htmlFor="edit-mode-self" className="cursor-pointer text-sm font-medium">
                    自建
                  </label>
                  <p className="text-xs text-muted-foreground">可读写，支持新建和导入模型</p>
                </div>
              </div>
              <div className="flex items-start gap-3 rounded-md border border-border p-3">
                <RadioGroupItem value="MANAGED" id="edit-mode-managed" className="mt-0.5" />
                <div>
                  <label htmlFor="edit-mode-managed" className="cursor-pointer text-sm font-medium">
                    托管
                  </label>
                  <p className="text-xs text-muted-foreground">只读，仅支持同步模型</p>
                </div>
              </div>
            </RadioGroup>
          </div>
        </div>

        <SheetFooter>
          <Button variant="outline" onClick={onClose} disabled={loading}>
            取消
          </Button>
          <Button onClick={handleSave} disabled={!title || loading}>
            {loading && <Loader2 className="mr-2 size-4 animate-spin" />}
            保存
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
```

- [ ] **Step 3: 创建数据库管理主页面**

`modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/databases/page.tsx`:

```tsx
'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import { Plus, Pencil, MoreHorizontal, Trash2 } from 'lucide-react'
import { Button } from '@web/components/ui/button'
import { Badge } from '@web/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@web/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@web/components/ui/dropdown-menu'
import {
  useModelDatabases,
  useUnregisterModelDatabase,
  type ModelDatabase,
} from '@web/hooks/model-database/use-model-databases'
import { RegisterDatabaseDialog } from './_components/RegisterDatabaseDialog'
import { EditDatabaseSheet } from './_components/EditDatabaseSheet'

export default function DatabasesPage() {
  const params = useParams<{ orgName: string; projectSlug: string }>()
  const { databases, loading } = useModelDatabases(params.projectSlug)
  const { unregister } = useUnregisterModelDatabase(params.projectSlug)

  const [registerOpen, setRegisterOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<ModelDatabase | null>(null)

  return (
    <div className="flex flex-col gap-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">数据库管理</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            接管此项目使用的 MySQL 数据库，设置访问模式
          </p>
        </div>
        <Button onClick={() => setRegisterOpen(true)} size="sm">
          <Plus className="mr-2 size-4" />
          接管数据库
        </Button>
      </div>

      <div className="rounded-md border border-border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>名称</TableHead>
              <TableHead>描述</TableHead>
              <TableHead>模式</TableHead>
              <TableHead className="w-16" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={4} className="py-8 text-center text-sm text-muted-foreground">
                  加载中...
                </TableCell>
              </TableRow>
            ) : databases.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="py-8 text-center text-sm text-muted-foreground">
                  暂无已接管的数据库，点击右上角"接管数据库"开始
                </TableCell>
              </TableRow>
            ) : (
              databases.map((db) => (
                <TableRow key={db.id}>
                  <TableCell>
                    <div className="flex flex-col">
                      <span className="font-medium">{db.title}</span>
                      {db.title !== db.name && (
                        <span className="text-xs text-muted-foreground">{db.name}</span>
                      )}
                    </div>
                  </TableCell>
                  <TableCell className="max-w-xs truncate text-sm text-muted-foreground">
                    {db.description || '—'}
                  </TableCell>
                  <TableCell>
                    {db.mode === 'SELF_HOSTED' ? (
                      <Badge variant="outline" className="border-green-500/30 bg-green-500/10 text-green-700 dark:text-green-400">
                        自建
                      </Badge>
                    ) : (
                      <Badge variant="outline" className="border-blue-500/30 bg-blue-500/10 text-blue-700 dark:text-blue-400">
                        托管
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-7"
                        onClick={() => setEditTarget(db)}
                      >
                        <Pencil className="size-3.5" />
                      </Button>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="size-7">
                            <MoreHorizontal className="size-3.5" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            className="text-destructive focus:text-destructive"
                            onClick={() => unregister(db.id)}
                          >
                            <Trash2 className="mr-2 size-3.5" />
                            取消接管
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <RegisterDatabaseDialog open={registerOpen} onOpenChange={setRegisterOpen} />
      <EditDatabaseSheet database={editTarget} onClose={() => setEditTarget(null)} />
    </div>
  )
}
```

- [ ] **Step 4: Commit**

```bash
git add "modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/databases/"
git commit -m "feat(ui): add databases management page with register and edit dialogs"
```

---

## Task 12: 导航栏 + ModelSidebar 改动

**Files:**
- Modify: `modelcraft-front/src/web/components/features/layout/AppLayout.tsx`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx`
- Modify: `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_hooks/use-model-crud.ts`

- [ ] **Step 1: AppLayout 导航栏加入"数据库"菜单项**

在 `AppLayout.tsx` 的 `projectNavSections` 中，找到 `数据建模` section，在 `数据模型` 之后、`枚举管理` 之前插入：

```typescript
{ label: '数据库', icon: '/icons/icon-storage.svg', href: `/org/${orgName}/project/${projectSlug}/databases` },
```

> 注意：如果 `icon-storage.svg` 不存在，可先用 `/icons/icon-list.svg` 占位。

- [ ] **Step 2: ModelSidebar 数据库列表改用已接管列表**

在 `use-model-crud.ts` 中，将 `useDatabases` 替换为 `useModelDatabases`：

找到第 92-95 行（`useDatabases` 调用）替换为：

```typescript
import { useModelDatabases } from '@web/hooks/model-database/use-model-databases'

// 替换原来的 useDatabases 调用：
const { databases: rawModelDatabases, loading: databasesLoading } = useModelDatabases(
  !state.connectionChecking && !state.connectionFailed ? projectSlug : null
)
// 将 ModelDatabase[] 转为 { name: string }[] 格式（保持下游兼容）
const databases = rawModelDatabases.map((db) => ({ name: db.name, mode: db.mode }))
```

- [ ] **Step 3: ModelSidebar 在 managed 模式下隐藏新建/导入按钮**

在 `ModelSidebar.tsx` 中，找到 Zone 2 的 Action buttons 部分（第 158-190 行），修改新建/导入按钮的条件，加入 managed 模式判断。

先找到 `databases` prop 的类型，在文件顶部或 props 定义处，将 `databases` 的元素类型从 `{ name: string }` 扩展为 `{ name: string; mode?: string }`。

然后在 Action buttons 区域，在 `viewMode === 'schema'` 条件的基础上，计算当前选中数据库的模式：

```tsx
// 找到当前选中数据库的 mode
const selectedDbMode = databases.find((db) => db.name === state.selectedDatabase)?.mode

// 新建模型和导入模型按钮只在 self_hosted 或未设置 mode 时显示
const canWrite = !selectedDbMode || selectedDbMode === 'SELF_HOSTED'
```

将 Zone 2 的 Action buttons 改为：

```tsx
{viewMode === 'schema' && (
  <div className="flex flex-col gap-1 px-3 py-2.5">
    {canWrite && (
      <>
        <Button
          ref={createModelBtnRef}
          size="sm"
          variant="outline"
          className={cn(
            'h-7 w-full justify-start px-2.5 text-xs font-normal transition-colors',
            !state.selectedDatabase && 'pointer-events-none opacity-40',
            pendingAction === 'nav_create_model' && state.selectedDatabase && 'ring-2 ring-amber-400 ring-offset-1 animate-pulse border-amber-400'
          )}
          onClick={handleCreateModelClick}
          disabled={!state.selectedDatabase}
        >
          <Plus className="mr-1 size-3.5" />
          新建模型
        </Button>
        <Button
          size="sm"
          variant="outline"
          className={cn(
            'h-7 w-full justify-start px-2.5 text-xs font-normal transition-colors',
            !state.selectedDatabase && 'pointer-events-none opacity-40'
          )}
          onClick={() => state.setImportDialogOpen(true)}
          disabled={!state.selectedDatabase}
        >
          <Download className="mr-1 size-3.5" strokeWidth={1.5} />
          导入模型
        </Button>
      </>
    )}
    {!canWrite && state.selectedDatabase && (
      <Button
        size="sm"
        variant="outline"
        className="h-7 w-full justify-start px-2.5 text-xs font-normal opacity-50"
        disabled
      >
        同步模型（即将推出）
      </Button>
    )}
  </div>
)}
```

- [ ] **Step 4: 编译 lint 检查**

```bash
cd modelcraft-front && npm run lint
```

Expected: 无报错。

- [ ] **Step 5: Commit**

```bash
git add modelcraft-front/src/web/components/features/layout/AppLayout.tsx
git add "modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_components/ModelSidebar.tsx"
git add "modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/model-editor/_hooks/use-model-crud.ts"
git commit -m "feat(ui): add databases nav item, restrict ModelSidebar to registered databases, hide write actions for managed mode"
```

---

## 验收检查

完成所有 Task 后，按以下顺序手动验证：

1. **后端启动**：`just dev`，无报错
2. **GraphQL Playground**：访问 `/graphql/org/{orgName}/project/{projectSlug}/`，执行 `{ modelDatabases { id name title mode } }` 应返回空数组
3. **接管数据库**：执行 `registerModelDatabase` mutation，确认数据写入并返回
4. **前端页面**：访问 `/org/{orgName}/project/{projectSlug}/databases`，页面正常展示
5. **接管 Dialog**：打开后下拉显示未接管的数据库列表，提交后列表刷新
6. **ModelSidebar**：切换到 model-editor，数据库选择器只显示已接管的 database
7. **托管模式**：将某个 database 设为 managed，ModelSidebar 中该 database 被选中时，新建/导入按钮隐藏
