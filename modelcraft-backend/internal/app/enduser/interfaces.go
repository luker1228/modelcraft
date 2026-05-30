package enduser

// interfaces.go — shared interface types for the enduser package.
// Previously defined in end_user_auth_service.go, moved here so that
// EndUserManagementAppService and other non-auth code can reference them
// without depending on the (now-deleted) auth service file.

import (
	"context"
	"database/sql"
	"time"

	domainAuth "modelcraft/internal/domain/auth"
	domainEnduser "modelcraft/internal/domain/enduser"
)

// SQLDBTX describes the database methods used by end-user repositories.
// Both *sql.DB and *sql.Tx satisfy this interface.
type SQLDBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// TxManager manages database transactions.
// Both *sql.DB and *sql.Tx satisfy SQLDBTX; fn receives a tx-scoped connection.
type TxManager interface {
	WithTx(ctx context.Context, db *sql.DB, fn func(ctx context.Context, txDB SQLDBTX) error) error
}

// RepositoryFactory creates repositories from a DB or Tx connection.
type RepositoryFactory interface {
	NewEndUserRepository(db SQLDBTX, orgName, projectSlug string) domainEnduser.EndUserRepository
	NewRefreshTokenRepository(db SQLDBTX) domainAuth.RefreshTokenRepository
}

// EndUserTokenIssueInput contains required data to issue an end-user access token.
type EndUserTokenIssueInput struct {
	UserID       string
	OrgName      string
	ProjectSlugs []string
	IsAdmin      bool
}

// EndUserTokenIssueResult is the issued end-user access token payload.
type EndUserTokenIssueResult struct {
	AccessToken string
	ExpiresAt   time.Time
}

// EndUserTokenIssuer issues end-user access tokens.
type EndUserTokenIssuer interface {
	IssueEndUserToken(ctx context.Context, input EndUserTokenIssueInput) (*EndUserTokenIssueResult, error)
}
