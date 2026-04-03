# Design: Adopt oapi-codegen for Auth/Org, Remove REST API for Business Domains

## Context

The current architecture maintains OpenAPI specifications for all domains (auth, org, projects, models, clusters, enums). This creates:

1. **Duplication**: Business APIs exist in both OpenAPI spec AND GraphQL schema
2. **Synchronization burden**: Changes must be made in two places
3. **Inconsistent patterns**: REST and GraphQL coexist for the same business logic
4. **Scope creep**: OpenAPI tool maintenance extends beyond its intended purpose (tenant management)

The project already has:
- GraphQL as the primary API for business logic (via gqlgen)
- OpenAPI for auth/org (tenant management) 
- Clear separation of concerns between:
  - **Tenant management** (Auth/Org) - REST-appropriate
  - **Business domains** (Projects/Models/Clusters/Enums) - GraphQL-appropriate

## Goals

- Single source of truth for business APIs: GraphQL schema only
- Simplified OpenAPI maintenance: only auth/org domains
- Clear architectural boundary: REST for identity, GraphQL for business
- Prevent GET proliferation on business APIs: mutations-only enforcement
- Reduce code generation scope and compile-time overhead
- Maintain backward compatibility where reasonable (migration path provided)

### Non-Goals

- Completely remove REST API layer (auth/org must remain REST)
- Support legacy REST clients indefinitely (provide 1-2 release notice)
- Maintain unused code paths or dead implementations

## Decisions

### Decision 1: Keep Auth and Org as REST-based
**What**: Auth and Organization endpoints remain in OpenAPI specs and are served via REST
**Why**: 
- Identity/auth is tenant-management concern, appropriate for REST
- Simple CRUD patterns work well with REST
- Existing clients (frontend, SDKs) depend on these
- Minimal synchronization burden (these change infrequently)

**Alternatives Considered**:
- Move to GraphQL entirely: Would require massive client migration, lower priority
- Keep minimal REST: Less clear architectural boundary

### Decision 2: Remove business domain YAML files
**What**: Delete `project.yaml`, `model.yaml`, `cluster.yaml`, `enum.yaml` from `api/openapi/`
**Why**:
- Business APIs are managed via GraphQL schema (`api/graph/schema/`)
- Eliminates duplication and synchronization overhead
- Prevents oapi-codegen from generating unused HTTP handlers
- Reduces compiled artifact size and startup time

**Alternatives Considered**:
- Keep specs for documentation: Use GraphQL schema documentation instead
- Archive to subdirectory: Delete approach is cleaner, removes temptation to resurrect
- Convert to GraphQL: Already done, just removing REST layer

### Decision 3: Update OpenAPI root to reference only auth/org/webhook
**What**: `openapi-root.yaml` will only include paths for `/api/auth/*`, `/api/org/*`, `/api/webhook/*`
**Why**:
- Reflects actual REST API surface
- Reduces OpenAPI bundle size
- Tooling (Swagger UI, code generators) focuses on correct scope
- Prevents confusion about API surface

**Alternatives Considered**:
- Keep all specs but mark as deprecated: Generates code that's not used, confusing
- Conditional code generation: Adds complexity to CI/CD

### Decision 4: Restrict GET operations on non-auth/org endpoints
**What**: Any remaining REST endpoints (if legacy or transitional) have no GET methods in OpenAPI
**Why**:
- Enforces mutation-only pattern for business APIs
- Prevents proliferation of read endpoints alongside GraphQL queries
- Single endpoint for reads (GraphQL only)
- Simplifies API surface and reduces cognitive load

**Alternatives Considered**:
- Allow GET for backward compatibility: Creates two sources of truth
- Remove endpoints entirely: This is the long-term goal; GET restriction is transitional

### Decision 5: Regenerate oapi-codegen code with reduced scope
**What**: Run `task generate-oapi` to produce `server.gen.go` with only auth/org/webhook handlers
**Why**:
- ServerInterface is updated to only include required methods
- Design handlers are no longer auto-generated, can be removed from delegation
- Compilation remains valid, type-safe
- Clear separation between tenant-management (REST) and business (GraphQL)

**Alternatives Considered**:
- Manual code editing: Undermines code generation value
- Keep generated code as-is: Leaves unused handlers in production code

## Technical Approach

### Phase 1: Update OpenAPI Specifications

1. **Update `api/openapi/openapi-root.yaml`**:
   - Remove all design domain path references (project, model, cluster, enum)
   - Keep auth, org, webhook path references
   - Update component schemas to only include referenced types

2. **Delete business domain YAML files**:
   - `api/openapi/project.yaml` → delete
   - `api/openapi/model.yaml` → delete
   - `api/openapi/cluster.yaml` → delete
   - `api/openapi/enum.yaml` → delete
   - Optional: Move to `api/openapi/archive/` for historical reference

3. **Verify webhook handling**:
   - Keep `webhook.yaml` (infrastructure for incoming webhooks)
   - May remove from generated interface if not used (evaluate separately)

### Phase 2: Update Code Generation

1. **Run OpenAPI bundling**:
   - `task bundle-oapi` → produces `api/openapi/openapi.yaml` with reduced scope

2. **Regenerate Go code**:
   - `task generate-oapi` → regenerates `internal/interfaces/http/generated/server.gen.go`
   - New ServerInterface has ~25 methods (auth/org/webhook) instead of ~60+

### Phase 3: Update Server Implementation

1. **Update `internal/interfaces/http/server.go`**:
   - Remove ProjectHandler, ModelHandler, ClusterHandler, EnumHandler fields
   - Remove corresponding constructor parameters
   - Remove all design endpoint methods (CreateProject, GetModel, etc.)
   - Keep auth, org, webhook handler delegation methods

2. **Update route registration in `internal/interfaces/http/chi_setup.go`**:
   - Remove design handler initialization
   - Update ServerConfig to remove design handler requirements
   - Handler creation for auth/org/webhook proceeds as before

3. **Update handler factory in `internal/interfaces/http/routes.go`**:
   - Remove or deprecate CreateDesignHandlers function
   - DesignHandlers struct becomes unused (can be removed or kept for reference)

### Phase 4: Update Main Entry Point

1. **Update `cmd/server/main.go`**:
   - Remove design handler initialization code
   - Remove design service/repository initialization
   - Keep auth, org, webhook handler initialization
   - GraphQL routes remain unchanged (served via Chi engine)

### Phase 5: Routing Architecture After Change

```
HTTP Server (port 8080)
├─ REST API Layer (Chi Router, oapi-codegen)
│  ├─ /api/auth/*        → Auth handler (unchanged)
│  ├─ /api/org/*         → Org handler (unchanged)
│  └─ /api/webhook/*     → Webhook handler (unchanged)
│
├─ GraphQL Layer (Chi Engine)
│  └─ /api/design/graphql → GraphQL endpoint (unchanged)
│
└─ OpenAPI Spec
   └─ /api/openapi.json → Spec for auth/org/webhook only
```

## Implementation Risks & Mitigations

### Risk 1: Breaking Change for REST API Clients
**Severity**: HIGH
**Mitigation**: 
- Provide 1-2 release notice in changelog
- Document migration path (show GraphQL equivalents)
- Support GraphQL endpoint for all business operations
- Consider gradual deprecation if client count is high

### Risk 2: Design Handlers Become Orphaned
**Severity**: MEDIUM  
**Mitigation**:
- Design handlers still exist (in designHandler package)
- They're just not called via HTTP REST route delegation
- GraphQL resolvers continue to use business services
- Plan for dedicated task to clean up HTTP handler layer if not needed

### Risk 3: Generated Code Compile Errors
**Severity**: LOW
**Mitigation**:
- Run `task generate-oapi` to verify
- Update type references in `server.go` incrementally
- Commit changes to generated files so diff is clear
- All changes are forward-compatible (removing methods from interface is safe)

### Risk 4: GraphQL Routes Stop Working
**Severity**: LOW
**Mitigation**:
- Chi engine wiring is independent of Chi/OpenAPI layer
- GraphQL routes registered separately in `main.go`
- Test GraphQL endpoint before/after changes
- Include GraphQL smoke tests in validation phase

## Migration Path for Consumers

### Before (REST-based for business logic)
```bash
# Get all models
GET /api/design/models?projectId=default

# Create model
POST /api/design/models/createModel {body}

# Update model
PUT /api/design/models/{id} {body}

# Delete model
DELETE /api/design/models/{id}
```

### After (GraphQL-based for business logic)
```bash
# Get all models (query)
POST /api/design/graphql {
  query {
    models(projectId: "default") { id name ... }
  }
}

# Create model (mutation)
POST /api/design/graphql {
  mutation {
    createModel(input: {...}) { id name ... }
  }
}

# Update model (mutation)
POST /api/design/graphql {
  mutation {
    updateModel(id: "...", input: {...}) { id name ... }
  }
}

# Delete model (mutation)
POST /api/design/graphql {
  mutation {
    deleteModel(id: "...") { success }
  }
}
```

## Rollback Plan

If issues arise during deployment:

1. Revert all changes to OpenAPI YAML files
2. Restore `internal/interfaces/http/generated/server.gen.go` from previous commit
3. Restore `internal/interfaces/http/server.go` handler methods
4. Restore handler initialization in `cmd/server/main.go`
5. Rebuild and redeploy: `task generate-oapi && task build && task run`

No database schema changes, no data migration required.

## Testing Strategy

1. **Unit Tests**: Verify auth/org handlers still work
2. **Integration Tests**: End-to-end auth flow, org creation
3. **GraphQL Tests**: Business logic via GraphQL (unchanged, existing tests cover)
4. **Negative Tests**: Verify removed REST endpoints return 404
5. **API Contract Tests**: Verify OpenAPI spec matches actual behavior

