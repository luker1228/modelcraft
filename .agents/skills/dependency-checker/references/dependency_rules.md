# Layered Dependency Rules

This document defines the strict unidirectional dependency rules for ModelCraft's DDD layered architecture.

## Architecture Layers

```
┌─────────────────┐
│   Interfaces    │  ← HTTP handlers, GraphQL resolvers, DTOs
│  (interfaces/)  │
└────────┬────────┘
         │ depends on
         ↓
┌─────────────────┐
│   Application   │  ← Use cases, transaction orchestration
│     (app/)      │
└────┬────────────┘
     │ depends on
     ↓            ↘
┌─────────────┐  ┌──────────────────┐
│   Domain    │  │ Infrastructure   │  ← Repositories, DB access
│  (domain/)  │  │ (infrastructure/)│
└─────────────┘  └────────┬─────────┘
                          │ depends on
                          ↓
                   ┌─────────────┐
                   │   Domain    │
                   │  (domain/)  │
                   └─────────────┘
```

## Layer Dependency Rules

### Domain Layer (`internal/domain/`)

**Allowed Dependencies:**
- Standard library (`$gostd`)
- Shared kernel (`modelcraft/pkg`)
- Other domain packages (`modelcraft/internal/domain`)

**Forbidden Dependencies:**
- ❌ Application layer (`modelcraft/internal/app`)
- ❌ Infrastructure layer (`modelcraft/internal/infrastructure`)
- ❌ Interfaces layer (`modelcraft/internal/interfaces`)

**Rationale:** Domain layer contains pure business logic and must remain independent of all other layers to ensure testability and portability.

### Infrastructure Layer (`internal/infrastructure/`)

**Allowed Dependencies:**
- Standard library (`$gostd`)
- Shared kernel (`modelcraft/pkg`)
- Domain layer (`modelcraft/internal/domain`)
- Other infrastructure packages (`modelcraft/internal/infrastructure`)

**Forbidden Dependencies:**
- ❌ Application layer (`modelcraft/internal/app`)
- ❌ Interfaces layer (`modelcraft/internal/interfaces`)

**Rationale:** Infrastructure implements domain interfaces but must not depend on higher layers to maintain clean separation of concerns.

### Application Layer (`internal/app/`)

**Allowed Dependencies:**
- Standard library (`$gostd`)
- Shared kernel (`modelcraft/pkg`)
- Domain layer (`modelcraft/internal/domain`)
- Infrastructure layer (`modelcraft/internal/infrastructure`)
- Other application packages (`modelcraft/internal/app`)

**Forbidden Dependencies:**
- ❌ Interfaces layer (`modelcraft/internal/interfaces`)

**Rationale:** Application layer orchestrates use cases and can use both domain and infrastructure, but must not know about presentation concerns.

### Interfaces Layer (`internal/interfaces/`)

**Allowed Dependencies:**
- Standard library (`$gostd`)
- Shared kernel (`modelcraft/pkg`)
- Application layer (`modelcraft/internal/app`)
- Other interface packages (`modelcraft/internal/interfaces`)

**Forbidden Dependencies:**
- ❌ Infrastructure layer directly (must go through Application layer)

**Rationale:** Interfaces handle HTTP/GraphQL concerns and delegate to application layer. Direct infrastructure access bypasses business logic.

## Common Violations and Fixes

### Violation 1: Domain depending on Infrastructure

```go
// ❌ BAD: internal/domain/modeldesign/model.go
package modeldesign

import "modelcraft/internal/infrastructure/persistence"

type Model struct {
    repo persistence.ModelRepository  // Domain knows about infrastructure
}
```

**Fix:** Use dependency injection through interfaces defined in domain

```go
// ✅ GOOD: internal/domain/modeldesign/repository.go
package modeldesign

// Repository interface defined in domain
type ModelRepository interface {
    Save(model *Model) error
    FindByID(id string) (*Model, error)
}

// ✅ GOOD: internal/infrastructure/persistence/model_repository.go
package persistence

import "modelcraft/internal/domain/modeldesign"

// Infrastructure implements domain interface
type ModelRepository struct {
    db *sql.DB
}

func (r *ModelRepository) Save(model *modeldesign.Model) error {
    // Implementation
}
```

### Violation 2: Infrastructure depending on Application

```go
// ❌ BAD: internal/infrastructure/persistence/model_repository.go
package persistence

import "modelcraft/internal/app/modeldesign"

type ModelRepository struct {
    service *modeldesign.ModelService  // Infrastructure depends on app
}
```

**Fix:** Remove the dependency. Infrastructure should only know about domain

```go
// ✅ GOOD: internal/infrastructure/persistence/model_repository.go
package persistence

import "modelcraft/internal/domain/modeldesign"

type ModelRepository struct {
    db *sql.DB
}

func (r *ModelRepository) Save(model *modeldesign.Model) error {
    // Implementation using domain entities only
}
```

### Violation 3: Application depending on Interfaces

```go
// ❌ BAD: internal/app/modeldesign/model_service.go
package modeldesign

import "modelcraft/internal/interfaces/graphql/model"

func (s *ModelService) CreateModel() *model.CreateModelPayload {
    // Application returns interface layer type
}
```

**Fix:** Use domain types in application layer, convert in interfaces layer

```go
// ✅ GOOD: internal/app/modeldesign/model_service.go
package modeldesign

import "modelcraft/internal/domain/modeldesign"

func (s *ModelService) CreateModel() (*modeldesign.Model, error) {
    // Application returns domain type
}

// ✅ GOOD: internal/interfaces/graphql/resolver/model_resolver.go
package resolver

func (r *Resolver) CreateModel(ctx context.Context) (*model.CreateModelPayload, error) {
    domainModel, err := r.modelService.CreateModel()
    if err != nil {
        return nil, err
    }
    // Convert domain type to GraphQL type
    return convertToGraphQL(domainModel), nil
}
```

### Violation 4: Interfaces depending on Infrastructure directly

```go
// ❌ BAD: internal/interfaces/http/handler.go
package http

import "modelcraft/internal/infrastructure/persistence"

type Handler struct {
    repo persistence.ModelRepository  // Interfaces bypassing application layer
}
```

**Fix:** Depend on application layer instead

```go
// ✅ GOOD: internal/interfaces/http/handler.go
package http

import "modelcraft/internal/app/modeldesign"

type Handler struct {
    service *modeldesign.ModelService  // Use application service
}
```

## Testing Dependencies

Test files (`*_test.go`) have relaxed dependency rules since they need to test the implementation:

- Tests can import from the same layer and below
- Tests can use test doubles/mocks from any layer
- Tests should still prefer testing through public interfaces

## Enforcement

Use `golangci-lint` with `depguard` linter to automatically detect violations:

```bash
# Check dependencies
golangci-lint run --disable-all --enable=depguard ./...

# Or use the helper script
python scripts/check_dependencies.py
```

## References

- See `.claude/rules/code-style/layered-dependency.md` for rule enforcement
- See `backend-patterns` skill for comprehensive DDD architecture guidance
