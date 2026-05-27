set shell := ["bash", "-cu"]

_default:
    @just --list

# Build and start all local deploy services
deploy:
    just --justfile deploy/justfile --working-directory deploy deploy

# Start all local deploy services without rebuilding
deploy-up:
    just --justfile deploy/justfile --working-directory deploy up

# Stop and remove local deploy containers
deploy-down:
    just --justfile deploy/justfile --working-directory deploy down

# Show local deploy service status
deploy-ps:
    just --justfile deploy/justfile --working-directory deploy ps

# Follow logs for all local deploy services
deploy-logs:
    just --justfile deploy/justfile --working-directory deploy logs

# Restart all local deploy services
deploy-restart:
    just --justfile deploy/justfile --working-directory deploy restart

# Stop containers and remove volumes. This clears MySQL data.
deploy-clean:
    just --justfile deploy/justfile --working-directory deploy clean

# Build and restart backend service
deploy-backend:
    just --justfile deploy/justfile --working-directory deploy backend

# Build and restart frontend service
deploy-frontend:
    just --justfile deploy/justfile --working-directory deploy frontend

# Build and restart agent service
deploy-agent:
    just --justfile deploy/justfile --working-directory deploy agent

# Start tools profile services
deploy-tools:
    just --justfile deploy/justfile --working-directory deploy tools

# Compatibility: start all docker services without rebuilding
docker: deploy-up

# Compatibility: start all docker services without rebuilding
docker-up: deploy-up

# Compatibility: build and start all docker services
docker-build: deploy

# Compatibility: deploy and rebuild docker services
docker-deploy: deploy

# Compatibility: stop and remove docker containers
docker-down: deploy-down

# Compatibility: show docker service status
docker-ps: deploy-ps

# Compatibility: follow docker service logs
docker-logs: deploy-logs

# Compatibility: restart docker services
docker-restart: deploy-restart

# Compatibility: stop docker services and remove volumes. This clears MySQL data.
docker-clean: deploy-clean

# Compatibility: build and restart backend service
docker-backend: deploy-backend

# Compatibility: build and restart frontend service
docker-frontend: deploy-frontend

# Compatibility: build and restart agent service
docker-agent: deploy-agent

# Compatibility: start tools profile services
docker-tools: deploy-tools
