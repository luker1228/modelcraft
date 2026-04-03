# Proposal: Environment-Based Configuration with .env Files

## Change ID
`add-env-based-config`

## Summary
Simplify configuration management by using environment-specific `.env` files instead of multiple configuration files. This approach centralizes all sensitive information in `.env` files and uses a single `config.yaml` as the base configuration template.

**Configuration Strategy**:
- **Single Config File**: Keep only `config.yaml` with default/template values
- **Local Development**: Use `.env` file for local development overrides
- **Automated Testing**: Use `.env.autotest` file for test environment
- **Docker/Production**: Inject sensitive values via docker-compose environment variables

## Problem Statement
Currently, the application has multiple configuration files (`config.yaml`, `config.docker.yaml`, `config.test.yaml`) which creates several issues:

1. **Duplication**: Same configuration structure repeated across multiple files
2. **Maintenance burden**: Changes to config structure require updating all files
3. **Security risk**: Sensitive values scattered across multiple config files
4. **Manual selection**: Requires explicit `-config` flag or manual file management
5. **Not following 12-Factor App**: Config files instead of environment variables for secrets

## Goals
1. **Single Source of Truth**: One `config.yaml` with template/default values
2. **Environment-Based .env Files**: Use `.env` and `.env.autotest` for different contexts
3. **Security Best Practice**: All sensitive data in .env files (not committed to git)
4. **Docker-Friendly**: Production uses environment variables from docker-compose
5. **Zero Breaking Changes**: Maintain backward compatibility with existing `-config` and `-env` flags

## Non-Goals
- Creating new configuration parameters
- Changing database migration or initialization logic
- Adding configuration validation beyond what Viper provides
- Modifying docker-compose file structure

## Proposed Solution

### Configuration Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  config.yaml (base template, committed to git)              │
│  - Default values                                           │
│  - Non-sensitive configuration                              │
│  - Structure documentation                                  │
└─────────────────────────────────────────────────────────────┘
                           ↓ Merged with
┌─────────────────────────────────────────────────────────────┐
│  .env files (environment-specific, NOT committed)           │
│  ├─ .env (local development)                                │
│  │  - DB_PASSWORD=local_password                            │
│  │  - DB_HOST=localhost                                     │
│  └─ .env.autotest (automated testing)                       │
│     - DB_PASSWORD=test_password                             │
│     - DB_DATABASE=modelcraft_test                           │
└─────────────────────────────────────────────────────────────┘
                           ↓ Or
┌─────────────────────────────────────────────────────────────┐
│  Docker Compose environment (production)                    │
│  - DB_PASSWORD=prod_password (from secrets)                 │
│  - DB_HOST=modelcraft-mysql                                 │
└─────────────────────────────────────────────────────────────┘
```

### Environment File Selection

Use `-env` flag (already exists) to specify which `.env` file to load:

**Local Development**:
```bash
go run cmd/server/main.go -env .env
# Or simply (default):
go run cmd/server/main.go
```

**Automated Testing**:
```bash
go run cmd/server/main.go -env .env.autotest
```

**Docker/Production**:
```bash
# No .env file needed - use docker-compose environment variables
docker-compose up
```

### File Structure

**Keep**:
- `config.yaml` - Single configuration file with defaults
- `.env` - Local development environment variables (add to .gitignore)
- `.env.example` - Template showing all available variables (committed to git)

**Create**:
- `.env.autotest` - Automated testing environment variables (add to .gitignore)
- `.env.autotest.example` - Template for test environment (committed to git)

**Delete**:
- `config.docker.yaml` - Replaced by docker-compose env vars
- `config.test.yaml` - Replaced by .env.autotest

## Impact Analysis

### Benefits
- **Reduced Duplication**: One config file instead of three, easier to maintain
- **Better Security**: All secrets in .env files (not committed), following 12-Factor App principles
- **Simpler Mental Model**: "One config + environment overrides" is easier to understand
- **Docker-Native**: Production environments use native docker-compose environment variables
- **No Code Changes**: Existing `-config` and `-env` flags continue to work
- **Cleaner Git History**: No sensitive values in config.yaml, only template structure

### Risks & Mitigation
| Risk | Mitigation |
|------|-----------|
| Existing deployments break | Keep `-config` and `-env` flags working; document migration path |
| Developers missing .env file | Provide `.env.example` and clear error messages |
| Docker secrets not configured | Document docker-compose env var requirements in README |
| Test environment misconfiguration | Provide `.env.autotest.example` template |

## Affected Components
- `configs/config.yaml` - Update to remove sensitive values, keep as template
- `configs/config.docker.yaml` - DELETE this file
- `configs/config.test.yaml` - DELETE this file
- `.env` - Update with all local development overrides
- `.env.autotest` - CREATE for automated testing
- `.env.example` - Update to show all available variables
- `.env.autotest.example` - CREATE as template
- `.gitignore` - Ensure .env and .env.autotest are ignored
- `docker-compose.yml` - Document required environment variables
- `README.md` or deployment docs - Add configuration guide
- CI/CD scripts - Update to use `-env .env.autotest`

## Implementation Details

### config.yaml Changes
Remove all sensitive values, keep structure with placeholders:

```yaml
server:
  port: "8080"
  mode: "debug"

database:
  type: "mysql"
  host: "localhost"          # Override in .env or docker-compose
  port: 3306
  username: "root"            # Override in .env or docker-compose
  password: ""                # MUST override in .env or docker-compose
  database: "modelcraft"      # Override in .env.autotest for tests
  charset: "utf8mb4"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
  log_level: "info"
  migrate_on_startup: true

redis:
  host: "localhost"           # Override in docker-compose
  port: 6379
  password: ""                # Override if needed

jwt:
  secret: ""                  # MUST override in .env or docker-compose

crypto:
  aes_key: ""                 # MUST override in .env or docker-compose (32 bytes)

logger:
  level: "info"
  output_path: "stdout"
  max_size: 100
  max_backups: 10
  max_age: 7
  compress: true
```

### .env File (Local Development)
```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=your-local-password
DB_DATABASE=modelcraft

# Security Keys
JWT_SECRET=your-local-jwt-secret
CRYPTO_AES_KEY=12345678901234567890123456789012

# Optional Redis Password
REDIS_PASSWORD=
```

### .env.autotest File (Automated Testing)
```bash
# Test Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=Root@SecurePass123#
DB_DATABASE=modelcraft_test     # Use separate test database

# Test Security Keys (can be dummy values)
JWT_SECRET=test-jwt-secret
CRYPTO_AES_KEY=12345678901234567890123456789012
```

### docker-compose.yml Environment Variables
```yaml
services:
  modelcraft:
    image: modelcraft:latest
    environment:
      # Server Configuration
      - GIN_MODE=release
      - PORT=8080

      # Database Configuration
      - DB_HOST=modelcraft-mysql
      - DB_PORT=3306
      - DB_USERNAME=modelcraft
      - DB_PASSWORD=${MODELCRAFT_DB_PASSWORD}  # From .env or secrets
      - DB_DATABASE=modelcraft

      # Security Keys (from secrets management)
      - JWT_SECRET=${MODELCRAFT_JWT_SECRET}
      - CRYPTO_AES_KEY=${MODELCRAFT_CRYPTO_KEY}

      # Redis Configuration
      - REDIS_HOST=modelcraft-redis
      - REDIS_PORT=6379
```

### Docker Deployment README Section
```markdown
## Docker Deployment Configuration

### Required Environment Variables

The application requires the following environment variables in production:

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_PASSWORD` | Database password | `secure_password_123` |
| `JWT_SECRET` | JWT signing secret | `random_secret_key` |
| `CRYPTO_AES_KEY` | AES encryption key (32 bytes) | `abf8fb41D97b0Cf7C43c9dbDac08A4B6` |

### Setting Environment Variables

**Option 1: .env file for docker-compose** (Development/Staging)
```bash
# Create .env file in project root
cat > .env << EOF
MODELCRAFT_DB_PASSWORD=your_db_password
MODELCRAFT_JWT_SECRET=your_jwt_secret
MODELCRAFT_CRYPTO_KEY=your_32_byte_aes_key
EOF

docker-compose up -d
```

**Option 2: Docker Secrets** (Production)
```bash
# Create secrets
echo "prod_password" | docker secret create db_password -
echo "prod_jwt_secret" | docker secret create jwt_secret -

# Reference in docker-compose.yml
secrets:
  - db_password
  - jwt_secret
```

**Option 3: Environment Variables** (CI/CD)
```bash
export DB_PASSWORD=secure_password
export JWT_SECRET=secret_key
export CRYPTO_AES_KEY=32_byte_key
docker-compose up -d
```
```

## Alternatives Considered

### Alternative 1: Keep Multiple Config Files with APP_ENV
Use `APP_ENV` variable to select between config files.
- **Rejected**: Still has duplication problem; doesn't solve security issues

### Alternative 2: Config File Inheritance
Base config + environment-specific overlays.
- **Rejected**: Over-engineering; Viper already merges env vars naturally

### Alternative 3: All Configuration via Environment Variables
No config.yaml at all, everything from env vars.
- **Rejected**: Too verbose; loses structure documentation; harder to understand defaults

## Success Criteria
1. ✅ Single `config.yaml` file with no sensitive values
2. ✅ `.env` file works for local development
3. ✅ `.env.autotest` file works for automated testing
4. ✅ Docker deployment works with docker-compose environment variables
5. ✅ `.env.example` and `.env.autotest.example` templates are clear and complete
6. ✅ All existing tests pass without modification to test code
7. ✅ Docker deployment README section documents all required env vars
8. ✅ `.gitignore` properly excludes `.env` and `.env.autotest`
9. ✅ Existing `-config` and `-env` flags continue to work (backward compatibility)
10. ✅ Clear error messages when required env vars are missing

## Related Specifications
This change modifies the existing configuration management capability.

**Modified Capability**: `config-management`
- Requirements for .env file-based configuration
- Requirements for docker-compose environment variable documentation
- Requirements for configuration template and example files

## Migration Path

### For Developers
1. Pull latest code with updated `config.yaml`
2. Copy `.env.example` to `.env`
3. Fill in sensitive values in `.env` for local database
4. No changes to workflow: `go run cmd/server/main.go` continues to work

### For CI/CD
1. Update test scripts to use `-env .env.autotest` flag
2. Create `.env.autotest` file with test database credentials
3. Or set environment variables directly in CI pipeline

### For Docker Deployment
1. Review docker-compose.yml environment variables
2. Create production `.env` file or use secrets management
3. Ensure all required env vars (DB_PASSWORD, JWT_SECRET, CRYPTO_AES_KEY) are set
4. Test deployment: `docker-compose up -d`

## Timeline
- **Proposal Review**: 1 day
- **Implementation**: 1-2 days
  - Update config.yaml (remove sensitive values)
  - Create .env templates
  - Update .gitignore
  - Write Docker deployment README section
- **Testing**: 1 day (verify all three environments)
- **Documentation**: 1 day (update CLAUDE.md and README)
- **Total**: 3-4 days

## Questions for Stakeholders
1. Should we support additional .env files for other environments (e.g., `.env.staging`)?
2. Do we need configuration validation at startup to fail fast if required env vars are missing?
3. Should we provide a script to help migrate from old config files to new .env format?
4. Do we want to support both docker secrets AND environment variables, or just environment variables?
