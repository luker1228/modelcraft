## ADDED Requirements

### Requirement: Automatic Database Migration on Startup

The application SHALL automatically apply database schema migrations on startup using Atlas Go SDK.

#### Scenario: Successful migration on first startup
- **WHEN** the application starts with an empty or missing database
- **THEN** the application automatically creates the database if it doesn't exist
- **AND** applies all schema migrations from `db/schema/mysql/` directory
- **AND** logs the migration progress and success message
- **AND** continues to start the application normally

#### Scenario: Existing database with matching schema
- **WHEN** the application starts and the database schema matches the migration files
- **THEN** Atlas reports no changes needed
- **AND** the application logs "Database schema is up to date"
- **AND** continues to start normally

#### Scenario: Existing database with schema drift
- **WHEN** the application starts and the database schema differs from migration files
- **THEN** Atlas applies the necessary schema changes
- **AND** logs each SQL statement being executed
- **AND** the application applies the changes automatically (with `--auto-approve`)
- **AND** continues to start normally after migration completes

#### Scenario: Migration failure
- **WHEN** database migration fails due to SQL errors, connection issues, or schema conflicts
- **THEN** the application logs the detailed error message
- **AND** exits with a non-zero status code
- **AND** does not continue running

### Requirement: Database Auto-Creation

The application SHALL automatically create the target database if it doesn't exist.

#### Scenario: Database does not exist
- **WHEN** the configured database does not exist on the MySQL server
- **THEN** the application connects to MySQL without specifying a database
- **AND** executes `CREATE DATABASE IF NOT EXISTS` statement
- **AND** proceeds with schema migration

#### Scenario: Database already exists
- **WHEN** the configured database already exists
- **THEN** no CREATE DATABASE statement is executed
- **AND** the application proceeds directly to schema migration

### Requirement: Migration Configuration

The application SHALL support configuration options to control migration behavior.

#### Scenario: Default configuration
- **WHEN** no migration-specific configuration is provided
- **THEN** migration is enabled by default
- **AND** migration files are loaded from `db/schema/mysql/` directory
- **AND** runs in auto-approve mode

#### Scenario: Disable migration via config
- **WHEN** `database.migrate_on_startup` is set to `false`
- **THEN** the application skips migration on startup
- **AND** only creates a log message "Database migration disabled by configuration"
- **AND** proceeds to start normally

### Requirement: Migration Logging

The application SHALL provide clear logging for all migration operations.

#### Scenario: Migration start
- **WHEN** migration begins
- **THEN** the application logs "Starting database migration..."
- **AND** logs the target database connection info (host:port/database)
- **AND** logs the migration source directory path

#### Scenario: Migration changes detected
- **WHEN** Atlas detects schema changes to apply
- **THEN** the application logs "Applying N pending schema changes:"
- **AND** logs each SQL statement with context

#### Scenario: Migration completion
- **WHEN** migration completes successfully
- **THEN** the application logs "Database migration completed successfully"
- **AND** includes count of changes applied (0 if none)

### Requirement: Atlas Go SDK Integration

The application SHALL use Atlas Go SDK for schema migration instead of CLI or external processes.

#### Scenario: Using Atlas Go SDK
- **WHEN** the application loads migration functionality
- **THEN** it imports Atlas Go SDK packages directly
- **AND** does not rely on `os/exec` to call Atlas CLI
- **AND** uses Atlas's `schema` and `migrate` packages programmatically
- **AND** constructs migration plans in-memory before execution
