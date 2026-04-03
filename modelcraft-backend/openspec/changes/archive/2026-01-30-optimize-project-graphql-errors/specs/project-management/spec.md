# Project Management Capability

## ADDED Requirements

### Requirement: Typed GraphQL Error Responses

The Project GraphQL API SHALL provide structured, typed error responses for all mutations, enabling clients to handle specific error scenarios programmatically.

#### Scenario: Create project with existing ID returns typed conflict error

- **WHEN** client calls `createProject` mutation with a project ID that already exists
- **THEN** the response SHALL include a `ProjectAlreadyExists` in the `errors` field
- **AND** the error SHALL contain `message` field with descriptive text
- **AND** the error MAY contain `suggestion` field
- **AND** the `project` field SHALL be null
- **AND** the error type SHALL be distinguishable via `__typename`

#### Scenario: Create project with invalid input returns typed validation error

- **WHEN** client calls `createProject` mutation with invalid input (e.g., empty title, invalid ID format)
- **THEN** the response SHALL include an `InvalidProjectInput` in the `errors` field
- **AND** the error SHALL contain `message` field describing which input is invalid
- **AND** the error MAY contain `suggestion` field with guidance
- **AND** the `project` field SHALL be null

#### Scenario: Update non-existent project returns typed not found error

- **WHEN** client calls `updateProject` mutation with a project ID that does not exist
- **THEN** the response SHALL include a `ProjectNotFound` in the `errors` field
- **AND** the error SHALL contain `message` field (including the project ID in the message)
- **AND** the `project` field SHALL be null

#### Scenario: Delete default project returns typed operation denied error

- **WHEN** client calls `deleteProject` mutation attempting to delete the default project
- **THEN** the response SHALL include a `CannotDeleteDefaultProject` in the `errors` field
- **AND** the error SHALL contain `message` field
- **AND** the `success` field SHALL be false

#### Scenario: Successful operation returns empty errors array

- **WHEN** any project mutation succeeds without errors
- **THEN** the response SHALL include an empty `errors` array
- **AND** the data field SHALL contain the expected result (project or success status)

### Requirement: Error Interface

All user-facing errors SHALL implement a common `Error` interface with a `message` field for consistent error handling.

#### Scenario: All error types implement Error interface

- **WHEN** any project operation returns an error
- **THEN** the error SHALL implement the `Error` interface
- **AND** the error SHALL have a `message: String!` field with a human-readable description

#### Scenario: Client queries error message without knowing specific type

- **WHEN** client queries errors using the `Error` interface
- **THEN** client SHALL be able to retrieve `message` for any error type
- **AND** client can use fragments for specific error types to get additional fields like `suggestion`

### Requirement: Project-Specific Error Types

The API SHALL define specific error types for Project domain operations with clear messages and optional suggestions.

#### Scenario: ProjectAlreadyExists provides clear message

- **WHEN** `ProjectAlreadyExists` is returned
- **THEN** it SHALL include `message: String!` describing the conflict
- **AND** it MAY include `suggestion: String` with guidance (e.g., "Please use a different project ID")

#### Scenario: ProjectNotFound provides clear message

- **WHEN** `ProjectNotFound` is returned
- **THEN** it SHALL include `message: String!` describing the not found error (including the project ID in message text)

#### Scenario: InvalidProjectInput provides validation details

- **WHEN** `InvalidProjectInput` is returned
- **THEN** it SHALL include `message: String!` describing the validation failure
- **AND** it MAY include `suggestion: String` with guidance for fixing the input (e.g., "Project ID must be lowercase letters, numbers, and hyphens")

#### Scenario: CannotDeleteDefaultProject explains operation restriction

- **WHEN** `CannotDeleteDefaultProject` is returned
- **THEN** it SHALL include `message: String!` explaining why the operation is denied
- **AND** it SHALL clarify that the default project cannot be deleted

### Requirement: Union Types for Mutation-Specific Errors

Each project mutation SHALL define a union type containing all possible error types for that specific operation.

#### Scenario: CreateProject mutation has specific error union

- **WHEN** client introspects `CreateProjectError` union type
- **THEN** it SHALL include `ProjectAlreadyExists` type
- **AND** it SHALL include `InvalidProjectInput` type
- **AND** these SHALL be the only possible error types for createProject operation

#### Scenario: UpdateProject mutation has specific error union

- **WHEN** client introspects `UpdateProjectError` union type
- **THEN** it SHALL include `ProjectNotFound` type
- **AND** it SHALL include `InvalidProjectInput` type
- **AND** these SHALL be the only possible error types for updateProject operation

#### Scenario: DeleteProject mutation has specific error union

- **WHEN** client introspects `DeleteProjectError` union type
- **THEN** it SHALL include `ProjectNotFound` type
- **AND** it SHALL include `CannotDeleteDefaultProject` type
- **AND** these SHALL be the only possible error types for deleteProject operation

#### Scenario: ArchiveProject mutation has specific error union

- **WHEN** client introspects `ArchiveProjectError` union type
- **THEN** it SHALL include `ProjectNotFound` type
- **AND** this SHALL be the only possible error type for archiveProject operation

#### Scenario: ActivateProject mutation has specific error union

- **WHEN** client introspects `ActivateProjectError` union type
- **THEN** it SHALL include `ProjectNotFound` type
- **AND** this SHALL be the only possible error type for activateProject operation

### Requirement: Backward Compatible Payload Structure

All project mutation payloads SHALL maintain existing fields while adding new typed error fields, ensuring zero breaking changes for existing clients.

#### Scenario: CreateProjectPayload includes both old and new error handling

- **WHEN** `createProject` mutation completes
- **THEN** the payload SHALL include `project: Project` field (nullable, null on error)
- **AND** the payload SHALL include `errors: [CreateProjectError!]!` field (empty array on success)
- **AND** existing clients using only `project` field SHALL continue to work

#### Scenario: UpdateProjectPayload includes both old and new error handling

- **WHEN** `updateProject` mutation completes
- **THEN** the payload SHALL include `success: Boolean!` field (backward compatible)
- **AND** the payload SHALL include `project: Project` field (nullable, null on error)
- **AND** the payload SHALL include `errors: [UpdateProjectError!]!` field (empty array on success)

#### Scenario: DeleteProjectPayload includes both old and new error handling

- **WHEN** `deleteProject` mutation completes
- **THEN** the payload SHALL include `success: Boolean!` field (backward compatible)
- **AND** the payload SHALL include `errors: [DeleteProjectError!]!` field (empty array on success)

#### Scenario: ArchiveProjectPayload includes both old and new error handling

- **WHEN** `archiveProject` mutation completes
- **THEN** the payload SHALL include `success: Boolean!` field (backward compatible)
- **AND** the payload SHALL include `errors: [ArchiveProjectError!]!` field (empty array on success)

#### Scenario: ActivateProjectPayload includes both old and new error handling

- **WHEN** `activateProject` mutation completes
- **THEN** the payload SHALL include `success: Boolean!` field (backward compatible)
- **AND** the payload SHALL include `errors: [ActivateProjectError!]!` field (empty array on success)

### Requirement: Error Adapter Integration

The GraphQL resolver layer SHALL integrate with the existing `pkg/bizerrors` error classification system to convert domain errors to typed GraphQL errors.

#### Scenario: bizerrors CONFLICT.PROJECT maps to ProjectAlreadyExists

- **WHEN** a domain service returns a `bizerrors.BusinessError` with code `CONFLICT.PROJECT`
- **THEN** the error adapter SHALL convert it to `ProjectAlreadyExists` GraphQL type
- **AND** SHALL populate `message` from `bizErr.Msg()`
- **AND** MAY populate `suggestion` based on error context

#### Scenario: bizerrors NOT_FOUND.PROJECT maps to ProjectNotFound

- **WHEN** a domain service returns a `bizerrors.BusinessError` with code `NOT_FOUND.PROJECT`
- **THEN** the error adapter SHALL convert it to `ProjectNotFound` GraphQL type
- **AND** SHALL populate `message` from `bizErr.Msg()`

#### Scenario: bizerrors PARAM_INVALID.PROJECT maps to InvalidProjectInput

- **WHEN** a domain service returns a `bizerrors.BusinessError` with code `PARAM_INVALID.PROJECT`
- **THEN** the error adapter SHALL convert it to `InvalidProjectInput` GraphQL type
- **AND** SHALL populate `message` from `bizErr.Msg()`
- **AND** MAY populate `suggestion` based on validation context

#### Scenario: bizerrors OPERATION_DENIED.PROJECT maps to CannotDeleteDefaultProject

- **WHEN** a domain service returns a `bizerrors.BusinessError` with code `OPERATION_DENIED.PROJECT`
- **AND** the operation is deleting the default project
- **THEN** the error adapter SHALL convert it to `CannotDeleteDefaultProject` GraphQL type
- **AND** SHALL populate `message` from `bizErr.Msg()`

#### Scenario: Unknown bizerrors code returns system error

- **WHEN** a domain service returns a `bizerrors.BusinessError` with an unknown or unmapped code
- **THEN** the error adapter SHALL log a warning
- **AND** SHALL return a generic system error to the client
- **AND** SHALL not expose internal error details in the GraphQL response
