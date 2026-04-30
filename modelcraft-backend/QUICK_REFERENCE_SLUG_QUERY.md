# Quick Reference: Adding `endUserPermissionBundleBySlug` Query

## File Checklist

### 1. GraphQL Schema
- **File:** `api/graph/project/schema/rbac.graphql`
- **Action:** Add query definition to `extend type Query`
- **Line:** After existing `endUserPermissionBundle` query
- **Pattern:** Same signature but with `slug` instead of `id`

```graphql
extend type Query {
  endUserPermissionBundleBySlug(slug: String!): EndUserPermissionBundle @hasPermission(action: "rbac:read")
}
```

### 2. Resolver Implementation
- **File:** `internal/interfaces/graphql/project/rbac.resolvers.go`
- **Action:** Add new resolver method
- **Pattern:** Extract context → Call app service → Convert DTO
- **Copy From:** `EndUserPermissionBundle()` method (lines 653-665)

```go
func (r *queryResolver) EndUserPermissionBundleBySlug(ctx context.Context, slug string) (*generated.EndUserPermissionBundle, error) {
  orgName, projectSlug, err := getOrgAndProjectFromContext(ctx)
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

### 3. Application Service
- **File:** `internal/app/rbac/bundle_app.go`
- **Action:** Add `GetBundleBySlug()` method
- **Pattern:** Copy logic from `GetBundleByID()` (lines 129-191), replace ID lookup with slug lookup
- **Key:** Call `rbacRepo.GetBundleBySlug()`, then load items and snapshots

### 4. Domain Repository Interface
- **File:** `internal/domain/rbac/repository.go`
- **Action:** Add method signature
- **Location:** After `GetBundleByID` method definition
- **Pattern:** Same scope pattern with slug parameter

```go
// GetBundleBySlug 根据 slug 获取权限包（org + project scoped）
GetBundleBySlug(ctx context.Context, orgName, projectSlug, slug string) (*EndUserPermissionBundle, error)
```

### 5. SQL Repository Implementation
- **File:** `internal/infrastructure/repository/sql_end_user_permission_repository.go`
- **Action:** Add implementation method
- **Pattern:** Copy from `GetBundleByID()` (lines 329-346), change to call `GetEndUserBundleBySlug`

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

### 6. sqlc Query Definition
- **File:** `db/queries/rbac/bundle.sql`
- **Action:** Add SQL query
- **Location:** After existing bundle queries
- **Pattern:** Lookup by slug instead of ID, maintain org+project scope

```sql
-- name: GetEndUserBundleBySlug :one
SELECT *
FROM end_user_permission_bundles
WHERE slug = ?
  AND org_name = ?
  AND project_slug = ?;
```

---

## Database Schema (Reference Only - No Changes Needed)

The `end_user_permission_bundles` table already has:
- **Column:** `slug VARCHAR(64) NOT NULL`
- **Index:** `UNIQUE KEY uq_bundles_org_project_slug (org_name, project_slug, slug)`

No schema changes required!

---

## Implementation Order

1. ✅ Add SQL query to `bundle.sql`
2. ✅ Run `just db generate` to generate sqlc types
3. ✅ Add domain repository interface method
4. ✅ Add SQL repository implementation
5. ✅ Add application service method
6. ✅ Add GraphQL resolver method
7. ✅ Update GraphQL schema
8. ✅ Run `just generate` to regenerate GraphQL code
9. ✅ Test with GraphQL query

---

## Testing the Query

```graphql
query {
  endUserPermissionBundleBySlug(slug: "my-bundle-slug") {
    id
    slug
    name
    description
    dataPermissionItems {
      id
      modelId
      grantType
      preset
    }
    createdAt
    updatedAt
  }
}
```

---

## Key Design Principles to Remember

1. **Org + Project Scope:** Always include both in database queries
2. **Error Handling:** Return `shared.NotFoundError` from repo → convert to `bizerrors` in app
3. **DTO Conversion:** Domain → GraphQL via `adapter.ToEndUserPermissionBundleDTO()`
4. **Consistency:** Match the pattern of existing `GetBundleByID` implementation
5. **sqlc Generation:** Run `just db generate` after SQL changes

---

## Files Summary

| Layer | File | Change |
|-------|------|--------|
| SQL | `db/queries/rbac/bundle.sql` | Add `GetEndUserBundleBySlug` query |
| Domain | `internal/domain/rbac/repository.go` | Add interface method |
| Infrastructure | `internal/infrastructure/repository/sql_end_user_permission_repository.go` | Add implementation |
| Application | `internal/app/rbac/bundle_app.go` | Add service method |
| Interface | `internal/interfaces/graphql/project/rbac.resolvers.go` | Add resolver |
| Schema | `api/graph/project/schema/rbac.graphql` | Add query definition |

