# Change: Optimize Project GraphQL Error Definitions

## Why

The current GraphQL API for Project operations returns simple boolean success flags and nullable data fields in payload types, making it difficult for clients to distinguish between different error scenarios. Clients cannot differentiate between "project not found" vs "project ID already exists" vs other validation errors without inspecting generic error messages.

Following GraphQL best practices (similar to GitHub's GraphQL API and Shopify's patterns), we should introduce structured, typed error handling using union types and error interfaces. This enables clients to:
- Handle specific error cases programmatically
- Provide targeted user feedback
- Implement proper retry logic based on error types

## What Changes

- Introduce `Error` interface with `message` field for all user-facing errors
- Define specific error types for Project operations:
  - `ProjectAlreadyExists` - Project ID conflict (CONFLICT.PROJECT)
  - `ProjectNotFound` - Project does not exist (NOT_FOUND.PROJECT)
  - `InvalidProjectInput` - Invalid parameters (PARAM_INVALID.PROJECT)
  - `CannotDeleteDefaultProject` - Operation denied on default project (OPERATION_DENIED.PROJECT)
- All error types support optional `suggestion: String` field for helpful guidance
- Create union types for each mutation's possible errors:
  - `CreateProjectError` union
  - `UpdateProjectError` union
  - `DeleteProjectError` union
  - `ArchiveProjectError` union
  - `ActivateProjectError` union
- Update payload types to include typed `errors` field instead of relying solely on nullable data
- Maintain backward compatibility by keeping existing `success` fields where appropriate

This change aligns with the existing error handling system in `pkg/bizerrors/` which already defines these error codes and types.

**Design Philosophy:**
- **Keep it simple**: Errors only need `message` (required) and `suggestion` (optional)
- **Avoid over-engineering**: No complex fields like `path`, `code`, error-specific IDs
- **Message-first**: All context (like project ID) is included in the message text
- **Optional suggestions**: Provide helpful guidance only when meaningful

## Impact

**Affected specs:**
- `project-management` (new spec to be created)

**Affected code:**
- `api/graph/schema/project.graphql` - GraphQL schema definitions
- `internal/interfaces/graphql/project.resolvers.go` - Resolver implementations
- `internal/interfaces/graphql/adapter/project_error_adapter.go` - New error adapter for converting bizerrors to GraphQL errors
- `pkg/bizerrors/common_errors.go` - May need to add ProjectNotFoundError definition if not exists

**Breaking changes:**
- **None** - This is an additive change. Existing fields (`success`, nullable project data) remain unchanged
- New `errors` field provides additional typed error information
- Clients can migrate incrementally to use typed errors

**Client migration path:**
- Old clients: Continue using `success: Boolean!` and nullable `project` fields
- New clients: Check `errors` array first for typed error handling, fallback to nullable fields

**Testing impact:**
- Existing integration tests continue to work
- New tests needed for each error scenario
- GraphQL schema validation tests required
