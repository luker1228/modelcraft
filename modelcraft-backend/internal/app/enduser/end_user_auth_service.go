package enduser

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"modelcraft/internal/domain/enduser"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
	"strings"
	"time"
)

const (
	// defaultRefreshTTL is the default refresh token TTL (7 days).
	defaultRefreshTTL = 7 * 24 * time.Hour
)

// PrivateDBProvider provides database connections for end-user auth storage.
// This interface allows decoupling from infrastructure layer (PrivateDBManager).
type PrivateDBProvider interface {
	// Deprecated: GetOrInit still accepts projectSlug for backward compatibility,
	// but EndUser is now Org-scoped (no project binding). The projectSlug parameter
	// will be removed in a future cleanup. Callers should pass "" for projectSlug.
	GetOrInit(ctx context.Context, orgName, projectSlug string) (*sql.DB, error)
}

// SQLDBTX describes the database methods used by end-user repositories.
// Both *sql.DB and *sql.Tx satisfy this interface.
type SQLDBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// EndUserTokenIssueInput contains required data to issue an end-user access token.
type EndUserTokenIssueInput struct {
	UserID       string
	OrgName      string
	ProjectSlugs []string
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

// RepositoryFactory creates repositories from a DB or Tx connection.
// This allows the service to work with transaction-scoped repositories.
type RepositoryFactory interface {
	// NewEndUserRepository creates an EndUserRepository from a DB/TX connection.
	NewEndUserRepository(db SQLDBTX, orgName, projectSlug string) enduser.EndUserRepository
	// NewEndUserSessionRepository creates an EndUserSessionRepository from a DB/TX connection.
	NewEndUserSessionRepository(db SQLDBTX, orgName, projectSlug string) enduser.EndUserSessionRepository
	// NewEndUserProjectAccessRepository creates an EndUserProjectAccessRepository from a DB/TX connection.
	NewEndUserProjectAccessRepository(db SQLDBTX, orgName, projectSlug string) enduser.EndUserProjectAccessRepository
}

// TxManager manages database transactions for private databases.
type TxManager interface {
	// WithTx begins a transaction on the given db, passes a tx-scoped db to fn,
	// commits on success, and rolls back on error or panic.
	WithTx(ctx context.Context, db *sql.DB, fn func(ctx context.Context, txDB SQLDBTX) error) error
}

// EndUserAuthAppService handles end-user authentication use cases.
type EndUserAuthAppService struct {
	dbProvider  PrivateDBProvider
	repoFactory RepositoryFactory
	txManager   TxManager
	tokenIssuer EndUserTokenIssuer
	refreshTTL  time.Duration
	logger      logfacade.Logger
}

// NewEndUserAuthAppService creates a new EndUserAuthAppService.
func NewEndUserAuthAppService(
	dbProvider PrivateDBProvider,
	repoFactory RepositoryFactory,
	txManager TxManager,
	tokenIssuer EndUserTokenIssuer,
	logger logfacade.Logger,
) *EndUserAuthAppService {
	return &EndUserAuthAppService{
		dbProvider:  dbProvider,
		repoFactory: repoFactory,
		txManager:   txManager,
		tokenIssuer: tokenIssuer,
		refreshTTL:  defaultRefreshTTL,
		logger:      logger,
	}
}

// RegisterEndUser handles self-registration for end-users.
func (s *EndUserAuthAppService) RegisterEndUser(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	// 1. Get DB connection (includes auto-migrate)
	db, err := s.dbProvider.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBProviderError(ctx, err)
	}

	// 2. Validate password strength
	if err := enduser.ValidatePasswordStrength(cmd.Password); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, err.Error())
	}

	// 3. Validate username format
	if err := enduser.ValidateUsername(cmd.Username); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, err.Error())
	}

	// 4. Hash password (bcrypt cost=12)
	hashedPwd, err := enduser.NewHashedPasswordFromPlain(cmd.Password)
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "hash password")
	}

	// 5. Generate user ID
	userID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate user id")
	}

	// 6. Create user entity (self-registration: CreatedBy is empty)
	user, err := enduser.NewEndUser(userID, cmd.OrgName, cmd.Username, "", hashedPwd)
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "create user entity")
	}

	// 7. Save user
	userRepo := s.repoFactory.NewEndUserRepository(db, cmd.OrgName, "")
	if err := userRepo.Save(ctx, user); err != nil {
		if shared.IsDuplicateKeyError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserConflict, cmd.Username)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 8. Generate opaque refresh token
	plaintext, tokenHash, err := generateRefreshToken()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate refresh token")
	}

	// 9. Create session
	sessionID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate session id")
	}
	expiresAt := time.Now().Add(s.refreshTTL)
	session := enduser.NewEndUserSession(sessionID, user.ID, tokenHash, expiresAt)

	// 10. Save session
	sessionRepo := s.repoFactory.NewEndUserSessionRepository(db, cmd.OrgName, "")
	if err := sessionRepo.Save(ctx, session); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	s.logger.Infof(ctx, "EndUser registered: id=%s, username=%s", user.ID, cmd.Username)

	return &RegisterResult{
		UserID:       user.ID,
		RefreshToken: plaintext,
		ExpiresAt:    expiresAt,
	}, nil
}

// LoginEndUser handles end-user login.
func (s *EndUserAuthAppService) LoginEndUser(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	db, err := s.dbProvider.GetOrInit(ctx, cmd.OrgName, "")
	if err != nil {
		return nil, s.convertDBProviderError(ctx, err)
	}

	userRepo := s.repoFactory.NewEndUserRepository(db, cmd.OrgName, "")
	user, err := userRepo.GetByUsername(ctx, cmd.OrgName, cmd.Username)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidCredentials)
	}
	if !user.VerifyPassword(cmd.Password) {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidCredentials)
	}
	if !user.IsActive() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserAccountDisabled)
	}

	projectAccessRepo := s.repoFactory.NewEndUserProjectAccessRepository(db, cmd.OrgName, "")
	accessibleProjects, err := projectAccessRepo.ListAccessibleProjectsByUserID(ctx, cmd.OrgName, user.ID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	accessToken := ""
	if len(accessibleProjects) > 0 {
		projectSlugs := make([]string, 0, len(accessibleProjects))
		for _, item := range accessibleProjects {
			projectSlugs = append(projectSlugs, item.ProjectSlug)
		}

		tokenResult, issueErr := s.issueAccessToken(ctx, user.ID, cmd.OrgName, projectSlugs)
		if issueErr != nil {
			return nil, issueErr
		}
		accessToken = tokenResult.AccessToken
	}

	plaintext, tokenHash, err := generateRefreshToken()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate refresh token")
	}

	sessionID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate session id")
	}
	expiresAt := time.Now().Add(s.refreshTTL)
	session := enduser.NewEndUserSession(sessionID, user.ID, tokenHash, expiresAt)

	sessionRepo := s.repoFactory.NewEndUserSessionRepository(db, cmd.OrgName, "")
	if err := sessionRepo.Save(ctx, session); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	s.logger.Infof(
		ctx,
		"EndUser login: id=%s, username=%s, projects=%d",
		user.ID,
		cmd.Username,
		len(accessibleProjects),
	)

	return &LoginResult{
		UserID:       user.ID,
		AccessToken:  accessToken,
		Projects:     toAppAccessibleProjects(accessibleProjects),
		RefreshToken: plaintext,
		ExpiresAt:    expiresAt,
	}, nil
}

// SelectProjectContext validates that user can select the target project without reissuing token.
func (s *EndUserAuthAppService) SelectProjectContext(
	ctx context.Context,
	cmd SelectProjectCommand,
) (*SelectProjectResult, error) {
	db, err := s.dbProvider.GetOrInit(ctx, cmd.OrgName, "")
	if err != nil {
		return nil, s.convertDBProviderError(ctx, err)
	}

	sessionRepo := s.repoFactory.NewEndUserSessionRepository(db, cmd.OrgName, "")
	session, err := sessionRepo.GetByTokenHash(ctx, hashToken(cmd.RefreshToken))
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if session == nil || !session.IsValid() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}

	userRepo := s.repoFactory.NewEndUserRepository(db, cmd.OrgName, "")
	user, err := userRepo.GetByID(ctx, cmd.OrgName, session.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}
	if !user.IsActive() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserAccountDisabled)
	}

	projectAccessRepo := s.repoFactory.NewEndUserProjectAccessRepository(db, cmd.OrgName, cmd.ProjectSlug)
	hasAccess, err := projectAccessRepo.HasProjectAccess(ctx, cmd.OrgName, session.UserID, cmd.ProjectSlug)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if !hasAccess {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserProjectAccessDenied, cmd.ProjectSlug)
	}

	return &SelectProjectResult{UserID: session.UserID, ProjectSlug: cmd.ProjectSlug}, nil
}

// LogoutEndUser handles end-user logout (idempotent).
func (s *EndUserAuthAppService) LogoutEndUser(ctx context.Context, cmd LogoutCommand) error {
	db, err := s.dbProvider.GetOrInit(ctx, cmd.OrgName, "")
	if err != nil {
		return s.convertDBProviderError(ctx, err)
	}

	tokenHash := hashToken(cmd.RefreshToken)
	sessionRepo := s.repoFactory.NewEndUserSessionRepository(db, cmd.OrgName, "")
	session, err := sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}

	if session == nil {
		return nil
	}

	if err := sessionRepo.RevokeByID(ctx, session.ID); err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}

	s.logger.Infof(ctx, "EndUser logout: session_id=%s", session.ID)
	return nil
}

// RefreshEndUserToken handles token refresh with rotation (transactional).
func (s *EndUserAuthAppService) RefreshEndUserToken(ctx context.Context, cmd RefreshCommand) (*RefreshResult, error) {
	db, err := s.dbProvider.GetOrInit(ctx, cmd.OrgName, "")
	if err != nil {
		return nil, s.convertDBProviderError(ctx, err)
	}

	tokenHash := hashToken(cmd.RefreshToken)
	sessionRepo := s.repoFactory.NewEndUserSessionRepository(db, cmd.OrgName, "")
	session, err := sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if session == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}
	if !session.IsValid() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}

	userRepo := s.repoFactory.NewEndUserRepository(db, cmd.OrgName, "")
	user, err := userRepo.GetByID(ctx, cmd.OrgName, session.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}
	if !user.IsActive() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserAccountDisabled)
	}

	projectAccessRepo := s.repoFactory.NewEndUserProjectAccessRepository(db, cmd.OrgName, "")
	accessibleProjects, err := projectAccessRepo.ListAccessibleProjectsByUserID(ctx, cmd.OrgName, session.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	accessToken := ""
	if len(accessibleProjects) > 0 {
		projectSlugs := make([]string, 0, len(accessibleProjects))
		for _, item := range accessibleProjects {
			projectSlugs = append(projectSlugs, item.ProjectSlug)
		}

		tokenResult, issueErr := s.issueAccessToken(ctx, session.UserID, cmd.OrgName, projectSlugs)
		if issueErr != nil {
			return nil, issueErr
		}
		accessToken = tokenResult.AccessToken
	}

	var result *RefreshResult
	err = s.txManager.WithTx(ctx, db, func(ctx context.Context, txDB SQLDBTX) error {
		txSessionRepo := s.repoFactory.NewEndUserSessionRepository(txDB, cmd.OrgName, "")

		if err := txSessionRepo.RevokeByID(ctx, session.ID); err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}

		newPlaintext, newTokenHash, err := generateRefreshToken()
		if err != nil {
			return bizerrors.WrapError(err, bizerrors.SystemError, "generate new refresh token")
		}

		newSessionID, err := bizutils.GenerateUUIDV7()
		if err != nil {
			return bizerrors.WrapError(err, bizerrors.SystemError, "generate session id")
		}
		expiresAt := time.Now().Add(s.refreshTTL)
		newSession := enduser.NewEndUserSession(newSessionID, session.UserID, newTokenHash, expiresAt)

		if err := txSessionRepo.Save(ctx, newSession); err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}

		result = &RefreshResult{
			UserID:       session.UserID,
			AccessToken:  accessToken,
			Projects:     toAppAccessibleProjects(accessibleProjects),
			RefreshToken: newPlaintext,
			ExpiresAt:    expiresAt,
		}
		return nil
	})
	if err != nil {
		if _, ok := err.(*bizerrors.BusinessError); ok {
			return nil, err
		}
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "refresh token transaction")
	}

	s.logger.Infof(ctx, "EndUser token refreshed: user_id=%s", session.UserID)
	return result, nil
}

// GetEndUserMe retrieves the current end-user's profile.
func (s *EndUserAuthAppService) GetEndUserMe(ctx context.Context, cmd GetMeCommand) (*enduser.EndUser, error) {
	// 1. Get DB connection
	db, err := s.dbProvider.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBProviderError(ctx, err)
	}

	// 2. Find user by ID
	userRepo := s.repoFactory.NewEndUserRepository(db, cmd.OrgName, "")
	user, err := userRepo.GetByID(ctx, cmd.OrgName, cmd.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 3. User not found → NOT_FOUND.END_USER
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, cmd.UserID)
	}

	// 4. Check if account is active
	if !user.IsActive() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserAccountDisabled)
	}

	return user, nil
}

func (s *EndUserAuthAppService) issueAccessToken(
	ctx context.Context,
	userID, orgName string,
	projectSlugs []string,
) (*EndUserTokenIssueResult, error) {
	if s.tokenIssuer == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "end-user token issuer not configured")
	}

	result, err := s.tokenIssuer.IssueEndUserToken(ctx, EndUserTokenIssueInput{
		UserID:       userID,
		OrgName:      orgName,
		ProjectSlugs: projectSlugs,
	})
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "issue end-user access token")
	}
	return result, nil
}

func toAppAccessibleProjects(items []enduser.AccessibleProject) []AccessibleProject {
	projects := make([]AccessibleProject, 0, len(items))
	for _, item := range items {
		projects = append(projects, AccessibleProject{Slug: item.ProjectSlug, Title: item.ProjectTitle})
	}
	return projects
}

// convertDBProviderError converts PrivateDBProvider errors to BusinessErrors.
func (s *EndUserAuthAppService) convertDBProviderError(ctx context.Context, err error) *bizerrors.BusinessError {
	if err == nil {
		return nil
	}
	if shared.IsNotFoundError(err) {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserClusterNotConfigured)
	}
	if containsClusterNotConfigured(err) {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserClusterNotConfigured)
	}
	return bizerrors.ConvertRepositoryError(ctx, err)
}

// containsClusterNotConfigured checks if error indicates cluster is not configured.
func containsClusterNotConfigured(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "cluster not configured") ||
		strings.Contains(msg, "cluster not found") ||
		strings.Contains(msg, "project has no cluster")
}

// generateRefreshToken generates a 32-byte CSPRNG → 64-char hex string.
// Returns plaintext token and its SHA256 hash.
func generateRefreshToken() (plaintext, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	plaintext = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(plaintext))
	hash = hex.EncodeToString(sum[:])
	return plaintext, hash, nil
}

// hashToken computes SHA256 hash of a token string.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
