# build-automation Spec Delta

## ADDED Requirements

### Requirement: Task Runner SHALL use Taskfile

The system SHALL use go-task (Taskfile) as the build automation tool instead of GNU Make.

#### Scenario: Build application using Taskfile

- **WHEN** developer runs `task build`
- **THEN** the application compiles successfully
- **AND** the binary is output to `bin/modelcraft`
- **AND** the behavior is identical to the previous `make build` command

#### Scenario: Run application using Taskfile

- **WHEN** developer runs `task run`
- **THEN** the application starts with default configuration
- **AND** the server listens on the configured port
- **AND** the behavior is identical to the previous `make run` command

#### Scenario: Development mode with hot reload

- **WHEN** developer runs `task dev`
- **THEN** Air starts with hot reload enabled
- **AND** code changes trigger automatic recompilation
- **AND** the behavior is identical to the previous `make dev` command

### Requirement: Code Quality Tasks SHALL be available

The system SHALL provide code quality tasks for formatting, linting, and checking code.

#### Scenario: Format code with Taskfile

- **WHEN** developer runs `task fmt`
- **THEN** gofumpt formats all Go files
- **AND** goimports organizes imports
- **AND** the behavior is identical to the previous `make fmt` command

#### Scenario: Check code format without modification

- **WHEN** developer runs `task fmt-check`
- **THEN** the system checks if code is properly formatted
- **AND** exits with error if formatting is needed
- **AND** does not modify any files
- **AND** the behavior is identical to the previous `make fmt-check` command

#### Scenario: Run linter

- **WHEN** developer runs `task lint`
- **THEN** golangci-lint runs with project configuration
- **AND** reports any code quality issues
- **AND** the behavior is identical to the previous `make lint` command

#### Scenario: Run all checks

- **WHEN** developer runs `task check-all`
- **THEN** the system runs fmt-check, lint, vet, and test-unit sequentially
- **AND** fails if any check fails
- **AND** the behavior is identical to the previous `make check-all` command

### Requirement: Tool Installation SHALL be automated

The system SHALL provide tasks to install required development tools.

#### Scenario: Install development tools

- **WHEN** developer runs `task install-tools`
- **THEN** gofumpt, golangci-lint, and goimports are installed
- **AND** tool versions are consistent with project requirements
- **AND** the behavior is identical to the previous `make install-tools` command

#### Scenario: Install GraphQL code generator

- **WHEN** developer runs `task install-gqlgen`
- **THEN** gqlgen v0.17.83 is installed
- **AND** the tool is available in PATH
- **AND** the behavior is identical to the previous `make install-gqlgen` command

### Requirement: GraphQL Generation SHALL be supported

The system SHALL provide tasks to generate GraphQL server code.

#### Scenario: Generate GraphQL code

- **WHEN** developer runs `task generate-gql`
- **THEN** gqlgen generates code from schema files
- **AND** resolver files are created or updated
- **AND** the behavior is identical to the previous `make generate-gql` command

### Requirement: Task Help SHALL be comprehensive

The system SHALL provide comprehensive help documentation for all available tasks.

#### Scenario: Display help information

- **WHEN** developer runs `task help` or `task`
- **THEN** a categorized list of all tasks is displayed
- **AND** tasks are organized by category (Application, Docker, Testing, Database, Code Quality)
- **AND** each task shows a brief description
- **AND** the output is more comprehensive than the previous `make` target list

#### Scenario: List all tasks

- **WHEN** developer runs `task --list`
- **THEN** all available tasks are listed with descriptions
- **AND** task dependencies are visible

### Requirement: Build Targets SHALL support multiple platforms

The system SHALL provide tasks to build binaries for multiple platforms.

#### Scenario: Build for all platforms

- **WHEN** developer runs `task build-all`
- **THEN** binaries are built for Linux (amd64), Darwin (amd64, arm64), and Windows (amd64)
- **AND** each binary is output to `bin/` with platform-specific naming
- **AND** the behavior is identical to the previous `make build-all` command

### Requirement: Testing Tasks SHALL support Go unit tests

The system SHALL provide comprehensive tasks for running Go unit tests.

#### Scenario: Run unit tests

- **WHEN** developer runs `task test-unit`
- **THEN** all Go unit tests execute with race detection
- **AND** coverage profile is generated
- **AND** the behavior is identical to the previous `make test-unit` command

#### Scenario: Run unit tests with coverage report

- **WHEN** developer runs `task test-unit-coverage`
- **THEN** unit tests run with coverage tracking
- **AND** HTML coverage report is generated at `coverage.html`
- **AND** total coverage percentage is displayed
- **AND** the behavior is identical to the previous `make test-unit-coverage` command

#### Scenario: Run tests for specific package

- **WHEN** developer runs `task test-unit-pkg PKG=./internal/domain/project`
- **THEN** only tests in the specified package execute
- **AND** the behavior is identical to the previous `make test-unit-pkg PKG=./internal/domain/project` command

### Requirement: Backward Compatibility SHALL be maintained

The system SHALL maintain functional compatibility with all previous Makefile targets.

#### Scenario: All Make targets have Task equivalents

- **WHEN** comparing Makefile targets to Taskfile tasks
- **THEN** every Make target has a corresponding Task with the same name
- **AND** every Task produces identical results to its Make equivalent
- **AND** command arguments and environment variables work the same way
