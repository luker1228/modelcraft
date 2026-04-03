# config-env-selection Specification

## Purpose
TBD - created by archiving change add-env-based-config. Update Purpose after archive.
## Requirements
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

