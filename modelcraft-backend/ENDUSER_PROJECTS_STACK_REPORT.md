# EndUserProjects Resolver - Full Stack Architecture Report

This document maps the complete data flow for implementing the `endUserProjects` resolver in the ModelCraft backend.

---

## 1. GraphQL Schema Layer

### File Paths
- **Org Schema Definitions**: `/data/home/lukemxjia/modelcraft/modelcraft-backend/api/graph/org/schema/`
  - `end_user.graphql` — EndUser types and queries
  - `project.graphql` — Project types and error unions
  - `schema.graphql` — Root Query/Mutation (extends)

### Current Schema Structure
**Location**: `/data/home/lukemxjia/modelcraft/modelcraft-backend/api/graph/org/schema/end_user.graphql`

- **Query**: `findUsers(where: UserWhereInput, after: String, first: Int): UserFindManyResult! @hasPermission(action: "end-user:read", allowEndUser: true)`
  - Already supports EndUser callers via `allowEndUser: true`
  - Returns `UserFindManyResult` (cursor-based pagination)

- **Types Available**:
  - `EndUser` — Org-scoped end-user account (id, username, isForbidden, isBuiltin, createdBy, createdAt, updatedAt)
  - `EndUserPublic` — Public projection for display (id, username, isBuiltin, createdAt)
  - `UserFindManyResult` — Pagination result (items: [EndUserPublic!]!, nextCursor, hasMore, reqId)

**Location**: `/data/home/lukemxjia/modelcraft/modelcraft-backend/api/graph/org/schema/project.graphql`

- **Project Type** (lines 167-176):
  ```graphql
  type Project implements Node {
    id: ID!
    slug: String!
    title: String!
    description: String!
    status: ProjectStatus!
    orgName: String!
    createdAt: String!
    updatedAt: String!
  }
  ```

- **Error Interface & Unions** (lines 8-30):
  - Error types: `ResourceNotFound`, `InvalidInput`
  - Project-specific: `ProjectAlreadyExists`, `CannotDeleteDefaultProject`, `DatabaseConnectionFailed`

### Query to Implement

Add to `end_user.graphql` (extending Query):
```graphql
extend type Query {
  # endUserProjects: list projects accessible to the current end-user
  # Accessible only to end-user callers (allowEndUser: true).
  # Uses end_user_role_users JOIN end_user_roles to find accessible projects.
  endUserProjects: [Project!]! @hasPermission(action: "end-user:read", allowEndUser: true)
}
```

---

## 2. GraphQL Resolver Layer

### File Structure
All resolver implementations live in `/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/interfaces/graphql/org/`

| Component | File | Purpose |
|-----------|------|---------|
| **Query Resolvers** | `project.resolvers.go` | Implements `projects` query (lines 378-414) |
| **EndUser Resolvers** | `end_user.resolvers.go` | Implements `findUsers` query (lines 283-335) |
| **Directive** | `directives.go` | `@hasPermission` directive with `allowEndUser` logic (lines 31-84) |
| **Resolver Struct** | `resolver.go` | DI container for all services |
| **Handler** | `handler.go` | GraphQL handler setup, directive registration |

### Key Resolver Patterns

#### Pattern 1: How `Projects` Query Works
**File**: `/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/interfaces/graphql/org/project.resolvers.go:378-414`

```go
// Projects is the resolver for the projects field.
func (r *queryResolver) Projects(ctx context.Context, input *generated.ListProjectsInput) ([]*generated.Project, error) {
  var result []*generated.Project
  graphqlErr := bizerrors.WithGraphqlErrorHandler(ctx, func() error {
    // Extract org from context
    orgName, err := ctxutils.GetOrgNameFromContext(ctx)
    if err != nil {
      return bizerrors.NewError(bizerrors.ParamInvalid, "organization context required")
    }

    // List projects via app service
    cmd := appProject.ListProjectsCommand{OrgName: orgName, Status: statusFilter}
    projects, err := r.ProjectAppService.ListProjectsByOrg(ctx, cmd)
    if err != nil {
      return err
    }

    // Convert domain to GraphQL
    result = adapter.ProjectMapper.ConvertProjectsToGraphQL(projects)
    return nil
  })

  if graphqlErr != nil {
    return nil, graphqlErr
  }
  return result, nil
}
```

**Key Points**:
1. Extract `orgName` from context via `ctxutils.GetOrgNameFromContext(ctx)`
2. Call app service: `r.ProjectAppService.ListProjectsByOrg(ctx, cmd)`
3. Convert domain objects to GraphQL using mapper: `adapter.ProjectMapper.ConvertProjectsToGraphQL(projects)`
4. Return `[]*generated.Project` slice (not paginated)

#### Pattern 2: How `FindUsers` Query Works (EndUser with allowEndUser: true)
**File**: `/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/interfaces/graphql/org/end_user.resolvers.go:283-335`

```go
// FindUsers is the resolver for the findUsers field.
func (r *queryResolver) FindUsers(ctx context.Context, where *generated.UserWhereInput, after *string, first *int32) (*generated.UserFindManyResult, error) {
  reqID := ctxutils.GetRequestID(ctx)

  // 1. Extract org from context (works for both tenant and end-user callers)
  orgName, err := ctxutils.GetOrgNameFromContext(ctx)
  if err != nil || orgName == "" {
    return nil, newGQLError("organization context required", "MISSING_ORGANIZATION")
  }

  // 2. Build command
  cmd := appEnduser.MetaUserFindManyCommand{OrgName: orgName}
  if after != nil {
    cmd.After = *after
  }
  if first != nil {
    cmd.First = int(*first)
  }
  if where != nil {
    cmd.Where = convertUserWhereInput(where)
  }

  // 3. Call app service
  result, err := r.MetaUserAppService.FindMany(ctx, cmd)
  if err != nil {
    var bizErr *bizerrors.BusinessError
    if errors.As(err, &bizErr) {
      return nil, newGQLError(bizErr.Msg(), bizErr.Info().GetCode())
    }
    return nil, err
  }

  // 4. Convert to GraphQL
  items := make([]*generated.EndUserPublic, 0, len(result.Items))
  for _, dto := range result.Items {
    items = append(items, &generated.EndUserPublic{
      ID:        dto.ID,
      Username:  dto.Username,
      IsBuiltin: dto.IsBuiltin,
      CreatedAt: dto.CreatedAt,
    })
  }

  return &generated.UserFindManyResult{
    Items:      items,
    NextCursor: &result.NextCursor,
    HasMore:    result.HasMore,
    ReqID:      reqID,
  }, nil
}
```

**Key Points**:
1. **Context extraction works for both tenant and end-user**: same `ctxutils.GetOrgNameFromContext(ctx)` call
2. **EndUser detection happens in @hasPermission directive** — see section 3 below
3. No explicit end-user ID extraction needed in resolver; directive already validated access
4. Return paginated result with cursor support

### Resolver Struct & Dependency Injection
**File**: `/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/interfaces/graphql/org/resolver.go`

```go
type Resolver struct {
  // Project CRUD
  ProjectAppService    *project.ProjectAppService
  ClusterAppService    *cluster.DatabaseClusterAppService
  AuthSchemaAppService *rls.AuthSchemaAppService

  // EndUser management
  EndUserMgmtAppService *appEnduser.EndUserManagementAppService
  MetaUserAppService    *appEnduser.MetaUserAppService

  // Permissions
  UserRoleService   *permission.UserRoleService
  PermissionService *permission.PermissionService
}
```

**For endUserProjects, you'll need**:
- `ProjectAppService` — to fetch projects
- `EndUserMgmtAppService` — to access enduser repository for `ListAccessibleProjectsByRoleAssignment`

---

## 3. Authorization Layer - @hasPermission Directive

### File Path
`/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/interfaces/graphql/org/directives.go`

### How @hasPermission Works with allowEndUser: true

**Directive Signature** (lines 31-84):
```go
func (d *HasPermissionDirective) HasPermission(
  ctx context.Context,
  obj interface{},
  next graphql.Resolver,
  action string,
  allowEndUser bool,
) (interface{}, error)
```

**Authorization Logic Flow** (lines 62-71):
```go
// End-user callers: default-deny. Only operations explicitly marked allowEndUser=true are accessible.
if ctxutils.IsEndUser(ctx) {
  if !allowEndUser {
    logger.Infof(ctx, "@hasPermission directive: end-user access denied (allowEndUser=false)")
    return nil, newPermissionDeniedError(action, userID, orgName, "end-user-default-deny")
  }
  logger.Infof(ctx, "@hasPermission directive: end-user access granted (allowEndUser=true)")
  return next(ctx)
}
```

**Key Insight**:
- When `allowEndUser: true`, EndUser callers **bypass all RBAC checks** and go directly to `next(ctx)`
- The directive doesn't validate permissions for end-users; it only checks `allowEndUser` flag
- For `endUserProjects`, you don't need extra permission checks in the resolver; the directive handles it

### Context Utilities - How to Detect End-User
**File**: `/data/home/lukemxjia/modelcraft/modelcraft-backend/pkg/ctxutils/userctx.go:158-162`

```go
// IsEndUser returns true if the request is from an EndUser caller.
func IsEndUser(ctx context.Context) bool {
  val, _ := ctx.Value(ContextKeyUserType).(string)
  return val == UserTypeEndUser  // "end_user"
}
```

**Other Context Methods Available**:
```go
// For all callers (tenant or end-user)
orgName, err := ctxutils.GetOrgNameFromContext(ctx)
userID, err := ctxutils.GetUserIDFromContext(ctx)

// EndUser-specific
endUserAdminID := ctxutils.GetEndUserAdminID(ctx)  // set by gateway for tenant admin callers
isEndUser := ctxutils.IsEndUser(ctx)

// Request tracing
reqID := ctxutils.GetRequestID(ctx)
```

---

## 4. Domain Repository Layer

### File Path
`/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/domain/enduser/end_user_repository.go`

### Interface Definition
```go
type EndUserRepository interface {
  // ... other methods ...

  // ListAccessibleProjectsByRoleAssignment: query end_user_role_users + end_user_roles
  // to find projects the end-user can access in the org.
  ListAccessibleProjectsByRoleAssignment(
    ctx context.Context,
    orgName, endUserID string,
  ) ([]AccessibleProject, error)

  // HasProjectAccessByRole: check if end-user has any role in the project
  HasProjectAccessByRole(
    ctx context.Context,
    orgName, endUserID, projectSlug string,
  ) (bool, error)
}

// AccessibleProject represents one project an end-user can access
type AccessibleProject struct {
  ProjectSlug  string
  ProjectTitle string
}
```

### Implementation
**File**: `/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go:316-363`

```go
func (r *SqlEndUserRepository) ListAccessibleProjectsByRoleAssignment(
  ctx context.Context,
  orgName, endUserID string,
) ([]enduser.AccessibleProject, error) {
  const query = `
    SELECT DISTINCT ur.role_id, r.project_slug, COALESCE(p.title, r.project_slug) AS project_title
    FROM end_user_role_users ur
    JOIN end_user_roles r
      ON r.id = ur.role_id
     AND r.org_name = ur.org_name
    LEFT JOIN projects p
      ON p.org_name = r.org_name
     AND p.slug = r.project_slug
     AND p.deleted_at = 0
    WHERE ur.org_name = ?
      AND ur.user_id = ?
      AND r.deleted_at = 0
    ORDER BY r.project_slug ASC
  `

  rows, err := r.db.QueryContext(ctx, query, orgName, endUserID)
  if err != nil {
    return nil, sqlerr.WrapSQLError(err)
  }
  defer rows.Close()

  seen := make(map[string]struct{})
  projects := make([]enduser.AccessibleProject, 0)
  for rows.Next() {
    var roleID, projectSlug, projectTitle string
    if scanErr := rows.Scan(&roleID, &projectSlug, &projectTitle); scanErr != nil {
      return nil, sqlerr.WrapSQLError(scanErr)
    }
    if _, ok := seen[projectSlug]; !ok {  // deduplicate
      seen[projectSlug] = struct{}{}
      projects = append(projects, enduser.AccessibleProject{
        ProjectSlug:  projectSlug,
        ProjectTitle: projectTitle,
      })
    }
  }
  if err = rows.Err(); err != nil {
    return nil, sqlerr.WrapSQLError(err)
  }
  return projects, nil
}
```

**Key Points**:
- Uses DISTINCT to deduplicate by project_slug
- LEFT JOIN projects to get project title (fallback to slug if not found)
- Filters by `r.deleted_at = 0` (soft-delete check)
- Returns sorted by `project_slug ASC`

---

## 5. Database Schema Layer

### End-User Authorization Tables
**File**: `/data/home/lukemxjia/modelcraft/modelcraft-backend/db/schema/mysql/12_end_user_auth.sql`

#### Key Tables:

1. **end_user_users** (Org-scoped)
   - Columns: id, org_name, username, password, is_forbidden, is_builtin, created_by, created_at, updated_at, deleted_at
   - Unique key: (org_name, username, delete_token)

2. **end_user_roles** (Project-scoped)
   - Columns: id, org_name, project_slug, name, is_implicit, created_at, updated_at, deleted_at
   - Unique key: (org_name, project_slug, name, delete_token)

3. **end_user_role_users** (Role-User mapping, Org-scoped)
   - Columns: id, org_name, role_id, user_id, created_at
   - Unique key: (org_name, role_id, user_id)
   - Foreign keys:
     - `fk_eu_role_users_role` → end_user_roles(org_name, id) ON DELETE CASCADE
     - `fk_eu_role_users_user` → end_user_users(org_name, id) ON DELETE CASCADE

#### Query Path for endUserProjects:
```
end_user_role_users (org_name, user_id) 
  ↓ JOIN end_user_roles (org_name, id)
    ↓ LEFT JOIN projects (org_name, slug)
      ↓ SELECT project_slug, project_title
```

---

## 6. Application Service Layer

### File Path
`/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/app/enduser/`

### Example Pattern - How to Access Repository
The `EndUserMgmtAppService` provides the entry point to access the repository:

```go
type EndUserManagementAppService struct {
  // ... services ...
  EndUserRepo enduser.EndUserRepository
}

// Example: how the app service would use it
func (s *EndUserManagementAppService) ListAccessibleProjects(
  ctx context.Context,
  orgName, endUserID string,
) ([]enduser.AccessibleProject, error) {
  return s.EndUserRepo.ListAccessibleProjectsByRoleAssignment(ctx, orgName, endUserID)
}
```

---

## 7. Type Adapters

### File Path
`/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/interfaces/graphql/org/adapter/`

### Project Mapper
**File**: `project_mapper.go`

```go
type projectMapperImpl struct{}

var ProjectMapper = projectMapperImpl{}

// ConvertProjectToGraphQL converts domain project to GraphQL
func (m projectMapperImpl) ConvertProjectToGraphQL(proj *project.Project) *generated.Project {
  // ... conversion logic
}

// ConvertProjectsToGraphQL converts slice of domain projects to GraphQL
func (m projectMapperImpl) ConvertProjectsToGraphQL(projects []*project.Project) []*generated.Project {
  result := make([]*generated.Project, 0, len(projects))
  for _, p := range projects {
    result = append(result, m.ConvertProjectToGraphQL(p))
  }
  return result
}
```

---

## 8. Implementation Recipe for endUserProjects Resolver

### Step 1: Update GraphQL Schema
**File**: `/data/home/lukemxjia/modelcraft/modelcraft-backend/api/graph/org/schema/end_user.graphql`

Add query (after `findUsers` at line 120):
```graphql
extend type Query {
  # List projects accessible to the current end-user via role assignments
  endUserProjects: [Project!]! @hasPermission(action: "end-user:read", allowEndUser: true)
}
```

### Step 2: Regenerate GraphQL Code
```bash
just generate-gql
```

This will update generated code in `/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/interfaces/graphql/org/generated/`

### Step 3: Implement Resolver
**File**: `/data/home/lukemxjia/modelcraft/modelcraft-backend/internal/interfaces/graphql/org/end_user.resolvers.go`

Add after `FindUsers` resolver (around line 336):
```go
// EndUserProjects is the resolver for the endUserProjects field.
func (r *queryResolver) EndUserProjects(ctx context.Context) ([]*generated.Project, error) {
  // 1. Extract org and end-user ID from context
  orgName, err := ctxutils.GetOrgNameFromContext(ctx)
  if err != nil || orgName == "" {
    return nil, newGQLError("organization context required", "MISSING_ORGANIZATION")
  }

  endUserID, err := ctxutils.GetUserIDFromContext(ctx)
  if err != nil || endUserID == "" {
    return nil, newGQLError("end-user identity required", "UNAUTHENTICATED")
  }

  // 2. Query accessible projects via role assignment
  accessibleProjects, err := r.EndUserMgmtAppService.EndUserRepo.ListAccessibleProjectsByRoleAssignment(
    ctx,
    orgName,
    endUserID,
  )
  if err != nil {
    var bizErr *bizerrors.BusinessError
    if errors.As(err, &bizErr) {
      return nil, newGQLError(bizErr.Msg(), bizErr.Info().GetCode())
    }
    return nil, err
  }

  // 3. Convert to GraphQL Project type
  // Map AccessibleProject domain objects to GraphQL Project via project slug lookup
  result := make([]*generated.Project, 0, len(accessibleProjects))
  for _, ap := range accessibleProjects {
    proj := &generated.Project{
      ID:          ap.ProjectSlug,  // ID same as slug per schema
      Slug:        ap.ProjectSlug,
      Title:       ap.ProjectTitle,
      Description: "",              // Not available from accessible projects query
      OrgName:     orgName,
      CreatedAt:   "",              // Not available
      UpdatedAt:   "",              // Not available
      Status:      generated.ProjectStatusActive,  // Assume active
    }
    result = append(result, proj)
  }

  return result, nil
}
```

**Alternative: Use App Service** (if one exists that wraps repository):
```go
// Call through app service
projects, err := r.EndUserMgmtAppService.ListAccessibleProjectsByRoleAssignment(ctx, orgName, endUserID)
```

### Step 4: Run Tests
```bash
just test
```

---

## 9. Key Data Flow Diagram

```
HTTP Request (EndUser caller)
  ↓
Gateway (sets context: user_type="end_user", user_id, org_name)
  ↓
OrgGraphQLHandler initializes @hasPermission directive
  ↓
Query: endUserProjects
  ↓
@hasPermission directive checks:
  - ctxutils.IsEndUser(ctx) == true
  - allowEndUser == true
  - GRANTS ACCESS ✓
  ↓
queryResolver.EndUserProjects(ctx)
  ↓
Extract from context:
  - orgName ← ctxutils.GetOrgNameFromContext(ctx)
  - endUserID ← ctxutils.GetUserIDFromContext(ctx)
  ↓
Repository query:
  - ListAccessibleProjectsByRoleAssignment(orgName, endUserID)
  ↓
SQL JOIN:
  - end_user_role_users (org_name, user_id) 
  - JOIN end_user_roles (org_name, id)
  - LEFT JOIN projects (org_name, slug)
  ↓
Return AccessibleProject[] with (ProjectSlug, ProjectTitle)
  ↓
Convert to GraphQL Project[] type
  ↓
Return to client
```

---

## 10. File Reference Summary

| Layer | File | Key Function/Type |
|-------|------|-------------------|
| **GraphQL Schema** | `api/graph/org/schema/end_user.graphql` | Query definition |
| **GraphQL Schema** | `api/graph/org/schema/project.graphql` | Project type |
| **Resolver** | `internal/interfaces/graphql/org/end_user.resolvers.go` | `EndUserProjects()` resolver implementation |
| **Resolver Struct** | `internal/interfaces/graphql/org/resolver.go` | `Resolver` with `EndUserMgmtAppService` |
| **Directive** | `internal/interfaces/graphql/org/directives.go` | `HasPermission()` with `allowEndUser` logic |
| **Context Utils** | `pkg/ctxutils/userctx.go` | `IsEndUser()`, `GetOrgNameFromContext()`, `GetUserIDFromContext()` |
| **Domain Repository** | `internal/domain/enduser/end_user_repository.go` | `EndUserRepository.ListAccessibleProjectsByRoleAssignment()` interface |
| **Repo Implementation** | `internal/infrastructure/repository/sql_enduser_repository.go` | SQL implementation (lines 316-363) |
| **Database Schema** | `db/schema/mysql/12_end_user_auth.sql` | `end_user_users`, `end_user_roles`, `end_user_role_users` tables |
| **Domain Type** | `internal/domain/enduser/end_user_repository.go` | `AccessibleProject` struct (lines 5-9) |

---

## 11. Error Handling Strategy

Follow the pattern from `FindUsers` resolver:

```go
// Business errors from repository
if err != nil {
  var bizErr *bizerrors.BusinessError
  if errors.As(err, &bizErr) {
    return nil, newGQLError(bizErr.Msg(), bizErr.Info().GetCode())
  }
  return nil, err  // Infrastructure errors propagate
}
```

Common errors:
- `MISSING_ORGANIZATION` — org_name not in context (caught by directive)
- `UNAUTHENTICATED` — user_id not in context (caught by directive)
- Database/SQL errors — wrapped by `sqlerr.WrapSQLError()`

---

## 12. Testing Strategy

Mock/test the resolver with:
1. **EndUser context**: `ctx = ctxutils.SetUserType(ctx, ctxutils.UserTypeEndUser)`
2. **Mock repository**: Inject mock `EndUserMgmtAppService` with `EndUserRepo`
3. **Test cases**:
   - EndUser with role assignments → returns projects
   - EndUser with no role assignments → returns empty slice
   - EndUser forbidden status → still returns accessible projects (authorization check is role-based, not forbiddenstatus-based)
   - Missing org_name context → error

