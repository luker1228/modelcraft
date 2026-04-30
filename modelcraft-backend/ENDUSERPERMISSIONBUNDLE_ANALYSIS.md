# ModelCraft Backend: `endUserPermissionBundle` Stack Analysis

## Overview
Complete tracing of the `endUserPermissionBundle` GraphQL query implementation through all layers, from GraphQL schema → resolver → application service → domain repository interface → SQL infrastructure implementation.

---

## 1. GraphQL Schema Definition

**File:** `/data/home/lukemxjia/modelcraft/modelcraft-backend/api/graph/project/schema/rbac.graphql`

### Query Definition (Line 745)
```graphql
extend type Query {
  endUserPermissionBundle(id: ID!): EndUserPermissionBundle @hasPermission(action: "rbac:read")
}
```

### Type Definition (Lines 157-183)
```graphql
type EndUserPermissionBundle implements Node {
  id: ID!
  """
  URL 友好的对外标识符，同项目内唯一，创建时由用户指定或从名称自动派生，之后不可修改。
  """
  slug: String!
  name: String!
  description: String
  """
  Item-centric 数据权限列表：每个模型最多一个 item。
  """
  dataPermissionItems: [EndUserBundleDataPermissionItem!]!
  """
  兼容旧字段（将逐步废弃），从 item 导出的 permission 视图。
  """
  permissions: [EndUserBundlePermissionEntry!]!
  """
  当前版本号（每次权限列表变更后递增）。初始创建时为 0，首次修改后变为 1。
  """
  currentVersion: Int!
  """
  最近历史快照列表（最多 5 个，按 version DESC 排列）
  """
  snapshots: [EndUserPermissionBundleSnapshot!]!
  createdAt: Time!
  updatedAt: Time!
}
```

### Key Related Types

**EndUserBundleDataPermissionItem** (Lines 189-204)
```graphql
type EndUserBundleDataPermissionItem {
  id: ID!
  bundleId: ID!
  modelId: ID!
  grantType: DataPermissionGrantType!
  preset: EndUserPermissionPreset
  customPermissionId: ID
  customPermission: EndUserPermission
  sortOrder: Int!
  createdAt: Time!
  updatedAt: Time!
}
```

---

## 2. GraphQL Resolver Implementation

**File:** `/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/interfaces/graphql/project/rbac.resolvers.go`

### Query Resolver (Lines 653-665)
```go
// EndUserPermissionBundle is the resolver for the endUserPermissionBundle field.
func (r *queryResolver) EndUserPermissionBundle(ctx context.Context, id string) (*generated.EndUserPermissionBundle, error) {
  orgName, projectSlug, err := getOrgAndProjectFromContext(ctx)
  if err != nil {
    return nil, err
  }
  bundle, appErr := r.RBACBundleSvc.GetBundleByID(ctx, orgName, projectSlug, id)
  if appErr != nil {
    logfacade.GetLogger(ctx).Error(ctx, "rbac operation failed", logfacade.Err(appErr), logfacade.Stack(appErr))
    return nil, appErr
  }
  return adapter.ToEndUserPermissionBundleDTO(bundle), nil
}
```

**Key Points:**
- Extracts `orgName` and `projectSlug` from GraphQL context
- Calls application service: `RBACBundleSvc.GetBundleByID()`
- Converts domain entity to GraphQL DTO via adapter
- Handles errors through logging

---

## 3. Application Layer Service

**File:** `/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/app/rbac/bundle_app.go`

### Service Type (Lines 49-63)
```go
type EndUserBundleAppService struct {
  rbacRepo  bundleRepository
  modelRepo modelRepository
}

func NewEndUserBundleAppService(
  rbacRepo rbacdomain.EndUserPermissionRepository,
  modelRepo modeldesign.ModelRepository,
) *EndUserBundleAppService {
  return &EndUserBundleAppService{
    rbacRepo:  rbacRepo,
    modelRepo: modelRepo,
  }
}
```

### GetBundleByID Method (Lines 129-191)
```go
func (s *EndUserBundleAppService) GetBundleByID(
  ctx context.Context,
  orgName, projectSlug, id string,
) (*rbacdomain.EndUserPermissionBundle, error) {
  // 1. Get bundle by ID from repository
  bundle, err := s.rbacRepo.GetBundleByID(ctx, orgName, projectSlug, id)
  if err != nil {
    if shared.IsNotFoundError(err) {
      return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionBundleNotFound, id)
    }
    return nil, bizerrors.ConvertRepositoryError(ctx, err)
  }

  // 2. Load bundle data permission items
  items, err := s.rbacRepo.ListBundleDataPermissionItems(ctx, id)
  if err != nil {
    return nil, bizerrors.ConvertRepositoryError(ctx, err)
  }
  bundle.Items = items

  // 3. Build legacy permissions view (for backward compatibility)
  legacyPermissions := make([]*rbacdomain.EndUserPermission, 0, len(items))
  for _, item := range items {
    // Handle CUSTOM grant type
    if item.GrantType == rbacdomain.PermissionTypeCustom {
      if item.CustomPermissionID != nil && *item.CustomPermissionID != "" {
        p, getErr := s.rbacRepo.GetPermissionByID(ctx, orgName, *item.CustomPermissionID)
        if getErr == nil {
          legacyPermissions = append(legacyPermissions, p)
        }
      }
    }
    // Handle PRESET grant type
    if item.GrantType == rbacdomain.PermissionTypePreset {
      if item.Preset != nil {
        ownerField, _ := s.tryGetOwnerField(ctx, item.ModelID)
        rowPolicy, expandErr := expandPreset(*item.Preset, ownerField)
        if expandErr != nil {
          continue
        }
        // Create synthetic permission from preset
        presetCopy := *item.Preset
        legacyPermissions = append(legacyPermissions, &rbacdomain.EndUserPermission{
          ID:           fmt.Sprintf("preset:%s:%s:%s", bundle.ID, item.ModelID, presetCopy),
          OrgName:      bundle.OrgName,
          ProjectSlug:  bundle.ProjectSlug,
          ModelID:      item.ModelID,
          Name:         presetPermissionName(presetCopy),
          Type:         rbacdomain.PermissionTypePreset,
          Preset:       &presetCopy,
          ColumnPolicy: nil,
          RowPolicy:    rowPolicy,
        })
      }
    }
  }
  bundle.Permissions = legacyPermissions

  // 4. Load bundle snapshots (history)
  snapshots, err := s.rbacRepo.ListBundleSnapshots(ctx, id)
  if err != nil {
    return nil, bizerrors.ConvertRepositoryError(ctx, err)
  }
  bundle.Snapshots = snapshots
  
  return bundle, nil
}
```

**Key Operations:**
1. Fetch bundle by ID from repository
2. Load associated data permission items
3. Build legacy permissions view (for backward compatibility)
4. Load bundle snapshots (version history)
5. Return enriched bundle entity

---

## 4. Domain Repository Interface

**File:** `/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/domain/rbac/repository.go`

### Interface Definition (Lines 6-170)

**Relevant Methods:**

```go
type EndUserPermissionRepository interface {
  // GetBundleByID 根据 ID 获取权限包（org + project scoped）
  GetBundleByID(ctx context.Context, orgName, projectSlug, id string) (*EndUserPermissionBundle, error)

  // ListBundlesByProject 列出项目下所有权限包（org + project scoped）
  ListBundlesByProject(ctx context.Context, orgName, projectSlug string) ([]*EndUserPermissionBundle, error)

  // ListBundleDataPermissionItems 列出 bundle 中的 item。
  ListBundleDataPermissionItems(ctx context.Context, bundleID string) ([]*EndUserBundleDataPermissionItem, error)

  // SaveBundleSnapshot 写入权限包快照
  SaveBundleSnapshot(ctx context.Context, snapshot *BundleSnapshot) error

  // ListBundleSnapshots 列出权限包最近 5 个历史快照（按 version DESC）
  ListBundleSnapshots(ctx context.Context, bundleID string) ([]BundleSnapshot, error)

  // GetBundleCurrentVersion 获取权限包当前最大版本号
  GetBundleCurrentVersion(ctx context.Context, bundleID string) (int, error)
}
```

**Scope Pattern:**
- All bundle queries are **org + project scoped** (prevent cross-project/cross-org data leakage)
- Bundle ID lookup requires both `orgName` and `projectSlug` parameters
- This matches GraphQL context extraction in the resolver

---

## 5. Infrastructure Repository Implementation

**File:** `/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/infrastructure/repository/sql_end_user_permission_repository.go`

### Type Definition (Lines 20-28)
```go
type SqlEndUserDataPermissionRepository struct {
  q dbgen.Querier  // sqlc-generated query interface
}

func NewSqlEndUserDataPermissionRepository(q dbgen.Querier) rbac.EndUserPermissionRepository {
  return &SqlEndUserDataPermissionRepository{q: dbgenwrap.NewSafeQuerier(q)}
}
```

### GetBundleByID Implementation (Lines 329-346)
```go
func (r *SqlEndUserDataPermissionRepository) GetBundleByID(
  ctx context.Context,
  orgName, projectSlug, id string,
) (*rbac.EndUserPermissionBundle, error) {
  row, err := r.q.GetEndUserBundleByID(ctx, dbgen.GetEndUserBundleByIDParams{
    ID:          id,
    OrgName:     orgName,
    ProjectSlug: projectSlug,
  })
  if err != nil {
    if sqlerr.IsNotFoundError(err) {
      return nil, shared.NewNotFoundError("end user bundle not found: " + id)
    }
    return nil, err
  }

  return toDomainBundle(row), nil
}
```

### ListBundleDataPermissionItems Implementation (Lines 511-528)
```go
func (r *SqlEndUserDataPermissionRepository) ListBundleDataPermissionItems(
  ctx context.Context,
  bundleID string,
) ([]*rbac.EndUserBundleDataPermissionItem, error) {
  rows, err := r.q.ListBundleDataPermissionItems(ctx, bundleID)
  if err != nil {
    return nil, sqlerr.WrapSQLError(err)
  }

  items := make([]*rbac.EndUserBundleDataPermissionItem, 0, len(rows))
  for _, row := range rows {
    items = append(items, toDomainBundleDataPermissionItem(row))
  }
  return items, nil
}
```

### Helper: toDomainBundle (Lines 299-314)
```go
func toDomainBundle(row dbgen.EndUserPermissionBundle) *rbac.EndUserPermissionBundle {
  var description *string
  if row.Description.Valid {
    d := row.Description.String
    description = &d
  }

  return &rbac.EndUserPermissionBundle{
    OrgName:     row.OrgName,
    ProjectSlug: row.ProjectSlug,
    ID:          row.ID,
    Slug:        row.Slug,
    Name:        row.Name,
    Description: description,
  }
}
```

---

## 6. SQL Query Definition (sqlc)

**File:** `/data/home/lukemxjia/modelcraft/modelcraft-backend/db/queries/rbac/bundle.sql`

### GetEndUserBundleByID Query (Lines 12-17)
```sql
-- name: GetEndUserBundleByID :one
SELECT *
FROM end_user_permission_bundles
WHERE id = ?
  AND org_name = ?
  AND project_slug = ?;
```

### ListEndUserBundlesByProject Query (Lines 19-24)
```sql
-- name: ListEndUserBundlesByProject :many
SELECT *
FROM end_user_permission_bundles
WHERE org_name = ?
  AND project_slug = ?
ORDER BY name;
```

### ListBundleDataPermissionItems Query (Lines 67-71)
```sql
-- name: ListBundleDataPermissionItems :many
SELECT *
FROM end_user_bundle_data_permission_items
WHERE bundle_id = ?
ORDER BY sort_order, created_at;
```

---

## 7. Database Schema

**File:** `/data/home/lukemxjia/modelcraft/modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql`

### end_user_permission_bundles Table (Lines 70-90)
```sql
CREATE TABLE `end_user_permission_bundles` (
  `id`            VARCHAR(36)   NOT NULL                   COMMENT '权限包 UUID',
  `slug`          VARCHAR(64)   NOT NULL                   COMMENT 'URL 友好标识符，同项目内唯一',
  `org_name`      VARCHAR(64)  NOT NULL                    COMMENT '所属组织',
  `project_slug`  VARCHAR(64)  NOT NULL                    COMMENT '所属项目',
  `name`          VARCHAR(128) NOT NULL                    COMMENT '权限包名称',
  `description`   TEXT         NULL                        COMMENT '权限包描述',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  -- 同一项目下 slug 唯一（对外标识符）
  UNIQUE KEY `uq_bundles_org_project_slug`
    (`org_name`, `project_slug`, `slug`),
  -- 同一项目下权限包名称唯一
  UNIQUE KEY `uq_bundles_org_project_name`
    (`org_name`, `project_slug`, `name`),
  -- 快速检索
  INDEX `idx_bundles_org_project` (`org_name`, `project_slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### end_user_bundle_data_permission_items Table (Lines 98-138)
```sql
CREATE TABLE `end_user_bundle_data_permission_items` (
  `id`                   VARCHAR(36)  NOT NULL,
  `bundle_id`            VARCHAR(36)  NOT NULL,
  `model_id`             VARCHAR(36)  NOT NULL,
  `grant_type`           ENUM('PRESET','CUSTOM') NOT NULL,
  `preset`               ENUM(
                          'READ_WRITE_ALL',
                          'READ_ALL',
                          'READ_WRITE_OWNER',
                          'READ_ALL_WRITE_OWNER'
                        )            NULL DEFAULT NULL,
  `custom_permission_id` VARCHAR(36)  NULL DEFAULT NULL,
  `sort_order`           INT          NOT NULL DEFAULT 0,
  `created_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  -- 同一 bundle 下同一 model 最多一个 item
  UNIQUE KEY `uq_bundle_items_bundle_model`
    (`bundle_id`, `model_id`),
  INDEX `idx_bundle_items_custom_permission` (`custom_permission_id`),
  INDEX `idx_bundle_items_model_id` (`model_id`),
  CONSTRAINT `fk_bundle_items_bundle`
    FOREIGN KEY (`bundle_id`) REFERENCES `end_user_permission_bundles` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE,
  -- Checks ensure PRESET and CUSTOM constraints
  CONSTRAINT `chk_bundle_items_preset`
    CHECK (grant_type != 'PRESET' OR (preset IS NOT NULL AND custom_permission_id IS NULL)),
  CONSTRAINT `chk_bundle_items_custom`
    CHECK (grant_type != 'CUSTOM' OR (custom_permission_id IS NOT NULL AND preset IS NULL))
);
```

---

## 8. Full Request Flow Diagram

```
GraphQL Query: endUserPermissionBundle(id: "bundle-123")
       ↓
    Resolver (rbac.resolvers.go)
    - Extract orgName, projectSlug from context
    - Call RBACBundleSvc.GetBundleByID()
       ↓
    Application Service (bundle_app.go)
    - GetBundleByID(ctx, orgName, projectSlug, id)
    - Call rbacRepo.GetBundleByID()
    - Call rbacRepo.ListBundleDataPermissionItems()
    - Build legacy permissions view
    - Call rbacRepo.ListBundleSnapshots()
    - Return enriched bundle
       ↓
    Domain Repository Interface (repository.go)
    - Defines abstract contract
       ↓
    SQL Repository Implementation (sql_end_user_permission_repository.go)
    - Execute dbgen.GetEndUserBundleByID(id, orgName, projectSlug)
    - Execute dbgen.ListBundleDataPermissionItems(bundleID)
    - Convert DB rows to domain entities via toDomainBundle()
       ↓
    sqlc Generated Queries (bundle.sql)
    - SELECT * FROM end_user_permission_bundles WHERE id=? AND org_name=? AND project_slug=?
    - SELECT * FROM end_user_bundle_data_permission_items WHERE bundle_id=?
       ↓
    MySQL Database
    - Return bundle row
    - Return data permission item rows
       ↓
    Response
    - Convert domain entities to GraphQL DTO
    - Return EndUserPermissionBundle type with all fields populated
```

---

## 9. Key Patterns & Design Decisions

### 9.1 Org + Project Scoping
- **Why:** Multi-tenant architecture requires strict data isolation
- **Implementation:** All repository queries include both `orgName` and `projectSlug` parameters
- **Benefit:** Prevents cross-project/cross-org data leakage

### 9.2 Item-Centric Model
- **Why:** Allows flexible permission composition (PRESET or CUSTOM per model)
- **Database:** `end_user_bundle_data_permission_items` table with composite unique key `(bundle_id, model_id)`
- **Schema Guarantee:** At most one item per bundle per model

### 9.3 Slug Field for External Reference
- **Schema:** Unique key `(org_name, project_slug, slug)`
- **Immutability:** Slug is set at creation and never modified
- **Use Case:** Enable URL-friendly API paths or external references

### 9.4 Legacy Permissions Field
- **Purpose:** Backward compatibility with older permission-based APIs
- **Implementation:** Application layer builds synthetic permissions from items
- **Status:** "Will be gradually deprecated" per schema comment

### 9.5 Snapshot System
- **Purpose:** Track version history for bundle auditing and rollback
- **Retention:** Keep latest 5 snapshots (auto-purge older ones)
- **Versioning:** Version incremented on each item change

### 9.6 Error Handling Strategy
- **Repository Layer:** Returns `shared.NotFoundError` for missing records
- **Application Layer:** Converts to business errors (`bizerrors.EndUserPermissionBundleNotFound`)
- **Resolver Layer:** Returns GraphQL error to client

---

## 10. Required Implementation for `endUserPermissionBundleBySlug`

To add a new query `endUserPermissionBundleBySlug(slug: String!, orgName: String!)`, you need to:

### 10.1 GraphQL Schema Changes
Add to `api/graph/project/schema/rbac.graphql`:
```graphql
extend type Query {
  endUserPermissionBundleBySlug(slug: String!, orgName: String!): EndUserPermissionBundle @hasPermission(action: "rbac:read")
}
```

### 10.2 Resolver Implementation
Add to `internal/interfaces/graphql/project/rbac.resolvers.go`:
```go
func (r *queryResolver) EndUserPermissionBundleBySlug(ctx context.Context, slug, orgName string) (*generated.EndUserPermissionBundle, error) {
  projectSlug, err := getProjectFromContext(ctx)  // Extract from context
  if err != nil {
    return nil, err
  }
  bundle, appErr := r.RBACBundleSvc.GetBundleBySlug(ctx, orgName, projectSlug, slug)
  if appErr != nil {
    logfacade.GetLogger(ctx).Error(ctx, "rbac operation failed", logfacade.Err(appErr), logfacade.Stack(appErr))
    return nil, appErr
  }
  return adapter.ToEndUserPermissionBundleDTO(bundle), nil
}
```

### 10.3 Application Service Method
Add to `internal/app/rbac/bundle_app.go`:
```go
func (s *EndUserBundleAppService) GetBundleBySlug(
  ctx context.Context,
  orgName, projectSlug, slug string,
) (*rbacdomain.EndUserPermissionBundle, error) {
  bundle, err := s.rbacRepo.GetBundleBySlug(ctx, orgName, projectSlug, slug)
  if err != nil {
    if shared.IsNotFoundError(err) {
      return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserPermissionBundleNotFound, slug)
    }
    return nil, bizerrors.ConvertRepositoryError(ctx, err)
  }

  // Load items and snapshots (same as GetBundleByID)
  items, err := s.rbacRepo.ListBundleDataPermissionItems(ctx, bundle.ID)
  if err != nil {
    return nil, bizerrors.ConvertRepositoryError(ctx, err)
  }
  bundle.Items = items

  // ... build legacy permissions ...

  snapshots, err := s.rbacRepo.ListBundleSnapshots(ctx, bundle.ID)
  if err != nil {
    return nil, bizerrors.ConvertRepositoryError(ctx, err)
  }
  bundle.Snapshots = snapshots
  
  return bundle, nil
}
```

### 10.4 Domain Repository Interface
Add to `internal/domain/rbac/repository.go`:
```go
// GetBundleBySlug 根据 slug 获取权限包（org + project scoped）
GetBundleBySlug(ctx context.Context, orgName, projectSlug, slug string) (*EndUserPermissionBundle, error)
```

### 10.5 SQL Repository Implementation
Add to `internal/infrastructure/repository/sql_end_user_permission_repository.go`:
```go
func (r *SqlEndUserDataPermissionRepository) GetBundleBySlug(
  ctx context.Context,
  orgName, projectSlug, slug string,
) (*rbac.EndUserPermissionBundle, error) {
  row, err := r.q.GetEndUserBundleBySlug(ctx, dbgen.GetEndUserBundleBySlugParams{
    Slug:        slug,
    OrgName:     orgName,
    ProjectSlug: projectSlug,
  })
  if err != nil {
    if sqlerr.IsNotFoundError(err) {
      return nil, shared.NewNotFoundError("end user bundle not found: " + slug)
    }
    return nil, err
  }

  return toDomainBundle(row), nil
}
```

### 10.6 SQL Query Definition
Add to `db/queries/rbac/bundle.sql`:
```sql
-- name: GetEndUserBundleBySlug :one
SELECT *
FROM end_user_permission_bundles
WHERE slug = ?
  AND org_name = ?
  AND project_slug = ?;
```

---

## 11. Lessons from Existing Implementation

1. **Always maintain org + project scope** - database queries include both
2. **Convert DB rows to domain entities** - keep infrastructure concerns separate
3. **Handle `NOT FOUND` errors explicitly** - return business errors, not repository errors
4. **Load related collections in app layer** - repositories should be responsible for basic CRUD
5. **Use sqlc for type-safe queries** - generated code prevents SQL injection
6. **Maintain backward compatibility** - legacy fields/views stay until fully deprecated
7. **Use adapters for type conversion** - keeps GraphQL schema decoupled from domain model

