// Package private provides DDL migration utilities for private project databases.
// Each project has its own private database (mc_private_{projectSlug}) for end-user data isolation.
package private

import (
	"context"
	"database/sql"
	"fmt"
	"modelcraft/pkg/logfacade"
	"regexp"
)

// projectSlugRegex validates projectSlug to prevent SQL injection.
// Allowed: lowercase letters, numbers, hyphens, underscores (1-64 chars).
var projectSlugRegex = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)

// PrivateMigrator handles DDL initialization for private project databases.
// All DDL statements use IF NOT EXISTS for idempotency.
type PrivateMigrator struct {
	logger logfacade.Logger
}

// NewPrivateMigrator creates a new PrivateMigrator.
func NewPrivateMigrator(logger logfacade.Logger) *PrivateMigrator {
	return &PrivateMigrator{logger: logger}
}

// Migrate creates the private database and tables if they don't exist.
// This method is idempotent - safe to call multiple times.
//
// Parameters:
//   - ctx: context for cancellation and tracing
//   - db: database connection (must have CREATE DATABASE privilege)
//   - projectSlug: the project identifier (used in database name: mc_private_{projectSlug})
//
// Returns error if:
//   - projectSlug format is invalid (SQL injection prevention)
//   - database creation fails
//   - table creation fails
func (m *PrivateMigrator) Migrate(ctx context.Context, db *sql.DB, projectSlug string) error {
	// Validate projectSlug to prevent SQL injection
	if !projectSlugRegex.MatchString(projectSlug) {
		return fmt.Errorf("invalid projectSlug format: %s", projectSlug)
	}

	dbName := fmt.Sprintf("mc_private_%s", projectSlug)

	m.logger.Info(ctx, "starting private database migration",
		logfacade.String("database", dbName),
		logfacade.String("project_slug", projectSlug))

	// Step 1: Create database if not exists
	if err := m.createDatabase(ctx, db, dbName); err != nil {
		return fmt.Errorf("create database %s: %w", dbName, err)
	}

	// Step 2: Switch to the private database
	// NOCA:yunding/go/sql-injection(design: projectSlug validated by regex)
	if _, err := db.ExecContext(ctx, fmt.Sprintf("USE `%s`", dbName)); err != nil {
		return fmt.Errorf("use database %s: %w", dbName, err)
	}

	// Step 3: Create tables
	if err := m.createTables(ctx, db, dbName); err != nil {
		return fmt.Errorf("create tables in %s: %w", dbName, err)
	}

	m.logger.Info(ctx, "private database migration completed",
		logfacade.String("database", dbName))

	return nil
}

// createDatabase creates the private database if it doesn't exist.
func (m *PrivateMigrator) createDatabase(ctx context.Context, db *sql.DB, dbName string) error {
	// NOCA:yunding/go/sql-injection(design: dbName derived from validated projectSlug)
	createDBSQL := fmt.Sprintf(`
		CREATE DATABASE IF NOT EXISTS %s
		DEFAULT CHARACTER SET utf8mb4
		COLLATE utf8mb4_unicode_ci
	`, quoteIdentifier(dbName))

	if _, err := db.ExecContext(ctx, createDBSQL); err != nil {
		return err
	}

	m.logger.Debug(ctx, "ensured database exists", logfacade.String("database", dbName))
	return nil
}

// createTables creates the users and accounts tables if they don't exist.
func (m *PrivateMigrator) createTables(ctx context.Context, db *sql.DB, dbName string) error {
	// Create users table
	createUsersSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id           VARCHAR(36)  NOT NULL PRIMARY KEY,
			username     VARCHAR(64)  NOT NULL,
			password     VARCHAR(255) NOT NULL,
			is_forbidden TINYINT(1)   NOT NULL DEFAULT 0,
			created_by   VARCHAR(36)  NOT NULL,
			created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uq_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`

	if _, err := db.ExecContext(ctx, createUsersSQL); err != nil {
		return fmt.Errorf("create users table: %w", err)
	}
	m.logger.Debug(ctx, "ensured users table exists", logfacade.String("database", dbName))

	// Create accounts table (sessions)
	createAccountsSQL := `
		CREATE TABLE IF NOT EXISTS accounts (
			id                 VARCHAR(36)  NOT NULL PRIMARY KEY,
			user_id            VARCHAR(36)  NOT NULL,
			refresh_token_hash VARCHAR(255) NOT NULL,
			expires_at         DATETIME     NOT NULL,
			revoked            TINYINT(1)   NOT NULL DEFAULT 0,
			created_at         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_user_id (user_id),
			UNIQUE KEY uq_token_hash (refresh_token_hash),
			CONSTRAINT fk_accounts_user FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`

	if _, err := db.ExecContext(ctx, createAccountsSQL); err != nil {
		return fmt.Errorf("create accounts table: %w", err)
	}
	m.logger.Debug(ctx, "ensured accounts table exists", logfacade.String("database", dbName))

	return nil
}

// quoteIdentifier quotes a MySQL identifier (database/table name) using backticks.
// This is safe because the identifier has already been validated by projectSlugRegex.
func quoteIdentifier(name string) string {
	return "`" + name + "`"
}
