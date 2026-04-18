package enduser

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"modelcraft/internal/domain/enduser"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
)

const (
	// defaultRefreshTTL is the default refresh token TTL (7 days).
	defaultRefreshTTL = 7 * 24 * time.Hour
)

// PrivateDBProvider provides database connections for private_{projectSlug} databases.
// This interface allows decoupling from infrastructure layer (PrivateDBManager).
type PrivateDBProvider interface {
	// GetOrInit returns (or initializes) a database connection for private_{projectSlug}.
	// Returns error if project's cluster is not configured.
	GetOrInit(ctx context.Context, orgName, projectSlug string) (*sql.DB, error)
}

// SQLDBTX describes the database methods used by end-user repositories.
// Both *sql.DB and *sql.Tx satisfy this interface.
type SQLDBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// RepositoryFactory creates repositories from a DB or Tx connection.
// This allows the service to work with transaction-scoped repositories.
type RepositoryFactory interface {
	// NewEndUserRepository creates an EndUserRepository from a DB/TX connection.
	NewEndUserRepository(db SQLDBTX) enduser.EndUserRepository
	// NewEndUserSessionRepository creates an EndUserSessionRepository from a DB/TX connection.
	NewEndUserSessionRepository(db SQLDBTX) enduser.EndUserSessionRepository
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
	refreshTTL  time.Duration
	logger      logfacade.Logger
}

// NewEndUserAuthAppService creates a new EndUserAuthAppService.
func NewEndUserAuthAppService(
	dbProvider PrivateDBProvider,
	repoFactory RepositoryFactory,
	txManager TxManager,
	logger logfacade.Logger,
) *EndUserAuthAppService {
	return &EndUserAuthAppService{
		dbProvider:  dbProvider,
		repoFactory: repoFactory,
		txManager:   txManager,
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
	user, err := enduser.NewEndUser(userID, cmd.Username, "", hashedPwd)
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "create user entity")
	}

	// 7. Save user
	userRepo := s.repoFactory.NewEndUserRepository(db)
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
	sessionRepo := s.repoFactory.NewEndUserSessionRepository(db)
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
	// 1. Get DB connection
	db, err := s.dbProvider.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBProviderError(ctx, err)
	}

	// 2. Find user by username
	userRepo := s.repoFactory.NewEndUserRepository(db)
	user, err := userRepo.GetByUsername(ctx, cmd.Username)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 3. User not found → INVALID_CREDENTIALS (prevent enumeration)
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidCredentials)
	}

	// 4. Verify password → INVALID_CREDENTIALS
	if !user.VerifyPassword(cmd.Password) {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidCredentials)
	}

	// 5. Check if account is active
	if !user.IsActive() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserAccountDisabled)
	}

	// 6. Generate refresh token
	plaintext, tokenHash, err := generateRefreshToken()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate refresh token")
	}

	// 7. Create and save session
	sessionID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate session id")
	}
	expiresAt := time.Now().Add(s.refreshTTL)
	session := enduser.NewEndUserSession(sessionID, user.ID, tokenHash, expiresAt)

	sessionRepo := s.repoFactory.NewEndUserSessionRepository(db)
	if err := sessionRepo.Save(ctx, session); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	s.logger.Infof(ctx, "EndUser login: id=%s, username=%s", user.ID, cmd.Username)

	return &LoginResult{
		UserID:       user.ID,
		RefreshToken: plaintext,
		ExpiresAt:    expiresAt,
	}, nil
}

// LogoutEndUser handles end-user logout (idempotent).
func (s *EndUserAuthAppService) LogoutEndUser(ctx context.Context, cmd LogoutCommand) error {
	// 1. Get DB connection
	db, err := s.dbProvider.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return s.convertDBProviderError(ctx, err)
	}

	// 2. Hash token
	tokenHash := hashToken(cmd.RefreshToken)

	// 3. Find session
	sessionRepo := s.repoFactory.NewEndUserSessionRepository(db)
	session, err := sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 4. Session not found → idempotent, return nil
	if session == nil {
		return nil
	}

	// 5. Revoke session
	if err := sessionRepo.RevokeByID(ctx, session.ID); err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}

	s.logger.Infof(ctx, "EndUser logout: session_id=%s", session.ID)
	return nil
}

// RefreshEndUserToken handles token refresh with rotation (transactional).
func (s *EndUserAuthAppService) RefreshEndUserToken(ctx context.Context, cmd RefreshCommand) (*RefreshResult, error) {
	// 1. Get DB connection
	db, err := s.dbProvider.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBProviderError(ctx, err)
	}

	// 2. Hash token
	tokenHash := hashToken(cmd.RefreshToken)

	// 3. Find session
	sessionRepo := s.repoFactory.NewEndUserSessionRepository(db)
	session, err := sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	// 4. Session not found → INVALID_REFRESH_TOKEN
	if session == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}

	// 5. Session not valid (revoked or expired) → INVALID_REFRESH_TOKEN
	if !session.IsValid() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}

	// 6. Token Rotation (transactional)
	var result *RefreshResult
	err = s.txManager.WithTx(ctx, db, func(ctx context.Context, txDB SQLDBTX) error {
		txSessionRepo := s.repoFactory.NewEndUserSessionRepository(txDB)

		// 6a. Revoke old token
		if err := txSessionRepo.RevokeByID(ctx, session.ID); err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}

		// 6b. Generate new token
		newPlaintext, newTokenHash, err := generateRefreshToken()
		if err != nil {
			return bizerrors.WrapError(err, bizerrors.SystemError, "generate new refresh token")
		}

		// 6c. Create new session
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
			RefreshToken: newPlaintext,
			ExpiresAt:    expiresAt,
		}
		return nil
	})

	if err != nil {
		// If already a BusinessError, return as-is
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
	userRepo := s.repoFactory.NewEndUserRepository(db)
	user, err := userRepo.GetByID(ctx, cmd.UserID)
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

// convertDBProviderError converts PrivateDBProvider errors to BusinessErrors.
func (s *EndUserAuthAppService) convertDBProviderError(ctx context.Context, err error) *bizerrors.BusinessError {
	// Check if it's a "cluster not configured" error
	// This would be determined by the PrivateDBProvider implementation
	// For now, we check if error message contains relevant keywords
	if err != nil && (shared.IsNotFoundError(err) || containsClusterNotConfigured(err)) {
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
