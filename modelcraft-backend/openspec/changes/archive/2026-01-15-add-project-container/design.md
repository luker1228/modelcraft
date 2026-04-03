# Design: Project Container Architecture

## Architecture Overview

The Project container introduces a new top-level organizational entity that groups related Clusters, Models, and Enums into logical workspaces.

```
┌────────────────────────────────────────────────────────┐
│                      Project                           │
│  ┌──────────────────────────────────────────────────┐ │
│  │ id: string (human-readable, e.g., "ecommerce")   │ │
│  │ title: string                                     │ │
│  │ description: string                               │ │
│  │ createdAt, updatedAt: timestamp                   │ │
│  └──────────────────────────────────────────────────┘ │
│                                                        │
│  ┌─────────────────┐  ┌─────────────────┐            │
│  │   Clusters      │  │     Models      │            │
│  │  (1 to many)    │  │   (1 to many)   │            │
│  │                 │  │                 │            │
│  │ - project_id    │  │ - project_id    │            │
│  └─────────────────┘  └─────────────────┘            │
│                                                        │
│  ┌─────────────────┐                                  │
│  │     Enums       │                                  │
│  │  (1 to many)    │                                  │
│  │                 │                                  │
│  │ - project_id    │                                  │
│  └─────────────────┘                                  │
└────────────────────────────────────────────────────────┘
```

## Domain Model

### Project Entity

```go
package project

import (
    "time"
    "modelcraft/pkg/bizerrors"
)

// Project 项目实体
type Project struct {
    ID          string    // 人类可读的项目标识 (e.g., "ecommerce", "crm")
    Title       string    // 项目显示标题
    Description string    // 项目描述
    Status      ProjectStatus  // 项目状态
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type ProjectStatus string

const (
    ProjectStatusActive   ProjectStatus = "active"
    ProjectStatusArchived ProjectStatus = "archived"
)

// Validate 验证项目实体
func (p *Project) Validate() error {
    if !isValidProjectID(p.ID) {
        return bizerrors.Errorf("invalid project ID format")
    }
    if p.Title == "" {
        return bizerrors.Errorf("project title cannot be empty")
    }
    return nil
}

// isValidProjectID 验证项目ID格式
// - 2-64字符
// - 小写字母、数字、连字符
// - 必须以字母开头
func isValidProjectID(id string) bool {
    if len(id) < 2 || len(id) > 64 {
        return false
    }
    // 正则: ^[a-z][a-z0-9-]*$
    // ... implementation
}
```

### Updated Model Locator

```go
package modeldesign

// ModelLocator 模型定位器 (扩展)
type ModelLocator struct {
    ProjectID    string  // 新增: 项目标识
    ModelName    string
    ClusterName  string
    DatabaseName string
}

// GetFullPath 获取完整路径
// 返回格式: project_id.cluster_name.database_name.model_name
func (l *ModelLocator) GetFullPath() string {
    return fmt.Sprintf("%s.%s.%s.%s",
        l.ProjectID, l.ClusterName, l.DatabaseName, l.ModelName)
}
```

### Updated Entities

**Cluster**:
```go
type DatabaseCluster struct {
    ID          string
    ProjectID   string  // 新增: 所属项目
    Name        string
    // ... other fields
}
```

**Model**:
```go
type DataModel struct {
    ModelMeta
    Fields []*FieldDefinition
}

type ModelMeta struct {
    ID           string
    ProjectID    string  // 新增: 所属项目
    ModelLocator        // 包含 ProjectID
    // ... other fields
}
```

**Enum**:
```go
type EnumDefinition struct {
    ID          string
    ProjectID   string  // 新增: 所属项目
    Name        string
    // ... other fields
}
```

## Database Schema

### New Table: projects

```sql
CREATE TABLE IF NOT EXISTS `projects` (
  -- Primary key
  `id` VARCHAR(64) NOT NULL COMMENT '项目唯一标识符（人类可读）',

  -- Basic info
  `title` VARCHAR(255) NOT NULL COMMENT '项目显示标题',
  `description` TEXT NULL COMMENT '项目描述信息',

  -- Status
  `status` VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '项目状态: active/archived',

  -- Metadata
  `settings` JSON NULL COMMENT '项目配置（预留）',

  -- Timestamps
  `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

  -- Constraints
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目定义表';
```

### Updated Tables

**database_clusters**:
```sql
ALTER TABLE `database_clusters`
  ADD COLUMN `project_id` VARCHAR(64) NOT NULL DEFAULT 'default' COMMENT '所属项目ID' AFTER `id`,
  ADD CONSTRAINT `fk_cluster_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE,
  DROP INDEX `idx_cluster_id`,
  ADD UNIQUE KEY `idx_cluster_name` (`project_id`, `name`) COMMENT '项目内集群名称唯一';
```

**models**:
```sql
ALTER TABLE `models`
  ADD COLUMN `project_id` VARCHAR(64) NOT NULL DEFAULT 'default' COMMENT '所属项目ID' AFTER `id`,
  ADD CONSTRAINT `fk_model_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE,
  DROP INDEX `idx_models_name`,
  ADD UNIQUE KEY `idx_models_name` (`project_id`, `cluster_name`, `database_name`, `name`) COMMENT '项目内模型名称唯一';
```

**model_enums**:
```sql
ALTER TABLE `model_enums`
  ADD COLUMN `project_id` VARCHAR(64) NOT NULL DEFAULT 'default' COMMENT '所属项目ID' AFTER `id`,
  ADD CONSTRAINT `fk_enum_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE,
  DROP INDEX `idx_model_enums_name`,
  ADD UNIQUE KEY `idx_enum_name` (`project_id`, `name`) COMMENT '项目内枚举名称唯一';
```

## Repository Layer

### ProjectRepository Interface

```go
package project

import "context"

type ProjectRepository interface {
    // Create 创建项目
    Create(ctx context.Context, project *Project) error

    // GetByID 根据ID获取项目
    GetByID(ctx context.Context, id string) (*Project, error)

    // List 列出项目
    List(ctx context.Context, status ...ProjectStatus) ([]*Project, error)

    // Update 更新项目
    Update(ctx context.Context, project *Project) error

    // Delete 删除项目（硬删除，需慎重）
    Delete(ctx context.Context, id string) error

    // ExistsByID 检查项目是否存在
    ExistsByID(ctx context.Context, id string) (bool, error)
}
```

### sqlc Implementation

```go
package repository

import (
    "context"
    "time"
    "modelcraft/internal/domain/project"
    "modelcraft/pkg/bizerrors"
    "sqlc"
)

type ProjectModel struct {
    ID          string     `db:"type:varchar(64);primaryKey"`
    Title       string     `db:"type:varchar(255);not null"`
    Description string     `db:"type:text"`
    Status      string     `db:"type:varchar(20);default:'active';index:idx_status"`
    Settings    string     `db:"type:json"`
    CreatedAt   time.Time  `db:"autoCreateTime"`
    UpdatedAt   time.Time  `db:"autoUpdateTime"`
}

func (ProjectModel) TableName() string {
    return "projects"
}

type GormProjectRepository struct {
    db *sql.DB
}

func NewGormProjectRepository(db *sql.DB) project.ProjectRepository {
    return &GormProjectRepository{db: db}
}

// Implementation methods...
```

## Application Layer

### ProjectAppService

```go
package app

import (
    "context"
    "time"
    "modelcraft/internal/domain/project"
    "modelcraft/pkg/bizerrors"
)

type ProjectAppService struct {
    projectRepo project.ProjectRepository
}

func NewProjectAppService(projectRepo project.ProjectRepository) *ProjectAppService {
    return &ProjectAppService{
        projectRepo: projectRepo,
    }
}

// CreateProject 创建项目
func (s *ProjectAppService) CreateProject(ctx context.Context, id, title, description string) (*project.Project, error) {
    // 1. 验证ID格式
    if !isValidProjectID(id) {
        return nil, bizerrors.Errorf("invalid project ID format")
    }

    // 2. 检查ID是否已存在
    exists, err := s.projectRepo.ExistsByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, bizerrors.Errorf("project ID already exists: %s", id)
    }

    // 3. 创建项目实体
    now := time.Now()
    proj := &project.Project{
        ID:          id,
        Title:       title,
        Description: description,
        Status:      project.ProjectStatusActive,
        CreatedAt:   now,
        UpdatedAt:   now,
    }

    // 4. 验证并保存
    if err := proj.Validate(); err != nil {
        return nil, err
    }

    if err := s.projectRepo.Create(ctx, proj); err != nil {
        return nil, err
    }

    return proj, nil
}

// GetProject 获取项目
func (s *ProjectAppService) GetProject(ctx context.Context, id string) (*project.Project, error) {
    return s.projectRepo.GetByID(ctx, id)
}

// ListProjects 列出项目
func (s *ProjectAppService) ListProjects(ctx context.Context) ([]*project.Project, error) {
    return s.projectRepo.List(ctx, project.ProjectStatusActive)
}

// UpdateProject 更新项目
func (s *ProjectAppService) UpdateProject(ctx context.Context, id string, title, description *string) (*project.Project, error) {
    // 1. 获取现有项目
    proj, err := s.projectRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 2. 更新字段
    if title != nil {
        proj.Title = *title
    }
    if description != nil {
        proj.Description = *description
    }
    proj.UpdatedAt = time.Now()

    // 3. 验证并保存
    if err := proj.Validate(); err != nil {
        return nil, err
    }

    if err := s.projectRepo.Update(ctx, proj); err != nil {
        return nil, err
    }

    return proj, nil
}

// DeleteProject 删除项目
func (s *ProjectAppService) DeleteProject(ctx context.Context, id string) error {
    // 注意: 这将级联删除所有关联的 clusters, models, enums
    // 建议在删除前检查是否有关联资源
    return s.projectRepo.Delete(ctx, id)
}
```

## API Layer

### GraphQL Schema

```graphql
# New types
type Project {
  id: ID!
  title: String!
  description: String
  status: ProjectStatus!
  createdAt: DateTime!
  updatedAt: DateTime!
}

enum ProjectStatus {
  ACTIVE
  ARCHIVED
}

input CreateProjectInput {
  id: ID!           # Human-readable, e.g., "ecommerce"
  title: String!
  description: String
}

input UpdateProjectInput {
  id: ID!
  title: String
  description: String
}

# Mutations
extend type Mutation {
  createProject(input: CreateProjectInput!): Project!
  updateProject(input: UpdateProjectInput!): Project!
  deleteProject(id: ID!): Boolean!
}

# Queries
extend type Query {
  project(id: ID!): Project
  projects(status: ProjectStatus): [Project!]!
}

# Update existing types to include projectId
extend type Model {
  projectId: ID!
}

extend type DatabaseCluster {
  projectId: ID!
}

extend type Enum {
  projectId: ID!
}
```

### REST Endpoints

```
# Project management
GET    /api/v1/projects              # List all projects
GET    /api/v1/projects/:id          # Get project by ID
POST   /api/v1/projects              # Create project
PUT    /api/v1/projects/:id          # Update project
DELETE /api/v1/projects/:id          # Delete project

# Project-scoped resources (new pattern)
GET    /api/v1/projects/:projectId/clusters
GET    /api/v1/projects/:projectId/models
GET    /api/v1/projects/:projectId/enums

# Backward compatible (defaults to "default" project)
GET    /api/v1/clusters              # Equivalent to /api/v1/projects/default/clusters
GET    /api/v1/models                # Equivalent to /api/v1/projects/default/models
```

## Migration Strategy

### Step 1: Schema Migration

```sql
-- 1. Create projects table
CREATE TABLE `projects` (...);

-- 2. Insert default project
INSERT INTO `projects` (id, title, description, status, created_at, updated_at)
VALUES ('default', 'Default Project', 'System default project for backward compatibility', 'active', NOW(), NOW());

-- 3. Add project_id columns
ALTER TABLE `database_clusters` ADD COLUMN `project_id` VARCHAR(64) NOT NULL DEFAULT 'default' AFTER `id`;
ALTER TABLE `models` ADD COLUMN `project_id` VARCHAR(64) NOT NULL DEFAULT 'default' AFTER `id`;
ALTER TABLE `model_enums` ADD COLUMN `project_id` VARCHAR(64) NOT NULL DEFAULT 'default' AFTER `id`;

-- 4. Add foreign keys
ALTER TABLE `database_clusters` ADD CONSTRAINT `fk_cluster_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE;
ALTER TABLE `models` ADD CONSTRAINT `fk_model_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE;
ALTER TABLE `model_enums` ADD CONSTRAINT `fk_enum_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE;

-- 5. Update unique indexes
ALTER TABLE `database_clusters` DROP INDEX `idx_cluster_id`, ADD UNIQUE KEY `idx_cluster_name` (`project_id`, `name`);
ALTER TABLE `models` DROP INDEX `idx_models_name`, ADD UNIQUE KEY `idx_models_name` (`project_id`, `cluster_name`, `database_name`, `name`);
ALTER TABLE `model_enums` DROP INDEX `idx_model_enums_name`, ADD UNIQUE KEY `idx_enum_name` (`project_id`, `name`);
```

### Step 2: Application Bootstrap

```go
// In server startup
func EnsureDefaultProject(ctx context.Context, projectRepo project.ProjectRepository) error {
    exists, err := projectRepo.ExistsByID(ctx, "default")
    if err != nil {
        return err
    }

    if !exists {
        defaultProject := &project.Project{
            ID:          "default",
            Title:       "Default Project",
            Description: "System default project for backward compatibility",
            Status:      project.ProjectStatusActive,
            CreatedAt:   time.Now(),
            UpdatedAt:   time.Now(),
        }
        return projectRepo.Create(ctx, defaultProject)
    }

    return nil
}
```

## Backward Compatibility

### API Behavior

1. **Omitted project_id**: Defaults to "default" project
2. **Existing endpoints**: Continue to work without changes
3. **New endpoints**: Provide explicit project context

### Example

```graphql
# Old way (still works, uses "default" project)
query {
  models {
    id
    name
  }
}

# New way (explicit project)
query {
  models(projectId: "ecommerce") {
    id
    name
    projectId
  }
}
```

## Error Handling

### New Error Codes

```go
// Project not found
bizerrors.Errorf("project not found: %s", projectID)

// Project ID format invalid
bizerrors.Errorf("invalid project ID format: must be 2-64 lowercase alphanumeric characters with hyphens, starting with a letter")

// Project ID already exists
bizerrors.Errorf("project ID already exists: %s", projectID)

// Cannot delete project with resources
bizerrors.Errorf("cannot delete project with existing resources: %d clusters, %d models, %d enums", ...)
```

## Testing Strategy

### Unit Tests
- Project entity validation
- Project ID format validation
- Repository CRUD operations

### Integration Tests
- Project creation and management
- Resource isolation within projects
- Migration script execution
- Backward compatibility with "default" project

### Test Cases
1. Create project with valid ID
2. Create project with invalid ID (should fail)
3. Create duplicate project (should fail)
4. Create cluster in project A, verify not visible in project B
5. Create model without project_id, verify it goes to "default"
6. Delete project with resources (decide behavior)
7. Migrate existing data to "default" project

## Performance Considerations

### Indexing
- Add `project_id` to all relevant indexes
- Update query patterns to filter by project_id first

### Query Optimization
```sql
-- Before
SELECT * FROM models WHERE name = 'User';

-- After (more selective)
SELECT * FROM models WHERE project_id = 'ecommerce' AND name = 'User';
```

### Connection Pooling
- No impact on connection pooling strategy
- Projects are logical containers, not separate databases

## Security Considerations

### Future Access Control
- Projects provide natural boundary for permissions
- Can add `project_permissions` table later
- Row-level security based on project_id

### Validation
- Strict project ID format prevents injection
- Foreign key constraints ensure referential integrity
- Cascade deletes should be carefully controlled

## Open Issues

1. **Project Deletion**: Should we prevent deletion if resources exist, or cascade delete?
   - Recommendation: Prevent by default, add force flag for cascade

2. **Project Archival**: Should archived projects hide their resources?
   - Recommendation: Phase 2 feature, keep simple for now

3. **Cross-Project References**: Future need for model relations across projects?
   - Recommendation: Out of scope, document as limitation

## Alternatives Considered

### Alternative 1: Workspace Instead of Project
- Similar concept, different terminology
- Rejected: "Project" is more widely understood in developer tools

### Alternative 2: Namespace Prefix
- Use prefixed names like "ecommerce:User" instead of separate entity
- Rejected: Less flexible, harder to enforce constraints

### Alternative 3: Tag-Based Organization
- Use tags to group resources instead of hierarchical project
- Rejected: Lacks strong isolation and unique constraints
