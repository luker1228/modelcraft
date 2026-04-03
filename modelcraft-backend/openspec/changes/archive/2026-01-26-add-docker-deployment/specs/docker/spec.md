# docker Specification

## Purpose
Defines Docker-specific capabilities for building and running ModelCraft containers.

## Requirements

## ADDED Requirements

### Requirement: Dockerfile SHALL use multi-stage build

The system SHALL provide a production-ready Dockerfile using multi-stage builds.

#### Scenario: Multi-stage build with Go builder stage

- **WHEN** building the Docker image
- **THEN** a multi-stage build is used with a separate builder stage
- **AND** the builder stage uses `golang:1.25-alpine` as base image
- **AND** the builder stage downloads dependencies before copying source code
- **AND** the final stage only contains the compiled binary and runtime dependencies
- **AND** the final image size is minimized

#### Scenario: Non-root user execution

- **WHEN** the Docker container runs
- **THEN** it uses a non-root user named `appuser` with UID 1001
- **AND** the user belongs to a group `appgroup` with GID 1001
- **AND** the application directory is owned by `appuser:appgroup`
- **AND** this reduces the security surface of running the application

#### Scenario: Alpine Linux runtime base

- **WHEN** the Dockerfile final stage is built
- **THEN** it uses `alpine:latest` as the base image
- **AND** only essential runtime packages are installed
- **AND** the image size is kept minimal

### Requirement: Docker SHALL provide configuration files

The system SHALL provide Docker-specific configuration files.

#### Scenario: Docker config file with bundled MySQL settings

- **WHEN** `configs/config.docker.yaml` exists
- **THEN** it references the MySQL service hostname as `mysql`
- **AND** it uses the bundled MySQL credentials configured in docker-compose
- **AND** port is set to 3306
- **AND** database name is `modelcraft`

#### Scenario: Environment variable example file

- **WHEN** `.env.example` exists in the project root
- **THEN** it contains all configurable environment variables
- **AND** each variable has a comment explaining its purpose
- **AND** default values are provided where appropriate
- **AND** the file can be copied to `.env` and customized

### Requirement: Docker SHALL provide optional service containers

The system SHALL provide optional service containers for convenience.

#### Scenario: Redis cache container

- **WHEN** `docker compose up` is executed
- **AND** the redis service is enabled in docker-compose.yml
- **THEN** a Redis 7 Alpine container starts
- **AND** it is accessible at hostname `redis` within the Docker network
- **AND** host port 6379 maps to container port 6379
- **AND** Redis data persists in a named volume

#### Scenario: phpMyAdmin database management UI

- **WHEN** `docker compose up` is executed
- **AND** the phpmyadmin service is enabled in docker-compose.yml
- **THEN** a phpMyAdmin container starts
- **AND** it is accessible at http://localhost:8081
- **AND** it connects to the MySQL container automatically
- **AND** users can manage the database through a web interface

#### Scenario: Disabling optional services

- **WHEN** optional services are not needed
- **THEN** they can be commented out in docker-compose.yml
- **AND** the ModelCraft application continues to work without them
- **AND** the services are not started when `docker compose up` is executed

### Requirement: Docker SHALL manage volumes for data persistence

The system SHALL use Docker volumes for data persistence.

#### Scenario: Named volume for MySQL data

- **WHEN** the MySQL container starts for the first time
- **THEN** a named volume `mysql_data` is created
- **AND** all MySQL data is stored in this volume
- **AND** data persists across container restarts and updates
- **AND** the volume can be backed up using standard Docker volume commands

#### Scenario: Bind mount for configuration files

- **WHEN** volumes are mounted in docker-compose.yml
- **THEN** the `./configs` directory is mounted to `/app/configs` in the container
- **AND** users can modify config files on the host
- **AND** changes take effect when containers are restarted
- **AND** multiple environments can use different config files

#### Scenario: Bind mount for application logs

- **WHEN** the application writes logs
- **THEN** logs are written to `/app/logs` inside the container
- **AND** this directory is mounted from `./logs` on the host
- **AND** logs can be viewed and rotated on the host without entering the container

### Requirement: Docker SHALL configure restart policies

The system SHALL configure appropriate restart policies for containers.

#### Scenario: Automatic restart on failure

- **WHEN** a container crashes unexpectedly
- **THEN** the restart policy `unless-stopped` takes effect
- **AND** the container is automatically restarted
- **AND** this happens up to the Docker daemon's restart limit
- **AND** crashes are logged for troubleshooting

#### Scenario: Manual stop prevents restart

- **WHEN** a container is stopped manually with `docker compose stop`
- **THEN** the restart policy does not restart the container
- **AND** the container remains stopped until explicitly started
- **AND** this allows for controlled maintenance and upgrades

### Requirement: Docker SHALL optimize build process

The system SHALL optimize the Docker build process.

#### Scenario: Layer caching with dependency download

- **WHEN** the Docker image is built
- **THEN** go.mod and go.sum are copied first
- **THEN** `go mod download` is executed
- **AND** this layer is cached when dependencies haven't changed
- **AND** subsequent builds are faster when only source code changes

#### Scenario: Build argument for Go build flags

- **WHEN** the binary is compiled in the builder stage
- **THEN** `ldflags="-w -s"` is used to reduce binary size
- **AND** `-w` removes DWARF debug information
- **AND** `-s` removes symbol table
- **AND** the resulting binary is significantly smaller

#### Scenario: Timezone configuration

- **WHEN** the Docker container is built
- **THEN** the `APK` cache for Alpine packages includes `tzdata`
- **AND** timezone is set to `Asia/Shanghai`
- **AND** the timezone can be overridden by setting the `TZ` environment variable
