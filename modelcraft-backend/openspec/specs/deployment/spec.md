# deployment Specification

## Purpose
TBD - created by archiving change add-docker-deployment. Update Purpose after archive.
## Requirements
### Requirement: Bundled MySQL SHALL be deployable

The system SHALL support deploying ModelCraft with a bundled MySQL container using Docker Compose.

#### Scenario: One-command deployment with bundled MySQL

- **WHEN** user runs `docker compose up`
- **THEN** the MySQL container starts and initializes the database
- **AND** the ModelCraft application container starts after MySQL is healthy
- **AND** the application connects to the MySQL container automatically
- **AND** all required database tables are created
- **AND** the application is accessible at http://localhost:8080

#### Scenario: MySQL container health check

- **WHEN** the MySQL container starts
- **THEN** it performs a health check to verify it is ready to accept connections
- **AND** the health check passes when MySQL reports a healthy status
- **AND** the ModelCraft container waits for MySQL to be healthy before starting

#### Scenario: Database initialization from migration scripts

- **WHEN** the MySQL container starts for the first time
- **THEN** any SQL scripts in the `db/migrations/` directory are executed automatically
- **AND** the database schema is initialized from these scripts
- **AND** subsequent restarts do not re-execute the initialization scripts

#### Scenario: Data persistence across container restarts

- **WHEN** the containers are stopped and restarted
- **THEN** the MySQL data volume persists
- **AND** all previously created data is retained
- **AND** the application reconnects successfully after restart

### Requirement: External MySQL SHALL be deployable

The system SHALL support deploying ModelCraft connecting to an external MySQL instance.

#### Scenario: Connecting to external MySQL via environment variables

- **WHEN** user sets `USE_EXTERNAL_MYSQL=true`
- **AND** provides `EXTERNAL_MYSQL_HOST`, `EXTERNAL_MYSQL_PORT`, `EXTERNAL_MYSQL_USER`, `EXTERNAL_MYSQL_PASSWORD`, `EXTERNAL_MYSQL_DATABASE` environment variables
- **THEN** the ModelCraft application connects to the specified external MySQL instance
- **AND** the bundled MySQL container is not used
- **AND** the application works normally with the external database

#### Scenario: Disabling bundled MySQL container

- **WHEN** user runs `docker compose up --scale mysql=0` with external MySQL configuration
- **THEN** the MySQL container does not start
- **AND** the ModelCraft application starts and connects to external MySQL
- **AND** no errors occur related to the missing MySQL container

#### Scenario: Validation of external MySQL connectivity

- **WHEN** external MySQL connection parameters are invalid
- **THEN** the application logs an error message
- **AND** the application fails to start or retries connection according to retry logic
- **AND** the health check endpoint returns unhealthy until successful connection

### Requirement: Configuration SHALL be overridable via environment variables

The system SHALL support overriding configuration values using environment variables.

#### Scenario: Database configuration via environment variables

- **WHEN** environment variables are set for database connection (`DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USERNAME`, `DATABASE_PASSWORD`, `DATABASE_DATABASE`)
- **THEN** these values override the default configuration from the config file
- **AND** the application uses the environment variable values for database connections
- **AND** the priority is: environment variables > config file > code defaults

#### Scenario: .env file support

- **WHEN** a `.env` file exists in the project root
- **AND** Docker Compose is started with `docker compose up`
- **THEN** environment variables defined in `.env` are loaded
- **AND** these variables are available to the application containers
- **AND** they can override default configuration values

#### Scenario: Optional configuration with sensible defaults

- **WHEN** an environment variable is not set
- **THEN** the application uses the value from the config file
- **AND** if not defined in config file, uses a sensible default value
- **AND** the application starts successfully without optional variables

### Requirement: Docker Compose SHALL provide service orchestration

The system SHALL provide a complete Docker Compose configuration with proper service dependencies.

#### Scenario: Service startup order

- **WHEN** `docker compose up` is executed
- **THEN** the MySQL container starts first
- **AND** the Redis container starts (if enabled)
- **AND** the ModelCraft container waits for MySQL to be healthy
- **AND** the phpMyAdmin container starts after MySQL (if enabled)

#### Scenario: Network isolation with custom network

- **WHEN** services are started via Docker Compose
- **THEN** all services are connected to the `modelcraft-network` bridge network
- **AND** services can communicate using service names as hostnames
- **AND** MySQL is accessible to ModelCraft as hostname `mysql`

#### Scenario: Port mapping for external access

- **WHEN** Docker Compose is started
- **THEN** port 8080 on host maps to port 8080 in ModelCraft container
- **AND** port 3306 on host maps to port 3306 in MySQL container
- **AND** port 6379 on host maps to port 6379 in Redis container
- **AND** port 8081 on host maps to port 80 in phpMyAdmin container

### Requirement: Health Check SHALL be provided

The system SHALL provide health check endpoints for service monitoring.

#### Scenario: Application health endpoint

- **WHEN** the application is running and database connection is healthy
- **THEN** the `/health` endpoint returns HTTP 200 status
- **AND** the response includes health status of core components
- **AND** the endpoint is accessible from the host machine

#### Scenario: Container-level health checks

- **WHEN** containers are running
- **THEN** the health check command in Docker Compose verifies the service is responding
- **AND** failed health checks trigger container restart policies
- **AND** health check failures are logged

#### Scenario: Startup health check delay

- **WHEN** a container first starts
- **THEN** the health check waits for the `start_period` before counting failures
- **AND** the ModelCraft container has a 40-second start period
- **AND** This allows time for database initialization and application startup

