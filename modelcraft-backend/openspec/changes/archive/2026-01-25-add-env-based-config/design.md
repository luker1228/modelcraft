# Design: Environment-Based Configuration with .env Files

**Change ID**: `add-env-based-config`

## Overview
This design simplifies configuration management by using a single `config.yaml` template file combined with environment-specific `.env` files. All sensitive information is moved to `.env` files (not committed to git), following 12-Factor App principles and industry best practices.

**Implementation**: Uses **godotenv** library for .env file loading, providing a clean and standard approach.

## Architecture

### Configuration Loading Strategy

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Startup                       │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
         ┌────────────────────────────┐
         │  Parse Command-Line Flags  │
         │    -config <file>          │
         │    -env <file>             │
         └────────────┬───────────────┘
                      │
                      ▼
         ┌────────────────────────────┐
         │  Load config.yaml          │
         │  (base template)           │
         │  - Default values          │
         │  - Non-sensitive config    │
         └────────────┬───────────────┘
                      │
                      ▼
         ┌────────────────────────────┐
         │  Load .env file            │
         │  (using godotenv library)  │
         │  - .env (default)          │
         │  - .env.autotest           │
         │  - or custom path          │
         └────────────┬───────────────┘
                      │
                      ▼
         ┌────────────────────────────┐
         │  Merge Environment Vars    │
         │  (from system env or       │
         │   docker-compose)          │
         └────────────┬───────────────┘
                      │
                      ▼
         ┌────────────────────────────┐
         │  Final Configuration       │
         │  Priority:                 │
         │  1. System env vars        │
         │  2. .env file values       │
         │  3. config.yaml defaults   │
         └────────────────────────────┘
```

### Key Design Decisions

#### 1. Single Configuration File Approach
**Decision**: Keep only one `config.yaml` as a template, use `.env` files for environment-specific values.

**Rationale**:
- **DRY Principle**: No duplication of configuration structure across files
- **Security**: Sensitive values never in git, only in .env files
- **Maintainability**: Changes to config structure happen in one place
- **12-Factor App**: Config in environment, not files
- **Industry Standard**: Aligns with Node.js, Ruby, Python ecosystems

**Trade-offs**:
- Requires developers to create `.env` file locally (mitigated by `.env.example`)
- Slightly different from previous multi-config approach (mitigated by documentation)

#### 2. .env File Selection via Existing Flag
**Decision**: Use existing `-env` flag to specify which `.env` file to load.

**Rationale**:
- **Zero Code Changes**: Flag already exists and works
- **Explicit Control**: Clear which environment file is being used
- **Backward Compatible**: Existing usage continues to work
- **Simple**: No new flags or environment variables needed

**Alternatives Considered**:
- `APP_ENV` environment variable: Rejected as unnecessary when `-env` flag exists
- Auto-detection: Rejected as too implicit and error-prone

#### 3. Docker Uses Native Environment Variables
**Decision**: Docker deployments inject vars via docker-compose, no `.env` file in container.

**Rationale**:
- **Container Best Practice**: Config via environment variables
- **Secrets Management**: Works with Docker secrets, Kubernetes ConfigMaps
- **Security**: No .env file to accidentally expose in images
- **Flexibility**: Easy to change config without rebuilding image

**Implementation**:
```yaml
# docker-compose.yml
services:
  modelcraft:
    environment:
      - DB_HOST=modelcraft-mysql
      - DB_PASSWORD=${MODELCRAFT_DB_PASSWORD}  # From host .env or secrets
      - JWT_SECRET=${MODELCRAFT_JWT_SECRET}
```

#### 4. File Deletion Strategy
**Decision**: Delete `config.docker.yaml` and `config.test.yaml`, keep only `config.yaml`.

**Rationale**:
- **Eliminates Duplication**: All config structure in one file
- **Forces Best Practice**: Developers must use .env for secrets
- **Cleaner Repository**: Fewer files to maintain
- **Clear Migration Path**: Old files no longer exist, forcing update

**Migration Support**:
- Document how to extract values from old config files to .env
- Provide examples for each environment
- Keep backward compatibility with `-config` flag if users need custom files

## Implementation Details

### godotenv Library Implementation

**Library Choice**: We use **godotenv v1.5.1** for loading .env files.

**Rationale**:
- **Industry Standard**: godotenv is the de facto standard for .env file loading in Go (similar to dotenv in Node.js/Python)
- **Simple API**: Clean and straightforward API for loading environment variables
- **Better Error Handling**: Clear error messages for missing or malformed .env files
- **System Integration**: Loads variables into system environment, making them available everywhere
- **Separation of Concerns**: godotenv handles .env parsing, Viper handles config merging

**Implementation in `pkg/config/config.go`**:

```go
import (
    "os"
    "github.com/joho/godotenv"
    "github.com/spf13/viper"
)

// LoadConfigWithOptions loads configuration using godotenv + Viper
func LoadConfigWithOptions(opts ConfigOptions) *Config {
    v := viper.New()

    // Load config.yaml (base template)
    if opts.ConfigFile != "" {
        loadConfigFile(v, opts.ConfigFile)
    }

    // Load .env file using godotenv
    if opts.EnvFile != "" {
        loadEnvFile(opts.EnvFile)
    }

    // Setup environment variable bindings
    setupEnvBindings(v)

    // Parse config to struct
    var config Config
    if err := v.Unmarshal(&config); err != nil {
        logfacade.GetDefault().Fatal("❌ 配置解析失败: %v", logfacade.Err(err))
    }

    return &config
}

// loadEnvFile uses godotenv to load .env file
func loadEnvFile(envFile string) {
    // Check if file exists
    if _, err := os.Stat(envFile); os.IsNotExist(err) {
        log.Printf("⚠️  环境变量文件 %s 不存在，跳过加载", envFile)
        return
    }

    // Load using godotenv
    if err := godotenv.Load(envFile); err != nil {
        log.Printf("⚠️  读取环境变量文件 %s 时出错: %v", envFile, err)
    } else {
        log.Printf("✅ 环境变量文件 %s 加载成功", envFile)
    }
}
```

**Benefits**:
1. **Cleaner Code**: Simpler than Viper's .env loading
2. **Standard Practice**: Aligns with Go ecosystem standards
3. **Better Integration**: Variables loaded into system environment before Viper reads them
4. **Explicit File Checking**: Clear error messages when .env files are missing

### Package Migration

During implementation, we also migrated configuration code from `configs/` package to `pkg/config/`:

**Files Updated**:
- `internal/infrastructure/repository/db_connection.go` - Changed import to `pkg/config`
- `internal/infrastructure/repository/cluster_connection_manager.go` - Changed import to `pkg/config`

This ensures consistency and follows the project's package structure conventions.

### Configuration File Changes

#### config.yaml (Updated)
Remove sensitive values, keep structure:
```yaml
server:
  port: "8080"
  mode: "debug"

database:
  type: "mysql"
  host: "localhost"
  port: 3306
  username: "root"
  password: ""                # MUST override via DB_PASSWORD env var
  database: "modelcraft"
  charset: "utf8mb4"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
  log_level: "info"
  migrate_on_startup: true

jwt:
  secret: ""                  # MUST override via JWT_SECRET env var

crypto:
  aes_key: ""                 # MUST override via CRYPTO_AES_KEY env var (32 bytes)
```

#### .env (Local Development)
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
```

#### .env.autotest (Automated Testing)
```bash
# Test Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=root
DB_PASSWORD=Root@SecurePass123#
DB_DATABASE=modelcraft_test

# Test Security Keys
JWT_SECRET=test-jwt-secret
CRYPTO_AES_KEY=12345678901234567890123456789012
```

### Usage Examples

#### Local Development
```bash
# Use default .env file
go run cmd/server/main.go

# Or explicitly
go run cmd/server/main.go -env .env
```

#### Automated Testing
```bash
# Use test environment file
go run cmd/server/main.go -env .env.autotest

# Or in pytest runner
pytest automated/ -v  # (runner script uses -env .env.autotest internally)
```

#### Docker Deployment
```bash
# docker-compose.yml sets environment variables
docker-compose up -d

# No .env file needed in container
# All values come from docker-compose environment section
```

## File Structure

### Files to Keep
- `config.yaml` - Single template file (committed)
- `.env.example` - Template for local development (committed)
- `.env.autotest.example` - Template for testing (committed)

### Files to Create
- `.env` - Local development values (NOT committed, in .gitignore)
- `.env.autotest` - Test environment values (NOT committed, in .gitignore)
- `.env.docker.example` - Example for docker-compose (committed)
- Docker deployment README section

### Files to Delete
- `config.docker.yaml` - No longer needed
- `config.test.yaml` - No longer needed

### .gitignore Updates
```
# Environment files (sensitive)
.env
.env.autotest
.env.local
.env.*.local

# Keep example files (templates)
!.env.example
!.env.autotest.example
!.env.docker.example
```

## Security Improvements

### Before (Current State)
- Passwords in `config.yaml` (committed to git)
- Passwords in `config.docker.yaml` (committed to git)
- Secrets scattered across multiple files
- Risk of committing production secrets

### After (New State)
- No passwords in any committed files
- All secrets in `.env` files (ignored by git)
- Clear separation: template (config.yaml) vs secrets (.env)
- Production secrets managed by docker-compose or secrets manager

## Migration Strategy

### For Existing Developers
1. Pull latest code
2. Copy `.env.example` to `.env`
3. Fill in database password and keys
4. Continue using `go run cmd/server/main.go`

### For CI/CD Pipelines
1. Create `.env.autotest` file or set environment variables
2. Update test runner: `go run cmd/server/main.go -env .env.autotest`
3. Or set env vars directly: `export DB_PASSWORD=test_pass`

### For Docker Deployments
1. Review docker-compose.yml environment section
2. Create `.env` file in project root with production values:
   ```bash
   MODELCRAFT_DB_PASSWORD=prod_password
   MODELCRAFT_JWT_SECRET=prod_secret
   MODELCRAFT_CRYPTO_KEY=prod_key_32_bytes_long_here
   ```
3. docker-compose automatically loads `.env` file
4. Or use Docker secrets for production

## Documentation Requirements

### Docker Deployment README
Create section documenting:
- Required environment variables table
- Three methods to provide values:
  1. `.env` file for docker-compose
  2. Docker secrets
  3. Direct environment variables
- Examples for each method
- Troubleshooting common issues

### Main Documentation Updates
- Update CLAUDE.md configuration section
- Remove references to multiple config files
- Document .env file approach
- Add quick start for each environment

## Testing Strategy

### Configuration File Testing
```bash
# Test local environment
go run cmd/server/main.go
curl http://localhost:8080/health

# Test autotest environment
go run cmd/server/main.go -env .env.autotest
# Should connect to modelcraft_test database

# Test Docker environment
docker-compose up -d
docker-compose exec modelcraft curl http://localhost:8080/health
```

### Validation Checks
- [ ] config.yaml parses correctly
- [ ] .env.example has all required variables
- [ ] .env.autotest.example has all required variables
- [ ] .gitignore excludes .env files
- [ ] No sensitive values in git history
- [ ] All three environments start successfully

## Benefits Summary

1. **Single Source of Truth**: One config.yaml structure
2. **Better Security**: No secrets in git, only in .env files
3. **Easier Maintenance**: Update structure once, not three times
4. **12-Factor Compliant**: Config in environment
5. **Zero Code Changes**: Existing Viper logic handles everything
6. **Backward Compatible**: -config and -env flags still work
7. **Docker-Native**: Environment variables from docker-compose
8. **Clear Templates**: .example files guide developers

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Developers forget to create .env | High | Provide .env.example, clear error messages |
| Wrong .env file used | Medium | Log which .env file is loaded, document clearly |
| Docker secrets not set | High | Document required vars, fail-fast on startup |
| Migration confusion | Medium | Provide detailed migration guide, examples |
| Lost production configs | High | Document extraction from old config files |

## Performance Impact
- **Startup Time**: No change (same Viper loading process)
- **Memory**: Slightly less (one config file vs three)
- **Runtime**: No impact (config loaded once at startup)

## Future Extensibility

### Adding New Environments
To support staging, preview, etc.:
1. Create `.env.staging` file
2. Create `.env.staging.example` template
3. Document in README
4. No code changes needed

### Configuration Validation
Future enhancement: Add startup validation
```go
func validateConfig(config *Config) error {
    if config.Database.Password == "" {
        return errors.New("DB_PASSWORD is required")
    }
    if config.JWT.Secret == "" {
        return errors.New("JWT_SECRET is required")
    }
    if len(config.Crypto.AESKey) != 32 {
        return errors.New("CRYPTO_AES_KEY must be exactly 32 bytes")
    }
    return nil
}
```

This would provide fail-fast behavior with clear error messages.
