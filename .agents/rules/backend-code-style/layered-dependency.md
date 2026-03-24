---
paths:
  - "internal/**/*.go"
---

# Code Style: Layered Dependency

Enforce strict unidirectional dependencies between architectural layers to maintain clean separation of concerns.

## Requirements

- **Interfaces Layer** (`internal/interfaces/`) MAY depend on:
  - Same layer (Interfaces)
  - Application Layer (`internal/app/`)
  - MUST NOT depend on Infrastructure Layer

- **Application Layer** (`internal/app/`) MAY depend on:
  - Domain Layer (`internal/domain/`)
  - Infrastructure Layer (`internal/infrastructure/`)
  - MUST NOT depend on Interfaces Layer

- **Infrastructure Layer** (`internal/infrastructure/`) MAY depend on:
  - Domain Layer (`internal/domain/`)
  - MUST NOT depend on Application or Interfaces Layer

- **Domain Layer** (`internal/domain/`) MUST NOT depend on any other layer:
  - MUST NOT import from `internal/interfaces/`
  - MUST NOT import from `internal/app/`
  - MUST NOT import from `internal/infrastructure/`
  - MAY only import from `pkg/` (shared utilities)

- **Shared Kernel** (`pkg/`) MUST NOT depend on any internal layer

## Dependency Rules Summary

```
┌─────────────────┐
│   Interfaces    │  ← HTTP handlers, GraphQL resolvers
│  (interfaces/)  │
└────────┬────────┘
         │ depends on
         ↓
┌─────────────────┐
│   Application   │  ← Use cases, orchestration
│     (app/)      │
└────┬────────────┘
     │ depends on
     ↓            ↘
┌─────────────┐  ┌──────────────────┐
│   Domain    │  │ Infrastructure   │  ← Repositories, DB
│  (domain/)  │  │ (infrastructure/)│
└─────────────┘  └────────┬─────────┘
                          │ depends on
                          ↓
                   ┌─────────────┐
                   │   Domain    │
                   │  (domain/)  │
                   └─────────────┘
```

## Examples

### ✅ Good Example

```go
// internal/interfaces/graphql/resolver/model_resolver.go
package resolver

import (
    "modelcraft/internal/app/modeldesign"  // ✅ Interfaces → App
    "modelcraft/internal/interfaces/graphql/model"  // ✅ Same layer
)

type ModelResolver struct {
    modelService *modeldesign.ModelService  // ✅ Depends on App layer
}
```

```go
// internal/app/modeldesign/model_service.go
package modeldesign

import (
    "modelcraft/internal/domain/modeldesign"  // ✅ App → Domain
    "modelcraft/internal/infrastructure/persistence"  // ✅ App → Infrastructure
)

type ModelService struct {
    modelRepo persistence.ModelRepository  // ✅ Depends on Infrastructure
    validator modeldesign.ModelValidator   // ✅ Depends on Domain
}
```

```go
// internal/infrastructure/persistence/model_repository.go
package persistence

import (
    "modelcraft/internal/domain/modeldesign"  // ✅ Infrastructure → Domain
)

type ModelRepository struct {}

func (r *ModelRepository) Save(model *modeldesign.Model) error {
    // ✅ Uses domain entities
    return nil
}
```

```go
// internal/domain/modeldesign/model.go
package modeldesign

import (
    "modelcraft/pkg/bizerrors"  // ✅ Domain → Shared Kernel only
)

type Model struct {
    ID   string
    Name string
}

func (m *Model) Validate() error {
    if m.Name == "" {
        return bizerrors.New("MODEL_INVALID", "name is required")
    }
    return nil
}
```

### ❌ Bad Example

```go
// internal/domain/modeldesign/model.go
package modeldesign

import (
    "modelcraft/internal/infrastructure/persistence"  // ❌ Domain → Infrastructure (FORBIDDEN)
)

type Model struct {
    repo persistence.ModelRepository  // ❌ Domain should not know about infrastructure
}
```

```go
// internal/infrastructure/persistence/model_repository.go
package persistence

import (
    "modelcraft/internal/app/modeldesign"  // ❌ Infrastructure → App (FORBIDDEN)
)

type ModelRepository struct {
    service *modeldesign.ModelService  // ❌ Reverse dependency
}
```

```go
// internal/app/modeldesign/model_service.go
package modeldesign

import (
    "modelcraft/internal/interfaces/graphql/model"  // ❌ App → Interfaces (FORBIDDEN)
)

func (s *ModelService) CreateModel() *model.ModelPayload {
    // ❌ Application layer should not depend on interface layer types
    return nil
}
```

```go
// internal/domain/modeldesign/validator.go
package modeldesign

import (
    "modelcraft/internal/app/modeldesign"  // ❌ Domain → App (FORBIDDEN)
)

type Validator struct {
    service *modeldesign.ModelService  // ❌ Domain depending on App layer
}
```

## Rationale

Unidirectional layer dependencies prevent circular dependencies, make code easier to test (inner layers can be tested without outer layers), and enforce the Dependency Inversion Principle where high-level modules don't depend on low-level implementation details.

---

See skill: `backend-patterns` for comprehensive DDD architecture guidance.
