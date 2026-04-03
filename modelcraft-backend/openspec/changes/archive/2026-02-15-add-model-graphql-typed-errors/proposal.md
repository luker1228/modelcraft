# Change: Add Typed Error Handling for Model Operations

## Why

The current GraphQL API for Model operations returns simple boolean success flags and nullable data fields in payload types, making it difficult for clients to distinguish between different error scenarios. Clients cannot differentiate between "model not found" vs "model name already exists" vs "project not found" vs "cluster not found" or other validation errors without inspecting generic error messages.

Following GraphQL best practices (similar to GitHub's and Shopify's patterns) and the existing error handling patterns already implemented for Project, Cluster, and Enum, we should introduce structured, typed error handling using union types and error interfaces. This enables clients to:

- Handle specific error cases programmatically
- Provide targeted user feedback
- Implement proper retry logic based on error types

## What Changes

- Define model-specific error types in GraphQL schema:
  - `ModelAlreadyExists` - Model name conflict within project (CONFLICT.MODEL)
  - `ModelNotFound` - Model does not exist (NOT_FOUND.MODEL)
  - `InvalidModelInput` - Invalid parameters (PARAM_INVALID)
  - `ProjectNotFound` - Project does not exist (NOT_FOUND.PROJECT) - reuse from project.graphql
  - `ClusterNotFound` - Cluster does not exist (NOT_FOUND.CLUSTER) - reuse from cluster.graphql

- For deletion operations:
  - `CannotDeleteDeployedModel` - Cannot delete deployed model (OPERATION_DENIED.MODEL)

- Create union types for each mutation's possible errors:
  - `GetModelError` union
  - `CreateModelError` union
  - `UpdateModelError` union
  - `DeleteModelError` union

- Update payload types to include optional `error` field of the corresponding error union type

- Create error adapter `internal/interfaces/graphql/adapter/model_error_adapter.go` for converting bizerrors to GraphQL errors

**Design Philosophy:**
- **Keep it simple**: Errors only need `message` (required) and `suggestion` (optional)
- **Avoid over-engineering**: No complex fields like `path`, `code`, error-specific IDs
- **Message-first**: All context (like model ID, project ID) is included in the message text
- **Optional suggestions**: Provide helpful guidance only when meaningful

This change aligns with the existing error handling system in `pkg/bizerrors/` which already defines `ModelNotFound` and `ModelAlreadyExists` error codes.

## Impact

**Affected specs:**
- `modeldesign-schema-operations` - will add typed error requirements

**Affected code:**
- `api/graph/schema/model.graphql` - GraphQL schema definitions (main changes)
- `internal/interfaces/graphql/model.resolvers.go` - Resolver implementations to use typed errors
- `internal/interfaces/graphql/adapter/model_error_adapter.go` - New error adapter for converting bizerrors to GraphQL errors

**Breaking changes:**
- **None** - This is an additive change. Existing fields (`success`, nullable model data) remain unchanged
- New `error` field provides additional typed error information
- Clients can migrate incrementally to use typed errors

**Client migration path:**
- Old clients: Continue using `success: Boolean!` and nullable `model` fields
- New clients: Check `error` field first for typed error handling, fallback to nullable fields

**Note**: The design follows the pattern from `enum-error-handling` which uses a single optional `error` field (not an array), which is simpler and consistent with the project and cluster error handling patterns.
