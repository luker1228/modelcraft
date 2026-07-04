package auth

// token_service_enduser.go — EndUser authentication methods on TokenService.
//
// These methods handle login/refresh/me for end-users (终端用户) whose records
// are stored in the unified users + user_orgs tables (same as tenant users).
// End-user auth flows and PAT whoami share the unified implementation here,
// while HTTP routing remains defined at the interface layer.

import (
	"context"
	"modelcraft/internal/domain/enduser"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
	"time"

	domainauth "modelcraft/internal/domain/auth"
)

// LoginEndUser authenticates an end-user by username/phone within an org scope.
func (s *TokenService) LoginEndUser(ctx context.Context, cmd LoginEndUserCommand) (*LoginEndUserResult, error) {
	if s.endUserRepoFactory == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "endUserRepoFactory not configured")
	}

	userRepo := s.endUserRepoFactory.NewEndUserRepository(s.systemDB, cmd.OrgName)
	resolvedOrgName := cmd.OrgName

	// Resolve identifier
	identifier := cmd.Identifier
	idType := cmd.IdentifierType
	if idType == "" {
		idType = IdentifierTypeUsername
	}

	var user *enduser.EndUser
	var err error

	switch idType {
	case IdentifierTypePhone:
		if identifier == "" {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, "phone is required")
		}
		// 手机号全局唯一，无需 orgName scope
		user, err = userRepo.GetByPhoneGlobal(ctx, identifier)
	default: // IdentifierTypeUsername
		if resolvedOrgName != "" {
			user, err = userRepo.GetByUsername(ctx, resolvedOrgName, identifier)
		} else {
			user, err = userRepo.GetByUsernameGlobal(ctx, identifier)
		}
	}

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

	if resolvedOrgName == "" {
		resolvedOrgName = user.OrgName
	}

	accessToken, err := s.jwtSigner.IssueAccessToken(user.ID, resolvedOrgName, user.IsAdmin)
	if err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "failed to issue access token")
	}

	plaintext, tokenHash, err := GenerateRefreshToken()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate refresh token")
	}

	sessionID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate session id")
	}

	expiresAt := time.Now().Add(s.refreshTTL)
	token := &domainauth.RefreshToken{
		ID:        sessionID,
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	if err := s.refreshTokenRepo.Save(ctx, token); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	logfacade.GetLogger(ctx).Infof(ctx, "EndUser login: id=%s, org=%s", user.ID, resolvedOrgName)

	return &LoginEndUserResult{
		UserID:       user.ID,
		OrgName:      resolvedOrgName,
		AccessToken:  accessToken,
		RefreshToken: plaintext,
		ExpiresAt:    expiresAt,
	}, nil
}

// RefreshEndUserToken rotates the end-user refresh token and issues a new access token.
// Includes reuse detection: a replayed revoked token triggers revocation of all user sessions.
func (s *TokenService) RefreshEndUserToken(
	ctx context.Context,
	cmd RefreshEndUserCommand,
) (*RefreshEndUserResult, error) {
	if s.endUserRepoFactory == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "endUserRepoFactory not configured")
	}

	tokenHash := HashToken(cmd.RefreshToken)
	token, err := s.refreshTokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if token == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}

	// Reuse detection: revoked token replayed → revoke all sessions.
	if token.IsRevoked() {
		_ = s.refreshTokenRepo.RevokeAllByUserID(ctx, token.UserID)
		if s.auditLogRepo != nil {
			_ = s.auditLogRepo.Insert(ctx, &domainauth.SecurityAuditLog{
				ID:     mustGenerateID(),
				UserID: token.UserID,
				Event:  domainauth.EventReuseDetected,
				Detail: map[string]any{"token_id": token.ID},
			})
		}
		logfacade.GetLogger(ctx).Warnf(ctx,
			"EndUser token reuse detected: user_id=%s, token_id=%s", token.UserID, token.ID,
		)
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}

	if !token.IsValid() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}

	// Fetch user to check active status and get admin flag.
	userRepo := s.endUserRepoFactory.NewEndUserRepository(s.systemDB, cmd.OrgName)
	user, err := userRepo.GetByID(ctx, cmd.OrgName, token.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserInvalidRefreshToken)
	}
	if !user.IsActive() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserAccountDisabled)
	}

	accessToken, err := s.jwtSigner.IssueAccessToken(token.UserID, cmd.OrgName, user.IsAdmin)
	if err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "failed to issue access token")
	}

	// Rotate: revoke old, issue new.
	if err := s.refreshTokenRepo.Revoke(ctx, token.ID); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	newPlaintext, newHash, err := GenerateRefreshToken()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate new refresh token")
	}

	newTokenID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.WrapError(err, bizerrors.SystemError, "generate token id")
	}

	expiresAt := time.Now().Add(s.refreshTTL)
	newToken := &domainauth.RefreshToken{
		ID:        newTokenID,
		UserID:    token.UserID,
		TokenHash: newHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	if err := s.refreshTokenRepo.Save(ctx, newToken); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	logfacade.GetLogger(ctx).Infof(ctx, "EndUser token refreshed: user_id=%s", token.UserID)

	return &RefreshEndUserResult{
		UserID:       token.UserID,
		OrgName:      cmd.OrgName,
		AccessToken:  accessToken,
		RefreshToken: newPlaintext,
		ExpiresAt:    expiresAt,
	}, nil
}

// GetEndUserMe retrieves the end-user profile for the given org+userID.
// The caller (handler) resolves identity from the Bearer JWT before calling this.
func (s *TokenService) GetEndUserMe(ctx context.Context, cmd GetEndUserMeCommand) (*enduser.EndUser, error) {
	if s.endUserRepoFactory == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.SystemError, "endUserRepoFactory not configured")
	}

	userRepo := s.endUserRepoFactory.NewEndUserRepository(s.systemDB, cmd.OrgName)
	user, err := userRepo.GetByID(ctx, cmd.OrgName, cmd.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, cmd.UserID)
	}
	return user, nil
}
