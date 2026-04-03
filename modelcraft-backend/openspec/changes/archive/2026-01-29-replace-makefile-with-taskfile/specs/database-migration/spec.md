# database-migration Spec Delta

## MODIFIED Requirements

### Requirement: Atlas Migration Tasks SHALL use Taskfile

The system SHALL provide database migration tasks using Taskfile commands.

#### Scenario: Install Atlas CLI

- **WHEN** developer runs `task install-atlas`
- **THEN** Atlas CLI is installed using the official installation script
- **AND** the atlas binary is available in PATH
- **AND** the behavior is identical to the previous `make install-atlas` command

#### Scenario: Create new migration

- **WHEN** developer runs `task migrate-create` with migration name
- **THEN** a new migration file is created in db/migrations/
- **AND** the behavior is identical to the previous `make migrate-create` command

#### Scenario: Apply pending migrations

- **WHEN** developer runs `task migrate-up`
- **THEN** all pending migrations are applied to the database
- **AND** migration status is updated
- **AND** the behavior is identical to the previous `make migrate-up` command

#### Scenario: Rollback last migration

- **WHEN** developer runs `task migrate-down`
- **THEN** the most recent migration is rolled back
- **AND** database schema reverts to previous state
- **AND** the behavior is identical to the previous `make migrate-down` command

#### Scenario: Check migration status

- **WHEN** developer runs `task migrate-status`
- **THEN** current migration status is displayed
- **AND** pending and applied migrations are listed
- **AND** the behavior is identical to the previous `make migrate-status` command

### Requirement: Test Data Cleanup SHALL use Taskfile

The system SHALL provide test data cleanup integrated with Taskfile.

#### Scenario: Clean test database data

- **WHEN** developer runs `task clean-data` from tests directory
- **THEN** Python cleanup script executes
- **AND** test data is removed from database
- **AND** the behavior is identical to calling the previous tests/Makefile clean-data target

#### Scenario: Clean test data from root

- **WHEN** developer runs `task test:clean-data` from root directory
- **THEN** the cleanup task in tests/Taskfile.yml is invoked
- **AND** test data is removed successfully
