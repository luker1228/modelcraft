# config-env-management Specification

## Purpose
Environment-Based Configuration Management simplifies configuration by using a single `config.yaml` template combined with environment-specific `.env` files. This approach eliminates configuration duplication, improves security by keeping secrets out of git, and follows 12-Factor App principles.

## ADDED Requirements

### Requirement: Single Configuration Template File

The system SHALL maintain a single `config.yaml` file as a configuration template with no sensitive values.

#### Scenario: Configuration template has no secrets

- **WHEN** reviewing the `config.yaml` file
- **THEN** all sensitive fields (passwords, secrets, keys) SHALL be empty strings or placeholders
- **AND** comments SHALL indicate which values MUST be overridden via environment variables
- **AND** all non-sensitive defaults SHALL be present

#### Scenario: Configuration structure is documented

- **WHEN** a developer examines `config.yaml`
- **THEN** the file SHALL contain all configuration sections (server, database, redis, jwt, crypto, logger)
- **AND** each section SHALL have default values for non-sensitive fields
- **AND** comments SHALL explain which fields require environment variable overrides

### Requirement: Environment-Specific .env Files

The system SHALL support environment-specific `.env` files for local development and automated testing.

#### Scenario: Local development uses .env file

- **WHEN** starting the application with `go run cmd/server/main.go`
- **AND** a `.env` file exists in the working directory
- **THEN** the system SHALL load configuration values from `.env`
- **AND** values SHALL override defaults from `config.yaml`

#### Scenario: Automated testing uses .env.autotest file

- **WHEN** starting the application with `go run cmd/server/main.go -env .env.autotest`
- **AND** a `.env.autotest` file exists
- **THEN** the system SHALL load configuration values from `.env.autotest`
- **AND** test database name SHALL be `modelcraft_test`

### Requirement: .env Example Template Files

The repository SHALL include `.env.example` and `.env.autotest.example` template files to guide developers.

#### Scenario: .env.example provides complete template

- **WHEN** a developer copies `.env.example` to `.env`
- **THEN** the example SHALL include all required environment variables
- **AND** the example SHALL include comments explaining each variable
- **AND** placeholder values SHALL clearly indicate what needs to be filled in

#### Scenario: .env.autotest.example provides test template

- **WHEN** setting up automated testing environment
- **THEN** `.env.autotest.example` SHALL include all required test environment variables
- **AND** test database name SHALL be `modelcraft_test`
- **AND** comments SHALL explain test-specific values

### Requirement: Git Ignore Configuration

The `.gitignore` file SHALL exclude sensitive .env files while keeping example templates.

#### Scenario: Sensitive .env files are not committed

- **WHEN** running `git status`
- **THEN** `.env` SHALL be ignored
- **AND** `.env.autotest` SHALL be ignored
- **AND** `.env.example` SHALL NOT be ignored (committed to repository)
- **AND** `.env.autotest.example` SHALL NOT be ignored (committed to repository)

### Requirement: Docker Deployment Documentation

The repository SHALL include comprehensive documentation for Docker deployment environment variables.

#### Scenario: Docker README documents required variables

- **WHEN** reading Docker deployment documentation
- **THEN** all required environment variables SHALL be listed in a table
- **AND** each variable SHALL have a description and example value
- **AND** documentation SHALL explain three methods: .env file, Docker secrets, direct environment variables

#### Scenario: Docker Compose example environment variables

- **WHEN** reviewing docker-compose.yml
- **THEN** comments SHALL indicate which environment variables are required
- **AND** examples SHALL show how to use `${VARIABLE_NAME}` syntax
- **AND** a reference to detailed Docker README SHALL be present

### Requirement: Configuration File Cleanup

The repository SHALL contain only one configuration file, eliminating duplicates.

#### Scenario: Redundant config files are removed

- **WHEN** listing files in the `configs/` directory
- **THEN** only `config.yaml` SHALL exist
- **AND** `config.docker.yaml` SHALL NOT exist (deleted)
- **AND** `config.test.yaml` SHALL NOT exist (deleted)

### Requirement: Backward Compatibility

The system SHALL maintain backward compatibility with existing `-config` and `-env` command-line flags.

#### Scenario: Explicit -config flag still works

- **WHEN** starting with `go run cmd/server/main.go -config custom.yaml`
- **THEN** the system SHALL load the specified config file
- **AND** no behavioral changes SHALL occur compared to previous version

#### Scenario: Explicit -env flag works for any .env file

- **WHEN** starting with `go run cmd/server/main.go -env .env.custom`
- **THEN** the system SHALL load the specified .env file
- **AND** values SHALL override config.yaml defaults

### Requirement: Configuration Loading Priority

The system SHALL load configuration with clear precedence: environment variables > .env file > config.yaml.

#### Scenario: Environment variables have highest priority

- **WHEN** an environment variable `DB_PASSWORD` is set
- **AND** `.env` file also defines `DB_PASSWORD`
- **AND** `config.yaml` has a default `database.password`
- **THEN** the environment variable value SHALL be used

#### Scenario: .env file overrides config.yaml

- **WHEN** `.env` file defines `DB_HOST=localhost`
- **AND** `config.yaml` has `database.host: "default-host"`
- **AND** no environment variable `DB_HOST` is set
- **THEN** the `.env` value SHALL be used

### Requirement: Docker Environment Variable Support

Docker deployments SHALL use native environment variables without requiring .env files in containers.

#### Scenario: Docker Compose provides environment variables

- **WHEN** running `docker-compose up -d`
- **AND** docker-compose.yml defines environment variables
- **THEN** the container SHALL use values from docker-compose environment section
- **AND** no .env file SHALL be needed inside the container

#### Scenario: Docker secrets are supported

- **WHEN** using Docker secrets for sensitive values
- **THEN** documentation SHALL explain how to configure secrets
- **AND** docker-compose.yml SHALL show example secret references

### Requirement: Configuration Validation

The system SHALL validate required configuration values at startup and fail fast with clear error messages when critical values are missing.

#### Scenario: Missing required password fails fast

- **WHEN** starting the application
- **AND** `DB_PASSWORD` environment variable is not set
- **AND** `database.password` in config.yaml is empty
- **THEN** the system SHALL log a clear error message indicating DB_PASSWORD is required
- **AND** startup SHALL fail (not continue with empty password)

#### Scenario: Invalid AES key length fails fast

- **WHEN** starting the application
- **AND** `CRYPTO_AES_KEY` is not exactly 32 bytes
- **THEN** the system SHALL log an error about invalid key length
- **AND** the error SHALL indicate the correct length requirement (32 bytes)
- **AND** startup SHALL fail

## Acceptance Criteria

This capability is considered complete when:

1. ✅ Only `config.yaml` exists in configs/ directory (config.docker.yaml and config.test.yaml deleted)
2. ✅ config.yaml contains no sensitive values (passwords, secrets, keys are empty or placeholders)
3. ✅ `.env` file works for local development (DB connects to localhost)
4. ✅ `.env.autotest` file works for automated testing (uses modelcraft_test database)
5. ✅ `.env.example` and `.env.autotest.example` are clear and complete
6. ✅ `.gitignore` excludes `.env` and `.env.autotest`, but not example files
7. ✅ Docker deployment README section documents all required environment variables
8. ✅ docker-compose.yml includes comments showing required env vars
9. ✅ All existing Go unit tests pass without modification
10. ✅ All Python integration tests pass with `-env .env.autotest`
11. ✅ Local: `go run cmd/server/main.go` connects to localhost database
12. ✅ Testing: `go run cmd/server/main.go -env .env.autotest` connects to test database
13. ✅ Docker: `docker-compose up -d` works with environment variables
14. ✅ Backward compatibility: `-config` and `-env` flags work as before
15. ✅ Documentation (CLAUDE.md, README) updated with new approach
