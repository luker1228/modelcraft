## ADDED Requirements

### Requirement: List available database tables
The system SHALL provide a `listTables` GraphQL query that returns all base tables in a specified database from the project's connected cluster, optionally excluding tables that already have a corresponding model.

#### Scenario: Returns unimported tables
- **WHEN** `listTables` is called with a valid `projectSlug`, `databaseName`, and `excludeExisting: true`
- **THEN** the system returns a list of `TableInfo` objects where none of the returned table names have an existing model with the same name in that project+database

#### Scenario: Returns all tables when excludeExisting is false
- **WHEN** `listTables` is called with `excludeExisting: false`
- **THEN** the system returns all base tables in the database, including those already imported as models

#### Scenario: excludeExisting defaults to true
- **WHEN** `listTables` is called without the `excludeExisting` field
- **THEN** the system behaves as if `excludeExisting: true` was passed

#### Scenario: Returns empty list when all tables are imported
- **WHEN** `listTables` is called and every table in the database already has a model
- **THEN** the system returns an empty list

#### Scenario: Requires project:read permission
- **WHEN** a user without `project:read` permission calls `listTables`
- **THEN** the system returns an authorization error

#### Scenario: Only returns TABLE_TYPE = BASE TABLE
- **WHEN** the database contains views or system tables
- **THEN** `listTables` SHALL NOT include views or system tables in the result
