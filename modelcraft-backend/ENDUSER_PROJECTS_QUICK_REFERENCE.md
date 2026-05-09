# EndUserProjects Resolver - Quick Reference

## Critical File Paths

| Purpose | Path | Key Lines |
|---------|------|-----------|
| **Schema** | `api/graph/org/schema/end_user.graphql` | Add query after line 120 |
| **Resolver** | `internal/interfaces/graphql/org/end_user.resolvers.go` | Add after line 335 |
| **Directive** | `internal/interfaces/graphql/org/directives.go` | Lines 62-71 (allowEndUser logic) |
| **Repository Interface** | `internal/domain/enduser/end_user_repository.go` | Lines 37-42 |
| **Repo Implementation** | `internal/infrastructure/repository/sql_enduser_repository.go` | Lines 316-363 |
| **Context Utils** | `pkg/ctxutils/userctx.go` | Lines 158-162 (IsEndUser check) |
| **Database Schema** | `db/schema/mysql/12_end_user_auth.sql` | Tables: end_user_role_users, end_user_roles |

## Key Functions to Know

### 1. Check if End-User Caller
```go
if ctxutils.IsEndUser(ctx) {
  // This is an end-user caller
}
```

### 2. Extract Context Values
```go
orgName, err := ctxutils.GetOrgNameFromContext(ctx)
userID, err := ctxutils.GetUserIDFromContext(ctx)
```

### 3. Query Accessible Projects
```go
accessibleProjects, err := r.EndUserMgmtAppService.EndUserRepo.
  ListAccessibleProjectsByRoleAssignment(ctx, orgName, userID)
```

## Data Flow Summary

```
EndUser Request → @hasPermission (allowEndUser: true) → Directive Grants Access
   ↓
Extract org_name & user_id from context
   ↓
Query repository: ListAccessibleProjectsByRoleAssignment()
   ↓
SQL: end_user_role_users JOIN end_user_roles LEFT JOIN projects
   ↓
Return AccessibleProject[] → Convert to GraphQL Project[]
   ↓
Return to client
```

## Authorization Check (Directive Logic)

**File**: `directives.go:62-71`

```go
if ctxutils.IsEndUser(ctx) {
  if !allowEndUser {
    // Deny access
    return nil, newPermissionDeniedError(...)
  }
  // Allow access (bypass RBAC)
  return next(ctx)
}
```

**Key Insight**: When `@hasPermission(..., allowEndUser: true)`, end-users bypass RBAC and go straight to resolver.

## Schema Definition

Add to `end_user.graphql`:
```graphql
extend type Query {
  endUserProjects: [Project!]! @hasPermission(action: "end-user:read", allowEndUser: true)
}
```

## Resolver Implementation Pattern

```go
func (r *queryResolver) EndUserProjects(ctx context.Context) ([]*generated.Project, error) {
  // 1. Extract context
  orgName, _ := ctxutils.GetOrgNameFromContext(ctx)
  userID, _ := ctxutils.GetUserIDFromContext(ctx)
  
  // 2. Query repository
  accessible, _ := r.EndUserMgmtAppService.EndUserRepo.
    ListAccessibleProjectsByRoleAssignment(ctx, orgName, userID)
  
  // 3. Convert to GraphQL
  result := make([]*generated.Project, len(accessible))
  for i, ap := range accessible {
    result[i] = &generated.Project{
      ID: ap.ProjectSlug,
      Slug: ap.ProjectSlug,
      Title: ap.ProjectTitle,
      OrgName: orgName,
      Status: generated.ProjectStatusActive,
    }
  }
  
  return result, nil
}
```

## Domain Types

### AccessibleProject (domain)
```go
type AccessibleProject struct {
  ProjectSlug  string
  ProjectTitle string
}
```

### Project (GraphQL)
```go
type Project {
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

## Database Query

```sql
SELECT DISTINCT ur.role_id, r.project_slug, COALESCE(p.title, r.project_slug)
FROM end_user_role_users ur
JOIN end_user_roles r ON r.id = ur.role_id AND r.org_name = ur.org_name
LEFT JOIN projects p ON p.org_name = r.org_name AND p.slug = r.project_slug AND p.deleted_at = 0
WHERE ur.org_name = ? AND ur.user_id = ? AND r.deleted_at = 0
ORDER BY r.project_slug ASC
```

## Testing Checklist

- [ ] Schema compiles after adding query
- [ ] `just generate-gql` completes without errors
- [ ] Resolver extracts orgName and userID correctly
- [ ] Repository returns empty list when no roles assigned
- [ ] Repository returns projects when roles exist
- [ ] GraphQL conversion works correctly
- [ ] Directive permits end-user access (allowEndUser: true)
- [ ] No permission checks needed in resolver (directive handles it)

## Common Gotchas

1. **Don't extract end-user ID in directive** — directive already validated; use context values in resolver
2. **Use allowEndUser: true** — enables end-user access (bypasses RBAC)
3. **Return empty slice, not nil** — GraphQL expects [Project!]!, not null
4. **Convert AccessibleProject correctly** — it only has ProjectSlug and ProjectTitle; other fields may need defaults
5. **Check soft-delete** — SQL filters by `r.deleted_at = 0` automatically

