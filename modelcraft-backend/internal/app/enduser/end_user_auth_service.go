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
	"time"
)

const (
	// defaultRefreshTTL is the default refresh token TTL (7 days).
	defaultRefreshTTL = 7 * 24 * time.Hour
)

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
}

// TxManager manages database transactions for private databases.
type TxManager interface {
	// WithTx begins a transaction on the given db, passes a tx-scoped db to fn,
	// commits on success, and rolls back on error or panic.
	WithTx(ctx context.Context, db *sql.DB, fn func(ctx context.Context, txDB SQLDBTX) error) error
}

// EndUserAuthAppService handles end-user authentication use cases.
type EndUserAuthAppService struct {
	db          *sql.DB
	repoFactory RepositoryFactory
	txManager   TxManager
	tokenIssuer EndUserTokenIssuer
	refreshTTL  time.Duration
	logger      logfacade.Logger
}

// NewEndUserAuthAppService creates a new EndUserAuthAppService.
func NewEndUserAuthAppService(
	db *sql.DB,
	repoFactory RepositoryFactory,
	txManager TxManager,
	tokenIssuer EndUserTokenIssuer,
	logger logfacade.Logger,
) *EndUserAuthAppService {
	return &EndUserAuthAppService{
		db:          db,
		repoFactory: repoFactory,
		txManager:   txManager,
		tokenIssuer: tokenIssuer,
		refreshTTL:  defaultRefreshTTL,
		logger:      logger,
	}
}

// RegisterEndUser handles self-registration for end-users.
func (s *EndUserAuthAppService) RegisterEndUser(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	if err := enduser.ValidatePasswordStrength(cmd.Password); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, err.Error())
	}

	if err := enduser.ValidateUsername(cmd.Username); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, err.Error())
	}

	hashedPwd, err := enduser.NewHashedPasswordFromPlain(cmd.Password)
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "hash password")
	}

	userID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate user id")
	}

	user, err := enduser.NewEndUser(userID, cmd.OrgName, cmd.Username, "", hashedPwd)
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "create user entity")
	}

	userRepo := s.repoFactory.NewEndUserRepository(s.db, cmd.OrgName, "")
	if err := userRepo.Save(ctx, user); err != nil {
		if shared.IsDuplicateKeyError(err) {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserConflict, cmd.Username)
		}
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
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

	sessionRepo := s.repoFactory.NewEndUserSessionRepository(s.db, cmd.OrgName, "")
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
	userRepo := s.repoFactory.NewEndUserRepository(s.db, cmd.OrgName, "")
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

	accessibleProjects, err := userRepo.ListAccessibleProjectsByRoleAssignment(ctx, cmd.OrgName, user.ID)
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

	sessionRepo := s.repoFactory.NewEndUserSessionRepository(s.db, cmd.OrgName, "")
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

// LogoutEndUser handles end-user logout (idempotent).
func (s *EndUserAuthAppService) LogoutEndUser(ctx context.Context, cmd LogoutCommand) error {
	tokenHash := hashToken(cmd.RefreshToken)
	sessionRepo := s.repoFactory.NewEndUserSessionRepository(s.db, cmd.OrgName, "")
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
	tokenHash := hashToken(cmd.RefreshToken)
	sessionRepo := s.repoFactory.NewEndUserSessionRepository(s.db, cmd.OrgName, "")
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

	userRepo := s.repoFactory.NewEndUserRepository(s.db, cmd.OrgName, "")
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

	accessibleProjects, err := userRepo.ListAccessibleProjectsByRoleAssignment(ctx, cmd.OrgName, session.UserID)
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
	err = s.txManager.WithTx(ctx, s.db, func(ctx context.Context, txDB SQLDBTX) error {
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
	userRepo := s.repoFactory.NewEndUserRepository(s.db, cmd.OrgName, "")
	user, err := userRepo.GetByID(ctx, cmd.OrgName, cmd.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, cmd.UserID)
	}

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
