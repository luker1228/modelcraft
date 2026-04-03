# Proposal: Add Docker Deployment Support

## Why

ModelCraft currently has Dockerfile and docker-compose.yml files, but they are not production-ready for easy deployment:

1. **No default MySQL configuration**: The existing docker-compose.yml includes MySQL but the config file points to a fixed external IP (9.135.32.8) rather than using the bundled MySQL service.

2. **No external MySQL support**: When users want to use their own MySQL database, there's no straightforward way to connect via environment variables.

3. **No one-click deployment experience**: Users need to manually configure multiple files before they can deploy with docker-compose.

This proposal aims to provide a seamless Docker deployment experience where users can:
- Run everything with a single `docker compose up` command
- Use the bundledMySQL by default
- Connect to their external MySQL via simple environment variable overrides

## What Changes

### 1. Enhanced Docker Compose Configuration

- Update `docker-compose.yml` to use bundled MySQL as default
- Add environment variable support for external MySQL override
- Keep MySQL and app in separate containers (multi-architecture)
- Ensure proper service dependencies and health checks

### 2. Docker-Ready Configuration

- Create `configs/config.docker.yaml` with sensible defaults for bundled MySQL
- Support environment variable overrides in the application config loading
- Provide `.env.example` with all configurable values documented

### 3. External MySQL Support

- Allow connecting to external MySQL via `USE_EXTERNAL_MYSQL` environment variable
- When `USE_EXTERNAL_MYSQL=true`, skip the mysql service in compose or allow override via:
  - `EXTERNAL_MYSQL_HOST`
  - `EXTERNAL_MYSQL_PORT`
  - `EXTERNAL_MYSQL_USER`
  - `EXTERNAL_MYSQL_PASSWORD`
  - `EXTERNAL_MYSQL_DATABASE`

### 4. Documentation

- Add `DEPLOYMENT.md` with quick start guide
- Document both bundled and external MySQL scenarios
- Include troubleshooting tips

## Impact

- **Affected specs**: `deployment` (new spec), `docker` (new spec)
- **Affected code**:
  - `docker-compose.yml` - Update with environment variable support
  - `configs/config.docker.yaml` - Add Docker-specific config
  - `.env.example` - Document all environment variables
  - `cmd/server/main.go` - Enhance config loading to support env var overrides
  - `configs/config.yaml` - Update default to point to bundled MySQL host

## Out of Scope

- Kubernetes deployment (separate proposal)
- Helm charts (separate proposal)
- Production-grade security hardening (SSL/TLS, secret management)
