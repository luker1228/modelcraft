# graphql-directive Specification

## Purpose
TBD - created by archiving change implement-casbin-permissions. Update Purpose after archive.
## Requirements
### Requirement: GraphQL Permission Directive Definition

The system SHALL provide a `@hasPermission` GraphQL directive that can be applied to field definitions to enforce operation-level authorization. The directive accepts an `action` parameter in `resource:operation` format.

#### Scenario: Directive definition in schema

- **GIVEN** the GraphQL schema is defined in api/graph/schema/base.graphql
- **WHEN** the schema is loaded
- **THEN** the directive is defined as `directive @hasPermission(action: String!) on FIELD_DEFINITION`
- **AND** the directive is registered in gqlgen.yml with resolver function

#### Scenario: Directive applied to mutation

- **GIVEN** a GraphQL mutation `createProject` is defined
- **WHEN** the mutation is annotated with `@hasPermission(action: "project:create")`
- **THEN** the directive is enforced before the resolver executes
- **AND** the resolver only executes if permission check passes

#### Scenario: Directive applied to query

- **GIVEN** a GraphQL query `projects` is defined
- **WHEN** the query is annotated with `@hasPermission(action: "project:read")`
- **THEN** the directive is enforced before the resolver executes
- **AND** the resolver only executes if permission check passes

### Requirement: User Context Extraction

The directive SHALL extract user identity (user_id, org_name) from the GraphQL context to perform permission checks. Context values are set by JWT authentication middleware.

#### Scenario: Extract user ID from context

- **GIVEN** a GraphQL request with authenticated user
- **AND** JWT middleware has set user_id="user123" in Chi context
- **WHEN** the @hasPermission directive executes
- **THEN** the directive calls GetUserIDFromContext(ctx)
- **AND** returns user_id="user123" successfully

#### Scenario: Extract organization name from context

- **GIVEN** a GraphQL request with authenticated user
- **AND** JWT middleware has set org_name="org1" in Chi context
- **WHEN** the @hasPermission directive executes
- **THEN** the directive calls GetOrgNameFromContext(ctx)
- **AND** returns org_name="org1" successfully

#### Scenario: Missing user context

- **GIVEN** a GraphQL request without authentication
- **WHEN** the @hasPermission directive attempts to extract user_id
- **THEN** GetUserIDFromContext returns error "user ID not found in context"
- **AND** the directive returns error "permission denied: user not authenticated"
- **AND** the resolver is not executed

### Requirement: Casbin Enforcer Integration

The directive SHALL call Casbin enforcer to check if the user has the required permission based on their assigned roles and organization context.

#### Scenario: Permission check success

- **GIVEN** a user with user_id="user123" in org_name="org1"
- **AND** the user has role "admin" which grants "project:create" permission
- **WHEN** the directive checks action="project:create"
- **THEN** the directive parses action into obj="project" and act="create"
- **AND** calls enforcer.Enforce(user_id, obj, act)
- **AND** enforcer returns true
- **AND** the directive allows the resolver to execute

#### Scenario: Permission check failure

- **GIVEN** a user with user_id="user456" in org_name="org1"
- **AND** the user has role "viewer" which does not grant "project:delete" permission
- **WHEN** the directive checks action="project:delete"
- **THEN** the directive parses action into obj="project" and act="delete"
- **AND** calls enforcer.Enforce(user_id, obj, act)
- **AND** enforcer returns false
- **AND** the directive returns error "permission denied: requires 'project:delete' in organization 'org1'"
- **AND** the resolver is not executed

#### Scenario: Enforcer error handling

- **GIVEN** the Casbin enforcer encounters an internal error
- **WHEN** enforcer.Enforce() is called
- **THEN** the directive logs the error with logfacade
- **AND** returns error "internal authorization error"
- **AND** the resolver is not executed

### Requirement: Action Format Validation

The directive SHALL validate that the action parameter follows the `resource:operation` format and reject invalid formats with a clear error message.

#### Scenario: Valid action format

- **GIVEN** a directive with action="project:create"
- **WHEN** the directive parses the action parameter
- **THEN** obj="project" and act="create" are extracted successfully
- **AND** the permission check proceeds

#### Scenario: Invalid action format without colon

- **GIVEN** a directive with action="createproject"
- **WHEN** the directive parses the action parameter
- **THEN** the directive returns error "invalid action format: must be 'resource:operation'"
- **AND** the resolver is not executed

#### Scenario: Invalid action format with multiple colons

- **GIVEN** a directive with action="project:model:create"
- **WHEN** the directive parses the action parameter
- **THEN** the directive returns error "invalid action format: must be 'resource:operation'"
- **AND** the resolver is not executed

### Requirement: Error Response Format

The directive SHALL return structured GraphQL errors with clear permission denial messages that include the required action and organization context.

#### Scenario: Permission denied error format

- **GIVEN** a user without required permission
- **WHEN** the directive denies access
- **THEN** the error message is "permission denied: requires '{action}' in organization '{org_name}'"
- **AND** the error is returned as a GraphQL error
- **AND** the HTTP status code is 200 (GraphQL convention)
- **AND** the error appears in the "errors" array of the GraphQL response

#### Scenario: Authentication error format

- **GIVEN** a request without valid authentication
- **WHEN** the directive cannot extract user context
- **THEN** the error message is "permission denied: user not authenticated"
- **AND** the error is returned as a GraphQL error

#### Scenario: Internal error format

- **GIVEN** an internal error occurs in the enforcer
- **WHEN** the directive handles the error
- **THEN** the error message is "internal authorization error"
- **AND** the detailed error is logged but not exposed to the client
- **AND** the error is returned as a GraphQL error

### Requirement: Directive Registration

The system SHALL register the @hasPermission directive in gqlgen configuration so that gqlgen generates the necessary code for directive enforcement.

#### Scenario: Directive registered in gqlgen.yml

- **GIVEN** gqlgen.yml configuration file
- **WHEN** the configuration is loaded
- **THEN** the directives section includes:
  ```yaml
  directives:
    hasPermission:
      from: modelcraft/internal/interfaces/graphql
      resolver: HasPermission
  ```
- **AND** the resolver function is HasPermission in internal/interfaces/graphql/directives.go

#### Scenario: Generate GraphQL code with directive

- **GIVEN** the directive is registered in gqlgen.yml
- **WHEN** `task generate-gql` is executed
- **THEN** gqlgen generates code in internal/interfaces/graphql/generated/generated.go
- **AND** the generated code calls the HasPermission resolver function for annotated fields
- **AND** no compilation errors occur

### Requirement: Backward Compatibility

The directive SHALL coexist with the existing JWT permission check functions (RequirePermissionFromContext) to support gradual migration without breaking existing functionality.

#### Scenario: Directive and middleware coexist

- **GIVEN** a GraphQL operation uses @hasPermission directive
- **AND** a REST endpoint uses RequirePermission middleware
- **WHEN** both are accessed with valid authentication
- **THEN** both permission systems work independently
- **AND** no conflicts or errors occur

#### Scenario: Gradual migration path

- **GIVEN** some GraphQL operations use @hasPermission directive
- **AND** other GraphQL operations use RequirePermissionFromContext in resolver code
- **WHEN** users invoke both types of operations
- **THEN** both permission checks function correctly
- **AND** no operations are left unprotected during migration

