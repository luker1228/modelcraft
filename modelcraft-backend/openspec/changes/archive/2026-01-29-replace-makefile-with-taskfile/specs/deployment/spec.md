# deployment Spec Delta

## MODIFIED Requirements

### Requirement: Local Development Deployment SHALL use Taskfile

The system SHALL support deploying the local development environment using Taskfile commands.

#### Scenario: Deploy local environment

- **WHEN** developer runs `task deploy-local`
- **THEN** Docker Compose starts MySQL and Redis containers using docker-compose.local.yml
- **AND** the system waits for services to be healthy
- **AND** the application builds and starts in background mode
- **AND** health checks verify service availability
- **AND** service URLs are displayed (Application, MySQL, Redis)
- **AND** the behavior is identical to the previous `make deploy-local` command

#### Scenario: Deploy Docker Compose environment

- **WHEN** developer runs `task deploy-docker`
- **THEN** full Docker Compose stack starts (application, MySQL, Redis)
- **AND** the system waits for services to be healthy
- **AND** health checks verify service availability
- **AND** service URLs are displayed
- **AND** the behavior is identical to the previous `make deploy-docker` command

#### Scenario: Stop all deployment environments

- **WHEN** developer runs `task deploy-stop`
- **THEN** local server processes are terminated
- **AND** Docker Compose containers are stopped and removed
- **AND** both full stack and local services are cleaned up
- **AND** the behavior is identical to the previous `make deploy-stop` command

### Requirement: Docker Operations SHALL use Taskfile

The system SHALL provide Docker-related tasks using Taskfile.

#### Scenario: Build Docker image

- **WHEN** developer runs `task docker-build`
- **THEN** Docker image is built with tag `modelcraft:latest`
- **AND** the behavior is identical to the previous `make docker-build` command

#### Scenario: Start Docker Compose services

- **WHEN** developer runs `task docker-compose-up`
- **THEN** all services defined in docker-compose.yml start in detached mode
- **AND** the behavior is identical to the previous `make docker-compose-up` command

#### Scenario: Complete Docker environment startup

- **WHEN** developer runs `task docker-up`
- **THEN** Docker Compose builds images if needed
- **AND** all services start
- **AND** service URLs are displayed (Application, Health check, phpMyAdmin)
- **AND** the behavior is identical to the previous `make docker-up` command

#### Scenario: View Docker service logs

- **WHEN** developer runs `task docker-compose-logs`
- **THEN** logs from all services stream to terminal
- **AND** the behavior is identical to the previous `make docker-compose-logs` command

#### Scenario: View application container logs

- **WHEN** developer runs `task docker-app-logs`
- **THEN** only ModelCraft application logs are displayed
- **AND** logs stream in follow mode
- **AND** the behavior is identical to the previous `make docker-app-logs` command

#### Scenario: Access application container shell

- **WHEN** developer runs `task docker-shell`
- **THEN** an interactive shell opens in the ModelCraft container
- **AND** the behavior is identical to the previous `make docker-shell` command
