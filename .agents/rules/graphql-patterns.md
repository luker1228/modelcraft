---
paths:
  - "internal/interfaces/graphql/**/*.go"
  - "api/graph/schema/**/*.graphql"
---

# API Design: GraphQL Patterns

Enforce GraphQL API design patterns including typed error handling, nullable fields, resolver implementation, and schema organization following industry best practices (GitHub, Shopify).

## Requirements

- Use typed error responses: define union types for mutation errors, implement Error interface with `message` field
- Return nullable data fields in mutation payloads (null on error, populated on success)
- Include single optional `error` field in payloads (not arrays)
- Resolvers must call application services, never repositories directly
- Define schemas in modular files under `api/graph/schema/` (base.graphql, model.graphql, etc.)
- Use error adapter pattern to convert `bizerrors.BusinessError` to GraphQL error types
- Keep error messages simple with optional `suggestion` field only when helpful

## Examples

### ✅ Good Example

```go
// GraphQL Schema - Typed error handling
// File: api/graph/schema/project.graphql

# Error interface - simple, message-first
interface Error {
  message: String!
}

# Specific error types
type ProjectNotFound implements Error {
  message: String!
}

type ProjectAlreadyExists implements Error {
  message: String!
  suggestion: String  # Optional, only when meaningful
}

type InvalidProjectInput implements Error {
  message: String!
  suggestion: String
}

# Union per mutation - documents all possible errors
union CreateProjectError = ProjectAlreadyExists | InvalidProjectInput
union UpdateProjectError = ProjectNotFound | InvalidProjectInput
union DeleteProjectError = ProjectNotFound | OperationDenied

# Payload with nullable data and single optional error
type CreateProjectPayload {
  project: Project        # Nullable - null on error
  error: CreateProjectError  # Single optional error (not array)
}

type Mutation {
  createProject(input: CreateProjectInput!): CreateProjectPayload!
  updateProject(id: ID!, input: UpdateProjectInput!): UpdateProjectPayload!
  deleteProject(id: ID!): DeleteProjectPayload!
}

// Resolver implementation
// File: internal/interfaces/graphql/project.resolvers.go

func (r *mutationResolver) CreateProject(ctx context.Context, input generated.CreateProjectInput) (*generated.CreateProjectPayload, error) {
    // Call application service (not repository)
    project, err := r.projectService.CreateProject(ctx, &app.CreateProjectInput{
        ID:    input.ID,
        Title: input.Title,
    })

    if err != nil {
        // Convert business error to GraphQL error using adapter
        if bizErr, ok := err.(*bizerrors.BusinessError); ok {
            return &generated.CreateProjectPayload{
                Project: nil,  // Null on error
                Error:   adapter.ConvertToCreateProjectError(bizErr),
            }, nil  // Return nil error - error in payload instead
        }
        return nil, err  // Unexpected system error
    }

    // Success case
    return &generated.CreateProjectPayload{
        Project: mapper.ToGraphQLProject(project),
        Error:   nil,
    }, nil
}

// Error adapter pattern
// File: internal/interfaces/graphql/adapter/project_error_adapter.go

func ConvertToCreateProjectError(bizErr *bizerrors.BusinessError) generated.CreateProjectError {
    switch bizErr.Info.GetCode() {
    case bizerrors.Conflict.GetCode():
        return &generated.ProjectAlreadyExists{
            Message:    bizErr.Msg(),
            Suggestion: ptr("Please use a different project ID"),
        }
    case bizerrors.ParamInvalid.GetCode():
        return &generated.InvalidProjectInput{
            Message:    bizErr.Msg(),
            Suggestion: ptr("Project ID must be lowercase with hyphens or underscores"),
        }
    default:
        return &generated.InvalidProjectInput{
            Message: bizErr.Msg(),
        }
    }
}

func ptr(s string) *string { return &s }
```

### ❌ Bad Example

```go
// Bad GraphQL Schema - Multiple anti-patterns

# ❌ Bad: Generic error with complex fields
type Error {
  code: String!
  message: String!
  path: [String!]  # Over-engineered
  extensions: JSON # Unnecessary complexity
}

# ❌ Bad: Array of errors instead of single error
type CreateProjectPayload {
  project: Project!  # ❌ Non-nullable - can't be null on error
  errors: [Error!]   # ❌ Array instead of single error
  success: Boolean!  # ❌ Redundant - use nullable project instead
}

# ❌ Bad: No typed errors - can't distinguish error types
type Mutation {
  createProject(input: CreateProjectInput!): CreateProjectPayload!
}

// Bad Resolver - Multiple anti-patterns
func (r *mutationResolver) CreateProject(ctx context.Context, input generated.CreateProjectInput) (*generated.CreateProjectPayload, error) {
    // ❌ Bad: Calling repository directly instead of service
    project, err := r.projectRepo.Create(ctx, &domain.Project{
        ID:    input.ID,
        Title: input.Title,
    })

    if err != nil {
        // ❌ Bad: Generic error handling, no typed errors
        return &generated.CreateProjectPayload{
            Project: nil,
            Errors: []*generated.Error{
                {
                    Code:    "UNKNOWN_ERROR",
                    Message: "Something went wrong",
                },
            },
            Success: false,
        }, nil
    }

    // ❌ Bad: Redundant success flag
    return &generated.CreateProjectPayload{
        Project: project,
        Errors:  []*generated.Error{},
        Success: true,
    }, nil
}
```

## Rationale

Typed GraphQL errors provide type-safe error handling, enable better client code generation, and make APIs self-documenting. Nullable data fields follow GraphQL best practices (GitHub, Shopify patterns). Single error fields keep responses simple while union types document all possible error cases. Error adapters decouple business errors from GraphQL types.

---

See skill: `backend-patterns` for comprehensive GraphQL patterns and error handling strategies.
