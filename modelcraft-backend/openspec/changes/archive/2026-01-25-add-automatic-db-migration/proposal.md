# Change: Add Automatic Database Migration on Startup

## Why

Currently, ModelCraft requires manual database schema setup using Atlas CLI commands before starting the application. This creates a manual deployment step that increases friction and potential for errors. Developers need to remember to run migrations, and new environment setups (Docker, local dev, CI/CD) all need manual schema application.

## What Changes

- Add automatic database migration execution on application startup using Atlas Go SDK
- Automatically create the target database if it doesn't exist
- Apply schema from `db/schema/mysql/*.sql` to the configured database
- Use Atlas's `schema apply` mechanism with `--auto-approve` for non-interactive migration
- Add configuration option to control migration behavior (enable/disable)

## Impact

- **Affected specs:** None (new capability)
- **Affected code:**
  - `cmd/server/main.go` - Add migration initialization step
  - `internal/infrastructure/repository/` - Add migration package
  - `configs/config.go` - Add migration configuration
  - `go.mod` - Add Atlas Go SDK dependency

- **Breaking changes:** None

- **Benefits:**
  - Zero-configuration database setup for new deployments
  - Consistent schema state across all environments
  - Simplified Docker deployment (no separate migration step needed)
  - Reduced deployment errors from manual migration steps
