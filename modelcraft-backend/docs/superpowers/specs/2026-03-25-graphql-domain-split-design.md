# GraphQL Domain Split Design

**Date:** 2026-03-25
**Status:** Approved

## Background

The design-time GraphQL API is publicly exposed. Consumers need to clearly understand which operations belong to the `org` domain and which belong to the `org/project` domain. Currently all operations are served from a single endpoint `/org/{orgName}/design/graphql` with no structural distinction.

## Goal

Split the design-time GraphQL API into two independent endpoints, each with its own schema and generated code, so that the URL itself communicates the operation's domain scope.

## Decision

**Two independent endpoints with separate gqlgen schemas (breaking change).**

| Endpoint | Domain | URL |
|----------|--------|-----|
| Org | Organization-level management | `POST /org/{orgName}/graphql` |
| Project | Project-level design operations | `POST /org/{orgName}/project/{projectSlug}/graphql` |

The old endpoint `/org/{orgName}/design/graphql` is removed with no redirect.

---

## Section 1: Schema File Structure

### Org Schema (`api/graph/org/schema/`)

| File | Contents |
|------|----------|
| `schema.graphql` | `schema { query: Query, mutation: Mutation }` |
| `base.graphql` | Shared scalars (Int64, Date, Time), Node interface, PageInfo |
| `project.graphql` | Project CRUD: `projects`, `project(slug)`, `createProject`, `updateProject`, `deleteProject`. Also includes Cluster types needed by `createProject` (clusterInput) |
| `permission.graphql` | Role & permission operations: `permissionRoles`, `createRole`, `deleteRole`, `createCustomRole`, `updatePermissionRole`, `addPermissionToRole`, `removePermissionFromRole`, `assignRoleToUser`, `userRoleAssignments` |
| `user_management.graphql` | User & org operations: `me`, `myOrganizations`, `organizationMembers`, `updateOrganization` |

### Project Schema (`api/graph/project/schema/`)

| File | Contents |
|------|----------|
| `schema.graphql` | `schema { query: Query, mutation: Mutation }` |
| `base.graphql` | Shared scalars, Node interface, PageInfo (independent copy) |
| `cluster.graphql` | Cluster & database operations: `databaseCluster`, `listDatabases`, `listTables`, `updateProjectCluster`, `testDatabaseConnection` |
| `model.graphql` | All model operations |
| `field.graphql` | All field operations |
| `enum.graphql` | All enum operations |
| `logical_foreign_key.graphql` | All logical FK operations |

**Note:** `base.graphql` exists independently in both schemas. gqlgen compiles each schema in isolation so there is no cross-schema type sharing.

---

## Section 2: gqlgen Configuration

Replace the single `gqlgen.yml` with two config files.

### `gqlgen.org.yml`

```yaml
schema:
  - api/graph/org/schema/*.graphql

exec:
  package: generated
  filename: internal/interfaces/graphql/org/generated/generated.go

model:
  filename: internal/interfaces/graphql/org/generated/model_gen.go
  package: generated

resolver:
  package: orggraphql
  layout: follow-schema
  dir: internal/interfaces/graphql/org
  filename_template: "{name}.resolvers.go"
```

### `gqlgen.project.yml`

```yaml
schema:
  - api/graph/project/schema/*.graphql

exec:
  package: generated
  filename: internal/interfaces/graphql/project/generated/generated.go

model:
  filename: internal/interfaces/graphql/project/generated/model_gen.go
  package: generated

resolver:
  package: projectgraphql
  layout: follow-schema
  dir: internal/interfaces/graphql/project
  filename_template: "{name}.resolvers.go"
```

**Note:** Both configs use `package: generated` (not `orggenerated` or `projectgenerated`). Isolation is achieved via separate directories (`org/generated/` vs `project/generated/`), not package names. Import paths will differ: `modelcraft/internal/interfaces/graphql/org/generated` vs `modelcraft/internal/interfaces/graphql/project/generated`.

### Code Generation Commands

```bash
gqlgen generate --config gqlgen.org.yml
gqlgen generate --config gqlgen.project.yml
```

Update Makefile/justfile to run both commands.

---

## Section 3: Resolver Structure

### `internal/interfaces/graphql/org/resolver.go`

```go
package orggraphql

type Resolver struct {
    ProjectAppService      *project.ProjectAppService
    ClusterAppService      *cluster.DatabaseClusterAppService
    OrganizationAppService *appOrg.OrganizationAppService
    UserRepo               user.UserRepository
    RoleAppService         *appRole.RoleAppService
    RoleService            *appPermission.RoleService
    PermissionService      *appPermission.PermissionService
    UserRoleService        *appPermission.UserRoleService
}
```

### `internal/interfaces/graphql/project/resolver.go`

```go
package projectgraphql

type Resolver struct {
    ModelDesignService       *modeldesign.ModelDesignAppService
    ReverseEngineerService   *modeldesign.ReverseEngineerAppService
    RepairModelUseCase       *modeldesign.RepairModelUseCase
    ActualSchemaQueryUseCase *modeldesign.ActualSchemaQueryUseCase
    GroupAppService          *modeldesign.ModelGroupAppService
    LogicalFKAppService      *modeldesign.LogicalFKAppService
    EnumAppService           *modeldesign.EnumAppService
    UserRoleService          *appPermission.UserRoleService
}
```

### Resolver File Migration

| Existing file | Destination |
|---------------|-------------|
| `user_management.resolvers.go` | `org/` |
| `permission.resolvers.go` | `org/` |
| `project.resolvers.go` | `org/` |
| `model.resolvers.go` | `project/` |
| `field.resolvers.go` | `project/` |
| `enum.resolvers.go` | `project/` |
| `logical_foreign_key.resolvers.go` | `project/` |
| `base.resolvers.go` | Both `org/` and `project/` (ping/hello) |

### Adapter File Migration

Adapters are **duplicated, not shared**. Each side gets its own copy in its own directory.

| Adapter Type | Org Adapters | Project Adapters |
|--------------|--------------|------------------|
| Error adapters | `project_error_adapter.go` | `model_error_adapter.go`, `field_error_adapter.go`, `enum_error_adapter.go`, `cluster_error_adapter.go`, `fk_error_adapter.go` |
| Mappers | `project_mapper.go`, `user_management_mapper.go` | `model_mapper.go`, `field_mapper.go`, `enum_mapper.go`, `cluster_mapper.go`, `group_mapper.go` |

**Rationale for duplication:** Each endpoint is independently deployable and maintainable. Sharing adapters would create a coupling that defeats the purpose of splitting. By duplicating, each side is fully self-contained, and changes to one side don't risk breaking the other.

---

## Section 4: Routing and Handlers

### Handler Functions

**`internal/interfaces/graphql/org/handler.go`**
```go
// OrgGraphQLHandler creates a GraphQL handler for the org domain
func OrgGraphQLHandler(resolver *Resolver) http.HandlerFunc

// OrgPlaygroundHandler serves GraphQL Playground for the org domain
func OrgPlaygroundHandler() http.HandlerFunc
```

**`internal/interfaces/graphql/project/handler.go`**
```go
// ProjectGraphQLHandler creates a GraphQL handler for the project domain
func ProjectGraphQLHandler(resolver *Resolver) http.HandlerFunc

// ProjectPlaygroundHandler serves GraphQL Playground for the project domain
func ProjectPlaygroundHandler() http.HandlerFunc
```

Each handler registers the `@hasPermission` directive independently. The directive is instantiated via `NewHasPermissionDirective(resolver.UserRoleService)` in each handler's `GraphQLHandler` function, so each endpoint gets its own directive instance bound to its own resolver's `UserRoleService`.

### Route Registration

`SetupDesignGraphQLRoutesOnChi` is replaced by two functions:

```go
// SetupOrgGraphQLRoutesOnChi registers /org/{orgName}/graphql
func SetupOrgGraphQLRoutesOnChi(router chi.Router, handlers *DesignHandlers, cfg *config.Config)

// SetupProjectGraphQLRoutesOnChi registers /org/{orgName}/project/{projectSlug}/graphql
func SetupProjectGraphQLRoutesOnChi(router chi.Router, handlers *DesignHandlers, cfg *config.Config)
```

Each function:
1. Creates its own Resolver instance from services in `handlers *DesignHandlers`
2. Calls the appropriate handler function to construct the HTTP handler
3. Registers the handler on the router

Route middleware chains:

```
Org endpoint (/org/{orgName}/graphql):
  ChiJWTAuthMiddleware → ChiGraphQLOrgMiddleware → OrgGraphQLHandler

Project endpoint (/org/{orgName}/project/{projectSlug}/graphql):
  ChiJWTAuthMiddleware → ChiGraphQLOrgMiddleware → ChiGraphQLProjectMiddleware → ProjectGraphQLHandler
```

### `main.go` Call Site

```go
designHandlers, _ := http.CreateDesignHandlers(repoFactory, cfg)
http.SetupOrgGraphQLRoutesOnChi(router, designHandlers, cfg)
http.SetupProjectGraphQLRoutesOnChi(router, designHandlers, cfg)
```

### Old Route Removal

`/org/{orgName}/design/graphql` is deleted with no redirect. This is a breaking change.

---

## Section 5: projectSlug Removal from Project Schema (Breaking Change)

The project endpoint URL already carries `{projectSlug}`. All `projectSlug` parameters are removed from the project schema and resolver arguments. Resolvers retrieve `projectSlug` via `ctxutils.GetProjectSlugFromContext(ctx)`.

### New Context Utility Function

Add to `pkg/ctxutils/userctx.go`:

```go
const (
    ContextKeyProjectSlug contextKey = "project_slug"
)

// SetProjectSlug stores the project slug in context.
func SetProjectSlug(ctx context.Context, projectSlug string) context.Context {
    return context.WithValue(ctx, ContextKeyProjectSlug, projectSlug)
}

// GetProjectSlugFromContext extracts project slug from context.
// Returns error if not found or empty.
func GetProjectSlugFromContext(ctx context.Context) (string, error) {
    val := ctx.Value(ContextKeyProjectSlug)
    if val == nil {
        return "", fmt.Errorf("project slug not found in context")
    }
    projectSlug, ok := val.(string)
    if !ok || projectSlug == "" {
        return "", fmt.Errorf("project slug not found in context")
    }
    return projectSlug, nil
}
```

### New Middleware

`ChiGraphQLProjectMiddleware` — extracts `{projectSlug}` from the Chi URL parameter and writes it into the request context using `ctxutils.SetProjectSlug()`. Add to `internal/middleware/chi_middleware.go` alongside `ChiGraphQLOrgMiddleware`.

```go
// ChiGraphQLProjectMiddleware extracts projectSlug from URL and sets it in context.
func ChiGraphQLProjectMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            projectSlug := chi.URLParam(r, "projectSlug")
            if projectSlug == "" {
                http.Error(w, "projectSlug not found in URL", http.StatusBadRequest)
                return
            }
            ctx := ctxutils.SetProjectSlug(r.Context(), projectSlug)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Affected Operations

All operations in the project schema that previously accepted `projectSlug` as a top-level argument or inside an input type:

**Queries:** `databaseCluster`, `listDatabases` (input), `listTables` (input), `model`, `models` (input), `modelByName`, `modelJsonSchema`, `modelGroups`, `fields`, `enum`, `enums`, `enumReferences`, `logicalForeignKeys`

**Mutations:** All model mutations, all field mutations, all enum mutations, all FK mutations, `updateProjectCluster`, `testDatabaseConnection` (input)

For each affected resolver, replace `args.ProjectSlug` (or `args.Input.ProjectSlug`) with:

```go
projectSlug, err := ctxutils.GetProjectSlugFromContext(ctx)
if err != nil {
    // Handle error - return appropriate error response
}
// Use projectSlug from context
```

---

## Section 6: `@hasPermission` Directive

Both endpoints register the directive independently. Each handler instantiates the directive using its own resolver's `UserRoleService`.

**In `internal/interfaces/graphql/org/handler.go`:**
```go
func OrgGraphQLHandler(resolver *Resolver) http.HandlerFunc {
    config := generated.Config{
        Resolvers: resolver,
    }
    config.Directives.HasPermission = NewHasPermissionDirective(resolver.UserRoleService).HasPermission
    srv := handler.NewDefaultServer(generated.NewExecutableSchema(config))
    // ... rest of handler
}
```

**In `internal/interfaces/graphql/project/handler.go`:**
```go
func ProjectGraphQLHandler(resolver *Resolver) http.HandlerFunc {
    config := generated.Config{
        Resolvers: resolver,
    }
    config.Directives.HasPermission = NewHasPermissionDirective(resolver.UserRoleService).HasPermission
    srv := handler.NewDefaultServer(generated.NewExecutableSchema(config))
    // ... rest of handler
}
```

This pattern mirrors the current `GraphQLHandler` implementation in `internal/interfaces/graphql/handler.go`.

---

## Section 7: `DesignHandlers` and Construction

`DesignHandlers` struct in `internal/interfaces/http/routes.go` (unchanged):

```go
type DesignHandlers struct {
    ModelAppService           *modeldesign.ModelDesignAppService
    ClusterAppService         *cluster.DatabaseClusterAppService
    ReverseEngineerAppService *modeldesign.ReverseEngineerAppService
    EnumAppService            *modeldesign.EnumAppService
    ProjectAppService         *project.ProjectAppService
    OrgAppService             *appOrg.OrganizationAppService
    RoleAppService            *appRole.RoleAppService
    GroupAppService           *modeldesign.ModelGroupAppService
    LogicalFKAppService       *modeldesign.LogicalFKAppService
    PermRoleService           *appPermission.RoleService
    PermPermissionService     *appPermission.PermissionService
    PermUserRoleService       *appPermission.UserRoleService
    ModelRepository           domainModelDesign.ModelRepository
    UserRepo                  domainUser.UserRepository
    ClusterManager            *repository.ClusterConnectionManager
    RepairModelUseCase        *modeldesign.RepairModelUseCase
    ActualSchemaQueryUseCase  *modeldesign.ActualSchemaQueryUseCase
}
```

`CreateDesignHandlers` remains the single construction entry point for all app services.

Both setup functions accept `*DesignHandlers` and construct their respective Resolver instances:

```go
func SetupOrgGraphQLRoutesOnChi(router chi.Router, handlers *DesignHandlers, cfg *config.Config) {
    // ... JWT setup ...
    
    orgResolver := &orggraphql.Resolver{
        ProjectAppService:      handlers.ProjectAppService,
        ClusterAppService:      handlers.ClusterAppService,
        OrganizationAppService: handlers.OrgAppService,
        UserRepo:               handlers.UserRepo,
        RoleAppService:         handlers.RoleAppService,
        RoleService:            handlers.PermRoleService,
        PermissionService:      handlers.PermPermissionService,
        UserRoleService:        handlers.PermUserRoleService,
    }
    
    // Register routes using orgResolver
    router.Route("/org/{orgName}/graphql", func(r chi.Router) {
        // middleware...
        r.Post("/", orggraphql.OrgGraphQLHandler(orgResolver))
        r.Get("/",  orggraphql.OrgPlaygroundHandler())
    })
}

func SetupProjectGraphQLRoutesOnChi(router chi.Router, handlers *DesignHandlers, cfg *config.Config) {
    // ... JWT setup ...
    
    projectResolver := &projectgraphql.Resolver{
        ModelDesignService:       handlers.ModelAppService,
        ReverseEngineerService:   handlers.ReverseEngineerAppService,
        RepairModelUseCase:       handlers.RepairModelUseCase,
        ActualSchemaQueryUseCase: handlers.ActualSchemaQueryUseCase,
        GroupAppService:          handlers.GroupAppService,
        LogicalFKAppService:      handlers.LogicalFKAppService,
        EnumAppService:           handlers.EnumAppService,
        UserRoleService:          handlers.PermUserRoleService,
    }
    
    // Register routes using projectResolver
    router.Route("/org/{orgName}/project/{projectSlug}/graphql", func(r chi.Router) {
        // middleware...
        r.Post("/", projectgraphql.ProjectGraphQLHandler(projectResolver))
        r.Get("/",  projectgraphql.ProjectPlaygroundHandler())
    })
}
```

**Dependency isolation:** Each Resolver holds only the services it uses. While both Resolvers are constructed from the same `DesignHandlers`, each is semantically isolated — the org endpoint can't accidentally call project services, and vice versa.

---

## Section 8: base.graphql and base.resolvers.go Maintenance

Both `api/graph/org/schema/base.graphql` and `api/graph/project/schema/base.graphql` define identical shared types: scalars, Node interface, PageInfo.

Similarly, both `org/base.resolvers.go` and `project/base.resolvers.go` will have identical implementations (ping, hello queries).

**Maintenance strategy:**

1. **Keep them synchronized manually** — When updating scalar definitions, Node interface, or base resolvers, update both copies
2. **Add CI sync check** — Create a pre-commit or CI script that verifies both `base.graphql` files have identical content. Add a similar check for `base.resolvers.go` implementations to prevent drift
3. **Alternative (future):** If base files diverge significantly, extract truly shared schema into a utility schema and use gqlgen includes (if adopting a different schema loader)

For now, synchronization is the responsibility of developers making schema/resolver changes. Add to `.pre-commit-hooks.yaml` or CI:

```bash
# Check base.graphql files are identical
diff api/graph/org/schema/base.graphql api/graph/project/schema/base.graphql || exit 1

# Check base.resolvers.go files are structurally similar
# (Optional: cmp for byte-exact match, or custom validation)
```

---

## Section 9: Migration and Deprecation

This is a **breaking change.** The old endpoint `/org/{orgName}/design/graphql` is removed immediately with no transition period.

**For API consumers:**
- Update all requests from `/org/{orgName}/design/graphql` to `/org/{orgName}/graphql` (org operations) or `/org/{orgName}/project/{projectSlug}/graphql` (project operations)
- The schema is split but no fields are removed — only reorganized into separate endpoints
- Both endpoints use the same authentication scheme (JWT)

**Release notes should:**
- Clearly document the new endpoint URLs
- Provide a mapping table of old operations to new endpoints
- Emphasize the breaking change

---

## Summary of Changes

| Area | Change |
|------|--------|
| Schema files | Split into `api/graph/org/schema/` and `api/graph/project/schema/` |
| gqlgen config | `gqlgen.yml` → `gqlgen.org.yml` + `gqlgen.project.yml` |
| Generated code | `org/generated/` + `project/generated/`, isolated by directory |
| Resolvers | `orggraphql.Resolver` + `projectgraphql.Resolver`, each in separate packages |
| Adapters | Duplicated — each side has its own copy in `org/adapter/` and `project/adapter/` |
| Handlers | `org/handler.go` + `project/handler.go`, independent directive registration |
| Routes | `/org/{orgName}/graphql` + `/org/{orgName}/project/{projectSlug}/graphql` |
| Setup functions | `SetupOrgGraphQLRoutesOnChi` + `SetupProjectGraphQLRoutesOnChi` |
| Old route | `/org/{orgName}/design/graphql` deleted, no redirect |
| projectSlug in schema | Removed from all project-domain operations; taken from context |
| Context utilities | `ctxutils.GetProjectSlugFromContext()`, `ctxutils.SetProjectSlug()` in `pkg/ctxutils/userctx.go` |
| New middleware | `ChiGraphQLProjectMiddleware` in `internal/middleware/chi_middleware.go` |
| `DesignHandlers` | Unchanged; both setup functions extract needed services from it |
| base.graphql | Maintained independently in both directories; CI sync check required |
| base.resolvers.go | Maintained independently in both directories; sync required |
| Code generation | `gqlgen generate --config gqlgen.org.yml` + `gqlgen generate --config gqlgen.project.yml` |
