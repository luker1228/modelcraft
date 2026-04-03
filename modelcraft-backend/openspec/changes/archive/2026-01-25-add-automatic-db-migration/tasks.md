# Implementation Tasks

## 1. Configuration
- [x] 1.1 Add `MigrateOnStartup` field to `configs.DatabaseConfig` struct
- [x] 1.2 Add default value for `MigrateOnStartup` in `setDefaults()` function
- [x] 1.3 Add environment variable binding for `DB_MIGRATE_ON_STARTUP`
- [x] 1.4 Update config.yaml example to show migration option
- [x] 1.5 Update config.docker.yaml example to show migration option

## 2. Atlas Dependency
- [x] 2.1 Add Atlas Go SDK to go.mod requirements
  - `go get ariga.io/atlas` (added but approach switched to direct SQL for simplicity)
- [x] 2.2 Run `go mod tidy` to clean up dependencies

## 3. Migration Service Package
- [x] 3.1 Create `internal/infrastructure/repository/migration/` directory
- [x] 3.2 Create `migration.go` with `NewMigrationService()` constructor
- [x] 3.3 Implement `CreateDatabaseIfNotExists()` function
- [x] 3.4 Implement `LoadSchemaFiles()` function to read SQL files from db/schema/mysql/
- [x] 3.5 Implement `ApplyMigrations()` function (using direct SQL execution for simplicity and reliability)
- [x] 3.6 Add logging throughout migration operations using logfacade

## 4. Integration with Server Startup
- [x] 4.1 Add migration initialization step in `cmd/server/main.go`
  - After database connection is established
  - Before any other database operations
- [x] 4.2 Add error handling for migration failures (exit with error)
- [x] 4.3 Add console logging for migration progress (with emoji indicators)
- [x] 4.4 Ensure migration runs after default project validation but before other services

## 4.5 Docker-Entrypoint-InitDB Integration
- [x] 4.5.1 Update `docker-compose.yml` to mount `db/schema/mysql` to `/docker-entrypoint-initdb.d`
- [x] 4.5.2 Verify SQL files use `CREATE TABLE IF NOT EXISTS` for idempotency
- [x] 4.5.3 Verify SQL files use numbered prefixes for correct execution order

## 5. Testing
- [ ] 5.1 Test fresh database initialization via docker-entrypoint-initdb.d
  - Start with fresh container: `docker compose down -v && docker compose up -d`
  - Verify MySQL container runs init scripts on first startup
  - Verify tables are created correctly
  - Verify foreign keys are created
- [ ] 5.2 Test idempotent init (manual testing recommended)
  - Stop and restart containers without removing volumes
  - Verify init scripts are NOT re-executed (only runs on first DB init)
  - Verify application starts normally
- [ ] 5.3 Test migration with schema drift (manual testing recommended)
  - Add a table/column manually to existing database
  - Restart containers
  - Note: docker-entrypoint-initdb.d only runs on fresh DB, not for drift
  - For drift, use application-level migration or manual SQL
- [ ] 5.4 Test migration disable via config (for local dev without Docker)
  - Set `migrate_on_startup: false`
  - Verify application-level migration is skipped
- [x] 5.5 Update db/README.md with Docker init behavior description

## 6. Documentation
- [x] 6.1 Update CLAUDE.md with migration behavior
- [ ] 6.2 Update main.go welcome banner to indicate migration status (optional)
- [ ] 6.3 Add health endpoint check for migration status (optional enhancement)
- [x] 6.4 Add Docker deployment notes (using docker-entrypoint-initdb.d)
