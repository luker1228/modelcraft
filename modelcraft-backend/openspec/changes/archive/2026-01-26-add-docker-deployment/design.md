# Design: Docker Deployment Support

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Compose Stack                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐         ┌──────────────┐                  │
│  │  ModelCraft  │────────▶│   MySQL 8.0  │                  │
│  │   Container  │         │   Container  │                  │
│  │   :8080      │         │    :3306     │                  │
│  └──────────────┘         └──────────────┘                  │
│         │                         │                          │
│         └─────────────────────────┴─────────┐                │
│                                           │                  │
│                                    Or External MySQL         │
│                                    (via env vars)            │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │    Redis     │  │  phpMyAdmin  │                        │
│  │   :6379      │  │    :8081     │                        │
│  └──────────────┘  └──────────────┘                        │
│       (optional)        (optional)                          │
└─────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. Multi-Container Architecture

**Decision**: Keep app and MySQL in separate containers.

**Rationale**:
- Follows Docker best practices (one process per container)
- Allows independent scaling and lifecycle management
- Easier to debug and monitor
- Standard pattern supported by orchestration tools

### 2. Environment Variable Override Pattern

**Decision**: Use environment variables for external MySQL configuration.

**Rationale**:
- Docker-native configuration mechanism
- No need to modify config files for different deployments
- Supports CI/CD pipelines seamlessly
- Well-understood pattern in the Docker ecosystem

### 3. Default Configuration Strategy

**Decision**: Create separate `config.docker.yaml` with defaults for bundled MySQL.

**Rationale**:
- Keeps original `config.yaml` for local development
- Allows different defaults per deployment scenario
- Reduces configuration confusion

### 4. Two Deployment Modes

**A. Bundled MySQL Mode (Default)**
```
docker compose up
```
- Uses MySQL container from docker-compose.yml
- App connects via service name `mysql`
- Suitable for testing, demos, and self-contained deployments

**B. External MySQL Mode**
```
USE_EXTERNAL_MYSQL=true \
EXTERNAL_MYSQL_HOST=your-mysql-host \
EXTERNAL_MYSQL_PORT=3306 \
EXTERNAL_MYSQL_USER=your-user \
EXTERNAL_MYSQL_PASSWORD=your-password \
EXTERNAL_MYSQL_DATABASE=modelcraft \
docker compose up --scale mysql=0
```
- Skips MySQL container with `--scale mysql=0`
- App connects via external settings
- Suitable for production with managed databases

## Configuration Flow

```
┌─────────────────────────────────────────────────────────┐
│                    Config Loading                        │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  1. Load config.yaml (or CONFIG_FILE path)              │
│      │                                                  │
│      ▼                                                  │
│  2. Override with environment variables if present      │
│      │                                                  │
│      ▼                                                  │
│  3. Viper hierarchical override                         │
│      • env vars like DB_HOST override config values    │
│      • Prefix: MODELCRAFT_ (or none for simplicity)    │
│                                                         │
│  Priority: Env vars > Config file > Defaults            │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Environment Variables Design

| Variable | Purpose | Default |
|----------|---------|---------|
| `USE_EXTERNAL_MYSQL` | Toggle external MySQL mode | `false` |
| `EXTERNAL_MYSQL_HOST` | External MySQL host | - |
| `EXTERNAL_MYSQL_PORT` | External MySQL port | `3306` |
| `EXTERNAL_MYSQL_USER` | External MySQL username | - |
| `EXTERNAL_MYSQL_PASSWORD` | External MySQL password | - |
| `EXTERNAL_MYSQL_DATABASE` | External MySQL database | `modelcraft` |

## Service Composition

### Core Services (always present)
- **modelcraft**: Main application (required)
- **mysql**: Bundled MySQL (optional, skip with external mode)

### Optional Services
- **redis**: Cache layer (can be disabled)
- **phpmyadmin**: MySQL admin UI (optional, for convenience)

## Health Check Strategy

1. **MySQL waits**: MySQL container configured to be ready quickly
2. **App depends on MySQL**: explicit `depends_on` with `condition: service_healthy`
3. **Retry logic**: Application-level connection retry in Go code
4. **Health check endpoint**: `/health` returns 200 when database is connected

## Migration Handling

- **Init scripts**: `docker-entrypoint-initdb.d` directory in MySQL volume
- **Existing migrations**: Copy from `db/migrations/` to init directory
- **Version management**: Consider integrated migration tool in future

## Security Considerations

### Current (Not Yet Addressed)
- Root password stored in docker-compose.yml (not ideal for production)
- No SSL/TLS encryption
- Plain text secrets in environment variables

### Recommendations (Out of Scope)
- Use Docker secrets for sensitive data
- Run with non-root user (app container already uses appuser)
- Network isolation between services
- Resource limits to prevent container overconsumption

## Trade-offs

| Aspect | Approach | Trade-off |
|--------|----------|-----------|
| Simplicity | Single docker-compose.yml | Less flexible for external-only deployments |
| Docker Native | Environment variables | Requires app code changes to support |
| Migration Strategy | Init scripts | No runtime migration management |

## Implementation Notes

1. **Viper Integration**: App already uses Viper, so environment variable support needs configuration mapping
2. **Network Naming**: Use dedicated network `modelcraft-network` for service isolation
3. **Volume Persistence**: Named volumes for MySQL and Redis data survive container recreation
