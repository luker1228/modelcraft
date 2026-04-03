# GraphQL Domain Split Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split design-time GraphQL API into two independent endpoints: org-level (`/org/{orgName}/graphql`) and project-level (`/org/{orgName}/project/{projectSlug}/graphql`), each with its own schema, generated code, and resolvers.

**Architecture:** The implementation follows a clean separation-of-concerns pattern. Schema files are split by domain into `api/graph/org/schema/` and `api/graph/project/schema/`. Each domain gets its own gqlgen configuration file, generated code directory, resolver package, adapter directory, and HTTP handler. The `DesignHandlers` struct remains a single construction point; two setup functions extract domain-specific services and wire them into domain-specific resolvers. Project-level operations no longer require `projectSlug` parameters in the schema — instead, URL context is extracted via middleware and passed through `context.Context`.

**Tech Stack:** gqlgen (schema-first code generation), Chi (HTTP router), context.Context (request scoping), pkg/ctxutils (context utilities)

---

## Chunk 1: Foundation — Context Utilities & Middleware

### Task 1: Add projectSlug context utilities

**Files:**
- Modify: `pkg/ctxutils/userctx.go`

- [ ] **Step 1: Open `pkg/ctxutils/userctx.go` and review the existing pattern**

Observe:
- `ContextKeyOrgName` constant
- `SetOrgName()` and `GetOrgNameFromContext()` functions
- Error handling pattern (return empty string + error)

- [ ] **Step 2: Add projectSlug constant and functions**

Add after the `ContextKeyUseCache` constant definition (around line 24):

```go
// ContextKeyProjectSlug is the key for storing project slug in context
ContextKeyProjectSlug contextKey = "project_slug"
```

- [ ] **Step 3: Add SetProjectSlug function**

Add after `SetOrgName()` (around line 72):

```go
// SetProjectSlug stores the project slug in context.
func SetProjectSlug(ctx context.Context, projectSlug string) context.Context {
	return context.WithValue(ctx, ContextKeyProjectSlug, projectSlug)
}
```

- [ ] **Step 4: Add GetProjectSlugFromContext function**

Add after `GetOrgNameFromContext()` (around line 92):

```go
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

- [ ] **Step 5: Run tests to verify no regressions**

```bash
go test ./pkg/ctxutils -v
```

Expected: All tests pass.

- [ ] **Step 6: Commit**

```bash
git add pkg/ctxutils/userctx.go
git commit -m "feat: add projectSlug context utilities"
```

---

### Task 2: Add ChiGraphQLProjectMiddleware

**Files:**
- Modify: `internal/middleware/chi_middleware.go` (or appropriate middleware file)

- [ ] **Step 1: Find the file containing ChiGraphQLOrgMiddleware**

Look for `internal/middleware/` directory and locate the middleware that has `ChiGraphQLOrgMiddleware`.

- [ ] **Step 2: Review ChiGraphQLOrgMiddleware implementation**

Observe the pattern:
- Takes no parameters, returns `func(http.Handler) http.Handler`
- Uses `chi.URLParam()` to extract URL parameter
- Uses `ctxutils.SetOrgName()` to inject into context

- [ ] **Step 3: Add ChiGraphQLProjectMiddleware after ChiGraphQLOrgMiddleware**

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

- [ ] **Step 4: Run tests to verify no regressions**

```bash
go test ./internal/middleware -v
```

Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/middleware/chi_middleware.go
git commit -m "feat: add ChiGraphQLProjectMiddleware for projectSlug context injection"
```

---

## Chunk 2: Schema Files

### Task 3: Create org schema directory and files

**Files:**
- Create: `api/graph/org/schema/schema.graphql`
- Create: `api/graph/org/schema/base.graphql`
- Create: `api/graph/org/schema/project.graphql`
- Create: `api/graph/org/schema/permission.graphql`
- Create: `api/graph/org/schema/user_management.graphql`

- [ ] **Step 1: Create org schema directory**

```bash
mkdir -p api/graph/org/schema
```

- [ ] **Step 2: Create `api/graph/org/schema/schema.graphql`**

```graphql
schema {
  query: Query
  mutation: Mutation
}
```

- [ ] **Step 3: Copy base.graphql to org schema**

```bash
cp api/graph/schema/base.graphql api/graph/org/schema/base.graphql
```

- [ ] **Step 4: Create `api/graph/org/schema/project.graphql`**

Copy content from `api/graph/schema/project.graphql` (contains project CRUD and cluster types). **Important:** Keep all project CRUD operations but do NOT include field operations (those go to project domain).

- [ ] **Step 5: Create `api/graph/org/schema/permission.graphql`**

Copy content from `api/graph/schema/permission.graphql` (all role and permission operations).

- [ ] **Step 6: Create `api/graph/org/schema/user_management.graphql`**

Copy content from `api/graph/schema/user_management.graphql` (user, org, membership operations).

- [ ] **Step 7: Verify schema files are in place**

```bash
ls -la api/graph/org/schema/
```

Expected output:
```
total 40
base.graphql
permission.graphql
project.graphql
schema.graphql
user_management.graphql
```

- [ ] **Step 8: Commit**

```bash
git add api/graph/org/schema/
git commit -m "feat: create org domain GraphQL schema files"
```

---

### Task 4: Create project schema directory and files

**Files:**
- Create: `api/graph/project/schema/schema.graphql`
- Create: `api/graph/project/schema/base.graphql`
- Create: `api/graph/project/schema/cluster.graphql`
- Create: `api/graph/project/schema/model.graphql`
- Create: `api/graph/project/schema/field.graphql`
- Create: `api/graph/project/schema/enum.graphql`
- Create: `api/graph/project/schema/logical_foreign_key.graphql`

- [ ] **Step 1: Create project schema directory**

```bash
mkdir -p api/graph/project/schema
```

- [ ] **Step 2: Create `api/graph/project/schema/schema.graphql`**

```graphql
schema {
  query: Query
  mutation: Mutation
}
```

- [ ] **Step 3: Copy base.graphql to project schema**

```bash
cp api/graph/schema/base.graphql api/graph/project/schema/base.graphql
```

- [ ] **Step 4: Create project-domain schema files**

For the following files, copy from `api/graph/schema/` but **remove all `projectSlug` parameters** from queries and mutations:

**`api/graph/project/schema/cluster.graphql`**
- Copy from `api/graph/schema/project.graphql` but extract only cluster-related types and operations
- Remove: `createProject`, `updateProject`, `deleteProject`, `projects`, `project(slug)`
- Keep: `databaseCluster`, `listDatabases`, `listTables`, `updateProjectCluster`, `testDatabaseConnection`
- **Remove `projectSlug` parameter** from all queries/mutations — it will come from context

**`api/graph/project/schema/model.graphql`**
- Copy from `api/graph/schema/model.graphql`
- **Remove `projectSlug` parameter** from all operations

**`api/graph/project/schema/field.graphql`**
- Copy from `api/graph/schema/field.graphql`
- **Remove `projectSlug` parameter** from all operations

**`api/graph/project/schema/enum.graphql`**
- Copy from `api/graph/schema/enum.graphql`
- **Remove `projectSlug` parameter** from all operations

**`api/graph/project/schema/logical_foreign_key.graphql`**
- Copy from `api/graph/schema/logical_foreign_key.graphql`
- **Remove `projectSlug` parameter** from all operations

- [ ] **Step 5: Verify schema files are in place**

```bash
ls -la api/graph/project/schema/
```

Expected output:
```
total 56
base.graphql
cluster.graphql
enum.graphql
field.graphql
logical_foreign_key.graphql
model.graphql
schema.graphql
```

- [ ] **Step 6: Commit**

```bash
git add api/graph/project/schema/
git commit -m "feat: create project domain GraphQL schema files with projectSlug removed"
```

---

## Chunk 3: gqlgen Configuration

### Task 5: Create gqlgen configurations

**Files:**
- Create: `gqlgen.org.yml`
- Create: `gqlgen.project.yml`

- [ ] **Step 1: Create `gqlgen.org.yml`**

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

models:
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
      - github.com/99designs/gqlgen/graphql.Int32
  UUID:
    model:
      - github.com/99designs/gqlgen/graphql.UUID
  Int:
    model:
      - github.com/99designs/gqlgen/graphql.Int32
  Int64:
    model:
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
```

- [ ] **Step 2: Create `gqlgen.project.yml`**

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

models:
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
      - github.com/99designs/gqlgen/graphql.Int32
  UUID:
    model:
      - github.com/99designs/gqlgen/graphql.UUID
  Int:
    model:
      - github.com/99designs/gqlgen/graphql.Int32
  Int64:
    model:
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
```

- [ ] **Step 3: Verify files were created**

```bash
ls -la gqlgen*.yml
```

Expected:
```
gqlgen.org.yml
gqlgen.project.yml
```

- [ ] **Step 4: Commit**

```bash
git add gqlgen.org.yml gqlgen.project.yml
git commit -m "feat: create gqlgen configurations for org and project domains"
```

---

### Task 6: Generate code for org domain

**Files:**
- Create: `internal/interfaces/graphql/org/generated/generated.go`
- Create: `internal/interfaces/graphql/org/generated/model_gen.go`

- [ ] **Step 1: Create org generated directory**

```bash
mkdir -p internal/interfaces/graphql/org/generated
```

- [ ] **Step 2: Run gqlgen for org domain**

```bash
gqlgen generate --config gqlgen.org.yml
```

Expected: 
- `internal/interfaces/graphql/org/generated/generated.go` created (~1MB)
- `internal/interfaces/graphql/org/generated/model_gen.go` created

- [ ] **Step 3: Verify generation succeeded**

```bash
ls -la internal/interfaces/graphql/org/generated/
```

Expected:
```
generated.go
model_gen.go
```

- [ ] **Step 4: Commit**

```bash
git add internal/interfaces/graphql/org/generated/
git commit -m "chore: generate gqlgen code for org domain"
```

---

### Task 7: Generate code for project domain

**Files:**
- Create: `internal/interfaces/graphql/project/generated/generated.go`
- Create: `internal/interfaces/graphql/project/generated/model_gen.go`

- [ ] **Step 1: Create project generated directory**

```bash
mkdir -p internal/interfaces/graphql/project/generated
```

- [ ] **Step 2: Run gqlgen for project domain**

```bash
gqlgen generate --config gqlgen.project.yml
```

Expected:
- `internal/interfaces/graphql/project/generated/generated.go` created
- `internal/interfaces/graphql/project/generated/model_gen.go` created

- [ ] **Step 3: Verify generation succeeded**

```bash
ls -la internal/interfaces/graphql/project/generated/
```

Expected:
```
generated.go
model_gen.go
```

- [ ] **Step 4: Commit**

```bash
git add internal/interfaces/graphql/project/generated/
git commit -m "chore: generate gqlgen code for project domain"
```

---

## Chunk 4: Org Domain Implementation

### Task 8: Create org resolver and handler

**Files:**
- Create: `internal/interfaces/graphql/org/resolver.go`
- Create: `internal/interfaces/graphql/org/handler.go`
- Create: `internal/interfaces/graphql/org/permission.resolvers.go` (copy + adapt)
- Create: `internal/interfaces/graphql/org/user_management.resolvers.go` (copy + adapt)
- Create: `internal/interfaces/graphql/org/project.resolvers.go` (copy + adapt)
- Create: `internal/interfaces/graphql/org/base.resolvers.go` (copy)

- [ ] **Step 1: Create `internal/interfaces/graphql/org/resolver.go`**

```go
package orggraphql

import (
	"modelcraft/internal/app/cluster"
	"modelcraft/internal/app/organization"
	"modelcraft/internal/app/permission"
	"modelcraft/internal/app/project"
	"modelcraft/internal/app/role"
	"modelcraft/internal/domain/user"
)

// Resolver is the GraphQL resolver for org domain
type Resolver struct {
	// Project CRUD
	ProjectAppService      *project.ProjectAppService
	ClusterAppService      *cluster.DatabaseClusterAppService

	// Organization
	OrganizationAppService *organization.OrganizationAppService
	UserRepo               user.UserRepository

	// Permission (Casbin)
	RoleAppService    *role.RoleAppService
	RoleService       *permission.RoleService
	PermissionService *permission.PermissionService
	UserRoleService   *permission.UserRoleService
}
```

- [ ] **Step 2: Create `internal/interfaces/graphql/org/handler.go`**

```go
package orggraphql

import (
	"context"
	"modelcraft/internal/interfaces/graphql/org/generated"
	"modelcraft/pkg/ctxutils"
	"net/http"

	playgroundpkg "modelcraft/pkg/graphql"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
)

// injectRequestIDMiddleware adds requestId to GraphQL response extensions
func injectRequestIDMiddleware(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	resp := next(ctx)

	requestID := ctxutils.GetRequestID(ctx)
	if requestID == "" {
		return resp
	}

	if resp.Extensions == nil {
		resp.Extensions = make(map[string]any)
	}
	resp.Extensions["requestId"] = requestID

	return resp
}

// OrgGraphQLHandler creates GraphQL handler for org domain
func OrgGraphQLHandler(resolver *Resolver) http.HandlerFunc {
	// Create @hasPermission directive
	hasPermissionDirective := NewHasPermissionDirective(resolver.UserRoleService)

	// Create GraphQL config
	config := generated.Config{
		Resolvers: resolver,
	}

	// Configure directive
	config.Directives.HasPermission = hasPermissionDirective.HasPermission

	// Create GraphQL handler
	h := handler.NewDefaultServer(generated.NewExecutableSchema(config))

	// Add response middleware to inject requestId
	h.AroundResponses(injectRequestIDMiddleware)

	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

// OrgPlaygroundHandler serves GraphQL Playground for org domain
func OrgPlaygroundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "orgName")
		if orgName == "" {
			orgName = "default"
		}

		endpoint := "/org/" + orgName + "/graphql"

		ginHandler := playgroundpkg.Handler(playgroundpkg.PlaygroundConfig{
			Endpoint: endpoint,
			Title:    "GraphQL Playground - Org API (" + orgName + ")",
		})

		c, _ := gin.CreateTestContext(w)
		c.Request = r
		ginHandler(c)
	}
}
```

- [ ] **Step 3: Copy resolver files to org directory**

```bash
cp internal/interfaces/graphql/permission.resolvers.go internal/interfaces/graphql/org/
cp internal/interfaces/graphql/user_management.resolvers.go internal/interfaces/graphql/org/
cp internal/interfaces/graphql/project.resolvers.go internal/interfaces/graphql/org/
cp internal/interfaces/graphql/base.resolvers.go internal/interfaces/graphql/org/
```

- [ ] **Step 4: Create directives adapter for org domain**

Copy `internal/interfaces/graphql/directives.go` to `internal/interfaces/graphql/org/directives.go`

- [ ] **Step 5: Update resolver package names in copied files**

In each of the copied `.resolvers.go` files in `org/` directory:
- Change `package graphql` to `package orggraphql`

- [ ] **Step 6: Commit**

```bash
git add internal/interfaces/graphql/org/
git commit -m "feat: create org domain resolver and handler"
```

---

## Chunk 5: Project Domain Implementation

### Task 9: Create project resolver and handler

**Files:**
- Create: `internal/interfaces/graphql/project/resolver.go`
- Create: `internal/interfaces/graphql/project/handler.go`
- Create: `internal/interfaces/graphql/project/*.resolvers.go` (copy + adapt)

- [ ] **Step 1: Create `internal/interfaces/graphql/project/resolver.go`**

```go
package projectgraphql

import (
	"modelcraft/internal/app/modeldesign"
	"modelcraft/internal/app/permission"
)

// Resolver is the GraphQL resolver for project domain
type Resolver struct {
	// Model design
	ModelDesignService       *modeldesign.ModelDesignAppService
	ReverseEngineerService   *modeldesign.ReverseEngineerAppService
	RepairModelUseCase       *modeldesign.RepairModelUseCase
	ActualSchemaQueryUseCase *modeldesign.ActualSchemaQueryUseCase
	GroupAppService          *modeldesign.ModelGroupAppService
	LogicalFKAppService      *modeldesign.LogicalFKAppService

	// Enum
	EnumAppService *modeldesign.EnumAppService

	// Permission
	UserRoleService *permission.UserRoleService
}
```

- [ ] **Step 2: Create `internal/interfaces/graphql/project/handler.go`**

Similar to org/handler.go but:
- Import `modelcraft/internal/interfaces/graphql/project/generated`
- Use `projectgraphql` package
- Endpoint: `/org/{orgName}/project/{projectSlug}/graphql`
- Title: "GraphQL Playground - Project API"

```go
package projectgraphql

import (
	"context"
	"modelcraft/internal/interfaces/graphql/project/generated"
	"modelcraft/pkg/ctxutils"
	"net/http"

	playgroundpkg "modelcraft/pkg/graphql"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
)

// injectRequestIDMiddleware adds requestId to GraphQL response extensions
func injectRequestIDMiddleware(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	resp := next(ctx)

	requestID := ctxutils.GetRequestID(ctx)
	if requestID == "" {
		return resp
	}

	if resp.Extensions == nil {
		resp.Extensions = make(map[string]any)
	}
	resp.Extensions["requestId"] = requestID

	return resp
}

// ProjectGraphQLHandler creates GraphQL handler for project domain
func ProjectGraphQLHandler(resolver *Resolver) http.HandlerFunc {
	// Create @hasPermission directive
	hasPermissionDirective := NewHasPermissionDirective(resolver.UserRoleService)

	// Create GraphQL config
	config := generated.Config{
		Resolvers: resolver,
	}

	// Configure directive
	config.Directives.HasPermission = hasPermissionDirective.HasPermission

	// Create GraphQL handler
	h := handler.NewDefaultServer(generated.NewExecutableSchema(config))

	// Add response middleware to inject requestId
	h.AroundResponses(injectRequestIDMiddleware)

	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

// ProjectPlaygroundHandler serves GraphQL Playground for project domain
func ProjectPlaygroundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "orgName")
		if orgName == "" {
			orgName = "default"
		}
		projectSlug := chi.URLParam(r, "projectSlug")
		if projectSlug == "" {
			projectSlug = "default"
		}

		endpoint := "/org/" + orgName + "/project/" + projectSlug + "/graphql"

		ginHandler := playgroundpkg.Handler(playgroundpkg.PlaygroundConfig{
			Endpoint: endpoint,
			Title:    "GraphQL Playground - Project API (" + orgName + "/" + projectSlug + ")",
		})

		c, _ := gin.CreateTestContext(w)
		c.Request = r
		ginHandler(c)
	}
}
```

- [ ] **Step 3: Copy project domain resolver files**

```bash
cp internal/interfaces/graphql/model.resolvers.go internal/interfaces/graphql/project/
cp internal/interfaces/graphql/field.resolvers.go internal/interfaces/graphql/project/
cp internal/interfaces/graphql/enum.resolvers.go internal/interfaces/graphql/project/
cp internal/interfaces/graphql/logical_foreign_key.resolvers.go internal/interfaces/graphql/project/
cp internal/interfaces/graphql/base.resolvers.go internal/interfaces/graphql/project/
```

- [ ] **Step 4: Copy adapter files to project domain**

```bash
mkdir -p internal/interfaces/graphql/project/adapter
cp internal/interfaces/graphql/adapter/model_* internal/interfaces/graphql/project/adapter/
cp internal/interfaces/graphql/adapter/field_* internal/interfaces/graphql/project/adapter/
cp internal/interfaces/graphql/adapter/enum_* internal/interfaces/graphql/project/adapter/
cp internal/interfaces/graphql/adapter/fk_* internal/interfaces/graphql/project/adapter/
cp internal/interfaces/graphql/adapter/cluster_* internal/interfaces/graphql/project/adapter/
cp internal/interfaces/graphql/adapter/group_* internal/interfaces/graphql/project/adapter/
```

- [ ] **Step 5: Copy other support files to project domain**

```bash
cp internal/interfaces/graphql/directives.go internal/interfaces/graphql/project/
cp internal/interfaces/graphql/field_selection.go internal/interfaces/graphql/project/
cp internal/interfaces/graphql/fk_converters.go internal/interfaces/graphql/project/
cp internal/interfaces/graphql/interfaces.go internal/interfaces/graphql/project/
```

- [ ] **Step 6: Update package names in copied files**

In all `.go` files in `project/` directory:
- Change `package graphql` to `package projectgraphql`

- [ ] **Step 7: Commit**

```bash
git add internal/interfaces/graphql/project/
git commit -m "feat: create project domain resolver and handler"
```

---

## Chunk 6: Route Setup

### Task 10: Update routes.go to register both endpoints

**Files:**
- Modify: `internal/interfaces/http/routes.go`

- [ ] **Step 1: Add imports for new packages**

Add to imports:
```go
"modelcraft/internal/interfaces/graphql/org"
orggraphql "modelcraft/internal/interfaces/graphql/org"
projectgraphql "modelcraft/internal/interfaces/graphql/project"
```

- [ ] **Step 2: Replace SetupDesignGraphQLRoutesOnChi with SetupOrgGraphQLRoutesOnChi**

Replace the function body to register `/org/{orgName}/graphql` route:

```go
// SetupOrgGraphQLRoutesOnChi registers GraphQL endpoints for org domain
func SetupOrgGraphQLRoutesOnChi(router chi.Router, handlers *DesignHandlers, cfg *config.Config) {
	typeMapper := domainModelDesign.NewMySQLTypeMapper()
	schemaComparisonService := domainModelDesign.NewMySQLSchemaComparisonService(typeMapper)
	deployRepo := ddl.NewDeploymentService(handlers.ClusterManager)
	repairUseCase := modeldesign.NewRepairModelUseCase(
		handlers.ModelRepository,
		handlers.ClusterManager,
		deployRepo,
		schemaComparisonService,
	)

	actualSchemaService := ddl.NewActualSchemaService()
	actualSchemaQueryUseCase := modeldesign.NewActualSchemaQueryUseCase(actualSchemaService, handlers.ClusterManager)

	// Create org resolver
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

	publicKey := LoadRSAPublicKey(cfg)
	if publicKey == nil && !cfg.Auth.Design.SkipJWTValidation {
		logfacade.GetLogger(context.Background()).Fatal(
			"GraphQL authentication enabled but no RSA public key configured. " +
				"Please configure CASDOOR_CERTIFICATE, CASDOOR_JWT_PUBLIC_KEY_PATH, or CASDOOR_JWT_PUBLIC_KEY",
		)
	}

	jwtConfig := &middleware.JWTAuthConfig{
		PublicKey:           publicKey,
		ModelCraftSecret:    []byte(cfg.JWT.Secret),
		SkipValidation:      cfg.Auth.Design.SkipJWTValidation,
		AcceptCasdoorJWT:    cfg.Auth.Design.AcceptCasdoorJWT,
		AcceptModelcraftJWT: cfg.Auth.Design.AcceptModelcraftJWT,
	}

	// Register org endpoint
	router.Route("/org/{orgName}/graphql", func(r chi.Router) {
		r.Use(middleware.ChiJWTAuthMiddleware(jwtConfig))
		r.Use(middleware.ChiGraphQLOrgMiddleware())
		r.Post("/", orggraphql.OrgGraphQLHandler(orgResolver))
		r.Get("/",  orggraphql.OrgPlaygroundHandler())
	})
}
```

- [ ] **Step 3: Add new function SetupProjectGraphQLRoutesOnChi**

Add after SetupOrgGraphQLRoutesOnChi:

```go
// SetupProjectGraphQLRoutesOnChi registers GraphQL endpoints for project domain
func SetupProjectGraphQLRoutesOnChi(router chi.Router, handlers *DesignHandlers, cfg *config.Config) {
	typeMapper := domainModelDesign.NewMySQLTypeMapper()
	schemaComparisonService := domainModelDesign.NewMySQLSchemaComparisonService(typeMapper)
	deployRepo := ddl.NewDeploymentService(handlers.ClusterManager)
	repairUseCase := modeldesign.NewRepairModelUseCase(
		handlers.ModelRepository,
		handlers.ClusterManager,
		deployRepo,
		schemaComparisonService,
	)

	actualSchemaService := ddl.NewActualSchemaService()
	actualSchemaQueryUseCase := modeldesign.NewActualSchemaQueryUseCase(actualSchemaService, handlers.ClusterManager)

	// Create project resolver
	projectResolver := &projectgraphql.Resolver{
		ModelDesignService:       handlers.ModelAppService,
		ReverseEngineerService:   handlers.ReverseEngineerAppService,
		RepairModelUseCase:       repairUseCase,
		ActualSchemaQueryUseCase: actualSchemaQueryUseCase,
		GroupAppService:          handlers.GroupAppService,
		LogicalFKAppService:      handlers.LogicalFKAppService,
		EnumAppService:           handlers.EnumAppService,
		UserRoleService:          handlers.PermUserRoleService,
	}

	publicKey := LoadRSAPublicKey(cfg)
	if publicKey == nil && !cfg.Auth.Design.SkipJWTValidation {
		logfacade.GetLogger(context.Background()).Fatal(
			"GraphQL authentication enabled but no RSA public key configured. " +
				"Please configure CASDOOR_CERTIFICATE, CASDOOR_JWT_PUBLIC_KEY_PATH, or CASDOOR_JWT_PUBLIC_KEY",
		)
	}

	jwtConfig := &middleware.JWTAuthConfig{
		PublicKey:           publicKey,
		ModelCraftSecret:    []byte(cfg.JWT.Secret),
		SkipValidation:      cfg.Auth.Design.SkipJWTValidation,
		AcceptCasdoorJWT:    cfg.Auth.Design.AcceptCasdoorJWT,
		AcceptModelcraftJWT: cfg.Auth.Design.AcceptModelcraftJWT,
	}

	// Register project endpoint
	router.Route("/org/{orgName}/project/{projectSlug}/graphql", func(r chi.Router) {
		r.Use(middleware.ChiJWTAuthMiddleware(jwtConfig))
		r.Use(middleware.ChiGraphQLOrgMiddleware())
		r.Use(middleware.ChiGraphQLProjectMiddleware())
		r.Post("/", projectgraphql.ProjectGraphQLHandler(projectResolver))
		r.Get("/",  projectgraphql.ProjectPlaygroundHandler())
	})
}
```

- [ ] **Step 4: Verify the changes compile**

```bash
go build ./cmd/server
```

Expected: No compilation errors.

- [ ] **Step 5: Commit**

```bash
git add internal/interfaces/http/routes.go
git commit -m "feat: update routes to register org and project GraphQL endpoints"
```

---

### Task 11: Update main.go to call both setup functions

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Locate the call to SetupDesignGraphQLRoutesOnChi**

Find the main.go file and locate where `http.SetupDesignGraphQLRoutesOnChi()` is called.

- [ ] **Step 2: Replace with two calls**

Replace:
```go
http.SetupDesignGraphQLRoutesOnChi(router, designHandlers, cfg)
```

With:
```go
http.SetupOrgGraphQLRoutesOnChi(router, designHandlers, cfg)
http.SetupProjectGraphQLRoutesOnChi(router, designHandlers, cfg)
```

- [ ] **Step 3: Verify compilation**

```bash
go build ./cmd/server
```

Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat: call both org and project GraphQL setup functions"
```

---

## Chunk 7: Cleanup & Finalization

### Task 12: Remove old design GraphQL route and files

**Files:**
- Delete: Old `internal/interfaces/graphql/handler.go` (if truly duplicating functionality)
- Modify: `internal/interfaces/http/chi_setup.go` (remove `/design/graphql` route registration if it exists there)

- [ ] **Step 1: Check if old handler.go is still needed**

Review `internal/interfaces/graphql/handler.go` to see if it still has unique logic. If all logic has been moved to `org/` and `project/` handlers, delete it. If it has shared utility functions, those should be refactored to `orggraphql/` and `projectgraphql/` handlers or moved to a shared location.

- [ ] **Step 2: Verify no other routes reference the old endpoint**

```bash
grep -r "/design/graphql" --include="*.go" internal/
```

Expected: No matches (or only in comments/docs).

- [ ] **Step 3: If chi_setup.go still has SetupDesignGraphQLRoutesOnChi call, remove it**

If it exists, remove that line.

- [ ] **Step 4: Verify no compilation errors**

```bash
go build ./cmd/server
```

Expected: No errors.

- [ ] **Step 5: Commit cleanup**

```bash
git add -A
git commit -m "chore: remove old design GraphQL route references"
```

---

### Task 13: Run tests to verify no regressions

**Files:**
- Test: Run existing GraphQL tests

- [ ] **Step 1: Run all GraphQL interface tests**

```bash
go test ./internal/interfaces/graphql/... -v
```

Expected: All existing tests pass (or skip if they reference old handler).

- [ ] **Step 2: Run all HTTP interface tests**

```bash
go test ./internal/interfaces/http/... -v
```

Expected: All tests pass.

- [ ] **Step 3: Run all middleware tests**

```bash
go test ./internal/middleware/... -v
```

Expected: All tests pass.

- [ ] **Step 4: Run full test suite**

```bash
go test ./... -v
```

Expected: No regressions.

- [ ] **Step 5: Commit test results**

If any tests fail, fix them first before committing.

```bash
git add -A
git commit -m "test: verify no regressions after GraphQL domain split"
```

---

### Task 14: Build and smoke test the application

**Files:**
- Binary: `cmd/server/main.go`

- [ ] **Step 1: Build the application**

```bash
go build -o modelcraft ./cmd/server
```

Expected: Binary created successfully.

- [ ] **Step 2: Verify binary works (help/version if available)**

```bash
./modelcraft --help
```

Or start it and verify it doesn't crash on startup.

- [ ] **Step 3: Test org endpoint GraphQL Playground loads**

Start the server (if possible in your environment) and visit:
```
GET http://localhost:8080/org/test-org/graphql
```

Expected: GraphQL Playground loads with correct endpoint URL.

- [ ] **Step 4: Test project endpoint GraphQL Playground loads**

```
GET http://localhost:8080/org/test-org/project/test-project/graphql
```

Expected: GraphQL Playground loads with correct endpoint URL.

- [ ] **Step 5: Verify schema in Playground**

In Playground, check that:
- Org endpoint shows only org-domain queries (no model/field operations)
- Project endpoint shows only project-domain queries (no user/org operations, no projectSlug parameters)

- [ ] **Step 6: Final commit**

```bash
git add -A
git commit -m "chore: build and smoke test application"
```

---

## Testing Strategy

**Unit Tests:**
- Context utilities (projectSlug setters/getters)
- Middleware (ChiGraphQLProjectMiddleware extracts projectSlug correctly)

**Integration Tests:**
- Both resolvers can be instantiated
- Both routes register successfully on Chi
- JWT middleware functions on both endpoints
- Permission directives are registered correctly

**Manual Testing (if environment permits):**
- Org Playground loads at `/org/{orgName}/graphql`
- Project Playground loads at `/org/{orgName}/project/{projectSlug}/graphql`
- Org queries execute successfully
- Project queries execute successfully and retrieve projectSlug from context

---

## Rollback Plan

If implementation encounters critical issues:

1. Revert all commits: `git revert <commit-range>`
2. Restore old handler logic if deleted
3. Restore old route registration in chi_setup.go
4. Re-run tests to verify

---

## Known Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Schema duplication (base.graphql in two places) | Add CI check to verify both are identical |
| Adapter duplication increases maintenance | Well-defined split by domain; adapters moved together with resolvers |
| Missing projectSlug parameter removes might break existing resolvers | Each resolver file must be manually reviewed and projectSlug references removed from args extraction |
| Context extraction might fail if middleware not applied | Middleware is applied in route registration; test to verify |
| Compilation issues from package name changes | Run `go build ./cmd/server` after each major change |
